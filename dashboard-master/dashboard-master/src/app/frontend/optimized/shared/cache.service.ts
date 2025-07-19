// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import {Injectable} from '@angular/core';
import {OnListChangeEvent} from '@api/root.ui';
import {BehaviorSubject, Observable} from 'rxjs';
import {map, shareReplay} from 'rxjs/operators';

interface CacheEntry<T = any> {
  data: T;
  timestamp: number;
  accessCount: number;
  lastAccess: number;
  size: number;
}

interface CacheStats {
  hitCount: number;
  missCount: number;
  totalSize: number;
  entryCount: number;
  evictionCount: number;
}

@Injectable({
  providedIn: 'root'
})
export class ResourceCacheService {
  private readonly cache = new Map<string, CacheEntry>();
  private readonly listEventCache = new Map<string, CacheEntry<OnListChangeEvent>>();
  private readonly computedCache = new Map<string, CacheEntry>();
  
  // Cache configuration
  private readonly maxCacheSize = 50 * 1024 * 1024; // 50MB max cache size
  private readonly maxEntries = 1000;
  private readonly defaultTTL = 5 * 60 * 1000; // 5 minutes TTL
  private readonly maxListEventEntries = 100;
  
  // Performance stats
  private stats: CacheStats = {
    hitCount: 0,
    missCount: 0,
    totalSize: 0,
    entryCount: 0,
    evictionCount: 0
  };

  private readonly stats$ = new BehaviorSubject<CacheStats>(this.stats);

  // Public observables for monitoring
  readonly cacheStats$ = this.stats$.asObservable().pipe(shareReplay(1));
  readonly hitRatio$ = this.cacheStats$.pipe(
    map(stats => stats.hitCount + stats.missCount > 0 ? stats.hitCount / (stats.hitCount + stats.missCount) : 0)
  );

  constructor() {
    // Schedule periodic cleanup
    setInterval(() => this.cleanupExpiredEntries(), 60000); // Every minute
  }

  /**
   * Get cached data with type safety
   */
  get<T>(key: string): T | null {
    const entry = this.cache.get(key);
    
    if (!entry) {
      this.recordMiss();
      return null;
    }

    if (this.isExpired(entry)) {
      this.cache.delete(key);
      this.recordMiss();
      return null;
    }

    // Update access statistics
    entry.accessCount++;
    entry.lastAccess = Date.now();
    this.recordHit();
    
    return entry.data as T;
  }

  /**
   * Set cached data with automatic size calculation
   */
  set<T>(key: string, data: T, ttl: number = this.defaultTTL): void {
    const size = this.calculateSize(data);
    
    // Check if we need to make space
    if (this.shouldEvict(size)) {
      this.evictLRU();
    }

    const entry: CacheEntry<T> = {
      data,
      timestamp: Date.now(),
      accessCount: 1,
      lastAccess: Date.now(),
      size
    };

    this.cache.set(key, entry);
    this.updateStats();
  }

  /**
   * Cache list change events with LRU eviction
   */
  cacheListEvent(event: OnListChangeEvent): void {
    const key = `${event.groupId}:${event.id}`;
    
    if (this.listEventCache.size >= this.maxListEventEntries) {
      // Evict oldest entry
      const oldestKey = this.listEventCache.keys().next().value;
      this.listEventCache.delete(oldestKey);
      this.stats.evictionCount++;
    }

    const entry: CacheEntry<OnListChangeEvent> = {
      data: event,
      timestamp: Date.now(),
      accessCount: 1,
      lastAccess: Date.now(),
      size: this.calculateSize(event)
    };

    this.listEventCache.set(key, entry);
  }

  /**
   * Get cached list event
   */
  getCachedListEvent(groupId: string, id: string): OnListChangeEvent | null {
    const key = `${groupId}:${id}`;
    const entry = this.listEventCache.get(key);
    
    if (!entry) {
      return null;
    }

    entry.accessCount++;
    entry.lastAccess = Date.now();
    return entry.data;
  }

  /**
   * Cache computed values with memoization
   */
  memoize<T>(key: string, computeFn: () => T, ttl: number = this.defaultTTL): T {
    const cached = this.get<T>(key);
    
    if (cached !== null) {
      return cached;
    }

    const computed = computeFn();
    this.set(key, computed, ttl);
    return computed;
  }

  /**
   * Invalidate cache entries by pattern
   */
  invalidatePattern(pattern: RegExp): void {
    const keysToDelete: string[] = [];
    
    for (const key of this.cache.keys()) {
      if (pattern.test(key)) {
        keysToDelete.push(key);
      }
    }

    keysToDelete.forEach(key => {
      const entry = this.cache.get(key);
      if (entry) {
        this.stats.totalSize -= entry.size;
        this.stats.entryCount--;
      }
      this.cache.delete(key);
    });

    this.updateStats();
  }

  /**
   * Clear all caches
   */
  clear(): void {
    this.cache.clear();
    this.listEventCache.clear();
    this.computedCache.clear();
    
    this.stats = {
      hitCount: 0,
      missCount: 0,
      totalSize: 0,
      entryCount: 0,
      evictionCount: 0
    };
    
    this.updateStats();
  }

  /**
   * Cleanup expired and least recently used entries
   */
  cleanup(): void {
    this.cleanupExpiredEntries();
    this.enforceMemoryLimits();
  }

  /**
   * Get cache statistics
   */
  getStats(): CacheStats {
    return {...this.stats};
  }

  /**
   * Get cache size in bytes
   */
  getCacheSize(): number {
    return this.stats.totalSize;
  }

  /**
   * Check if cache entry is expired
   */
  private isExpired(entry: CacheEntry): boolean {
    return Date.now() - entry.timestamp > this.defaultTTL;
  }

  /**
   * Calculate approximate size of data in bytes
   */
  private calculateSize(data: any): number {
    try {
      return new Blob([JSON.stringify(data)]).size;
    } catch {
      // Fallback for circular references or non-serializable data
      return JSON.stringify(data).length * 2; // Rough estimate for UTF-16
    }
  }

  /**
   * Check if we should evict entries to make space
   */
  private shouldEvict(newEntrySize: number): boolean {
    return (
      this.stats.totalSize + newEntrySize > this.maxCacheSize ||
      this.stats.entryCount >= this.maxEntries
    );
  }

  /**
   * Evict least recently used entries
   */
  private evictLRU(): void {
    const entries = Array.from(this.cache.entries())
      .map(([key, entry]) => ({key, ...entry}))
      .sort((a, b) => a.lastAccess - b.lastAccess);

    // Evict 20% of entries or at least 1
    const evictCount = Math.max(1, Math.floor(entries.length * 0.2));
    
    for (let i = 0; i < evictCount && i < entries.length; i++) {
      const entryToEvict = entries[i];
      this.cache.delete(entryToEvict.key);
      this.stats.totalSize -= entryToEvict.size;
      this.stats.entryCount--;
      this.stats.evictionCount++;
    }
  }

  /**
   * Clean up expired entries
   */
  private cleanupExpiredEntries(): void {
    const now = Date.now();
    const expiredKeys: string[] = [];

    for (const [key, entry] of this.cache.entries()) {
      if (now - entry.timestamp > this.defaultTTL) {
        expiredKeys.push(key);
      }
    }

    expiredKeys.forEach(key => {
      const entry = this.cache.get(key);
      if (entry) {
        this.stats.totalSize -= entry.size;
        this.stats.entryCount--;
      }
      this.cache.delete(key);
    });

    // Also cleanup list event cache
    for (const [key, entry] of this.listEventCache.entries()) {
      if (now - entry.timestamp > this.defaultTTL) {
        this.listEventCache.delete(key);
      }
    }

    this.updateStats();
  }

  /**
   * Enforce memory limits
   */
  private enforceMemoryLimits(): void {
    while (this.stats.totalSize > this.maxCacheSize || this.stats.entryCount > this.maxEntries) {
      this.evictLRU();
    }
  }

  /**
   * Record cache hit
   */
  private recordHit(): void {
    this.stats.hitCount++;
  }

  /**
   * Record cache miss
   */
  private recordMiss(): void {
    this.stats.missCount++;
  }

  /**
   * Update statistics
   */
  private updateStats(): void {
    this.stats.totalSize = Array.from(this.cache.values())
      .reduce((total, entry) => total + entry.size, 0);
    this.stats.entryCount = this.cache.size;
    this.stats$.next({...this.stats});
  }
}