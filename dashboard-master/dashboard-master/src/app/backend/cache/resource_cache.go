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

package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

)

// CacheEntry represents a cached resource with metadata
type CacheEntry struct {
	Data      interface{}
	CachedAt  time.Time
	ExpiresAt time.Time
	Version   string
}

// IsExpired checks if the cache entry has expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// ResourceCache provides intelligent caching for Kubernetes resources
type ResourceCache struct {
	// Core cache storage
	cache    map[string]*CacheEntry
	cacheMu  sync.RWMutex
	
	// Configuration
	defaultTTL    time.Duration
	maxCacheSize  int
	
	// Informers for real-time updates
	informerFactory informers.SharedInformerFactory
	informers       map[string]cache.SharedIndexInformer
	
	// Lifecycle management
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *slog.Logger
	
	// Statistics
	stats    CacheStats
	statsMu  sync.RWMutex
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	Hits           int64
	Misses         int64
	Evictions      int64
	TotalRequests  int64
	LastEviction   time.Time
	CacheSize      int
}

// NewResourceCache creates a new resource cache with informers
func NewResourceCache(ctx context.Context, client kubernetes.Interface, logger *slog.Logger) *ResourceCache {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = slog.Default()
	}
	
	childCtx, cancel := context.WithCancel(ctx)
	
	cache := &ResourceCache{
		cache:        make(map[string]*CacheEntry),
		defaultTTL:   5 * time.Minute,
		maxCacheSize: 10000,
		informers:    make(map[string]cache.SharedIndexInformer),
		ctx:          childCtx,
		cancel:       cancel,
		logger:       logger,
	}
	
	// Create informer factory with 30 second resync period
	cache.informerFactory = informers.NewSharedInformerFactory(client, 30*time.Second)
	
	// Setup informers for critical resources
	cache.setupInformers()
	
	// Start cleanup routine
	go cache.cleanupRoutine()
	
	return cache
}

// setupInformers configures informers for frequently accessed resources
func (rc *ResourceCache) setupInformers() {
	// Pod informer
	podInformer := rc.informerFactory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    rc.onPodAdd,
		UpdateFunc: rc.onPodUpdate,
		DeleteFunc: rc.onPodDelete,
	})
	rc.informers["pods"] = podInformer
	
	// Deployment informer
	deploymentInformer := rc.informerFactory.Apps().V1().Deployments().Informer()
	deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    rc.onDeploymentAdd,
		UpdateFunc: rc.onDeploymentUpdate,
		DeleteFunc: rc.onDeploymentDelete,
	})
	rc.informers["deployments"] = deploymentInformer
	
	// Service informer
	serviceInformer := rc.informerFactory.Core().V1().Services().Informer()
	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    rc.onServiceAdd,
		UpdateFunc: rc.onServiceUpdate,
		DeleteFunc: rc.onServiceDelete,
	})
	rc.informers["services"] = serviceInformer
	
	// ReplicaSet informer
	replicaSetInformer := rc.informerFactory.Apps().V1().ReplicaSets().Informer()
	replicaSetInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    rc.onReplicaSetAdd,
		UpdateFunc: rc.onReplicaSetUpdate,
		DeleteFunc: rc.onReplicaSetDelete,
	})
	rc.informers["replicasets"] = replicaSetInformer
}

// Start begins the cache and informer operations
func (rc *ResourceCache) Start() {
	rc.logger.Info("Starting resource cache with informers")
	
	// Start all informers
	rc.informerFactory.Start(rc.ctx.Done())
	
	// Wait for cache sync
	for resourceType, informer := range rc.informers {
		rc.logger.Debug("Waiting for cache sync", "resource_type", resourceType)
		if !cache.WaitForCacheSync(rc.ctx.Done(), informer.HasSynced) {
			rc.logger.Error("Failed to sync cache", "resource_type", resourceType)
		}
	}
	
	rc.logger.Info("Resource cache started successfully")
}

// Stop gracefully stops the cache
func (rc *ResourceCache) Stop() {
	rc.logger.Info("Stopping resource cache")
	rc.cancel()
}

// Get retrieves a cached item or returns nil if not found/expired
func (rc *ResourceCache) Get(key string) interface{} {
	rc.cacheMu.RLock()
	entry, exists := rc.cache[key]
	rc.cacheMu.RUnlock()
	
	rc.statsMu.Lock()
	rc.stats.TotalRequests++
	if exists && !entry.IsExpired() {
		rc.stats.Hits++
		rc.statsMu.Unlock()
		rc.logger.Debug("Cache hit", "key", key)
		return entry.Data
	}
	rc.stats.Misses++
	rc.statsMu.Unlock()
	
	// Clean up expired entry
	if exists && entry.IsExpired() {
		rc.cacheMu.Lock()
		delete(rc.cache, key)
		rc.cacheMu.Unlock()
		rc.logger.Debug("Removed expired cache entry", "key", key)
	}
	
	rc.logger.Debug("Cache miss", "key", key)
	return nil
}

// Set stores an item in the cache with default TTL
func (rc *ResourceCache) Set(key string, data interface{}) {
	rc.SetWithTTL(key, data, rc.defaultTTL)
}

// SetWithTTL stores an item in the cache with custom TTL
func (rc *ResourceCache) SetWithTTL(key string, data interface{}, ttl time.Duration) {
	rc.cacheMu.Lock()
	defer rc.cacheMu.Unlock()
	
	// Check cache size and evict if necessary
	if len(rc.cache) >= rc.maxCacheSize {
		rc.evictOldest()
	}
	
	now := time.Now()
	entry := &CacheEntry{
		Data:      data,
		CachedAt:  now,
		ExpiresAt: now.Add(ttl),
		Version:   rc.getResourceVersion(data),
	}
	
	rc.cache[key] = entry
	rc.logger.Debug("Cached item", "key", key, "ttl", ttl)
}

// Delete removes an item from the cache
func (rc *ResourceCache) Delete(key string) {
	rc.cacheMu.Lock()
	defer rc.cacheMu.Unlock()
	
	if _, exists := rc.cache[key]; exists {
		delete(rc.cache, key)
		rc.logger.Debug("Deleted cache entry", "key", key)
	}
}

// GetPodList returns cached pod list or nil if not available
func (rc *ResourceCache) GetPodList(namespace string, selector string) *corev1.PodList {
	key := fmt.Sprintf("podlist:%s:%s", namespace, selector)
	if data := rc.Get(key); data != nil {
		if podList, ok := data.(*corev1.PodList); ok {
			return podList
		}
	}
	return nil
}

// SetPodList caches a pod list
func (rc *ResourceCache) SetPodList(namespace string, selector string, podList *corev1.PodList) {
	key := fmt.Sprintf("podlist:%s:%s", namespace, selector)
	rc.SetWithTTL(key, podList, 2*time.Minute) // Pods change frequently
}

// GetDeploymentList returns cached deployment list or nil if not available
func (rc *ResourceCache) GetDeploymentList(namespace string, selector string) *v1.DeploymentList {
	key := fmt.Sprintf("deploymentlist:%s:%s", namespace, selector)
	if data := rc.Get(key); data != nil {
		if deploymentList, ok := data.(*v1.DeploymentList); ok {
			return deploymentList
		}
	}
	return nil
}

// SetDeploymentList caches a deployment list
func (rc *ResourceCache) SetDeploymentList(namespace string, selector string, deploymentList *v1.DeploymentList) {
	key := fmt.Sprintf("deploymentlist:%s:%s", namespace, selector)
	rc.SetWithTTL(key, deploymentList, 5*time.Minute) // Deployments change less frequently
}

// GetServiceList returns cached service list or nil if not available
func (rc *ResourceCache) GetServiceList(namespace string, selector string) *corev1.ServiceList {
	key := fmt.Sprintf("servicelist:%s:%s", namespace, selector)
	if data := rc.Get(key); data != nil {
		if serviceList, ok := data.(*corev1.ServiceList); ok {
			return serviceList
		}
	}
	return nil
}

// SetServiceList caches a service list
func (rc *ResourceCache) SetServiceList(namespace string, selector string, serviceList *corev1.ServiceList) {
	key := fmt.Sprintf("servicelist:%s:%s", namespace, selector)
	rc.SetWithTTL(key, serviceList, 10*time.Minute) // Services change infrequently
}

// InvalidateNamespace removes all cached items for a specific namespace
func (rc *ResourceCache) InvalidateNamespace(namespace string) {
	rc.cacheMu.Lock()
	defer rc.cacheMu.Unlock()
	
	keysToDelete := make([]string, 0)
	for key := range rc.cache {
		if contains(key, namespace) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	
	for _, key := range keysToDelete {
		delete(rc.cache, key)
	}
	
	rc.logger.Info("Invalidated namespace cache", "namespace", namespace, "keys_removed", len(keysToDelete))
}

// GetStats returns current cache statistics
func (rc *ResourceCache) GetStats() CacheStats {
	rc.statsMu.RLock()
	defer rc.statsMu.RUnlock()
	
	rc.cacheMu.RLock()
	cacheSize := len(rc.cache)
	rc.cacheMu.RUnlock()
	
	stats := rc.stats
	stats.CacheSize = cacheSize
	return stats
}

// Event handlers for informers
func (rc *ResourceCache) onPodAdd(obj interface{}) {
	if pod, ok := obj.(*corev1.Pod); ok {
		rc.invalidateRelatedPodCaches(pod.Namespace)
		rc.logger.Debug("Pod added, invalidated related caches", "pod", pod.Name, "namespace", pod.Namespace)
	}
}

func (rc *ResourceCache) onPodUpdate(oldObj, newObj interface{}) {
	if pod, ok := newObj.(*corev1.Pod); ok {
		rc.invalidateRelatedPodCaches(pod.Namespace)
		rc.logger.Debug("Pod updated, invalidated related caches", "pod", pod.Name, "namespace", pod.Namespace)
	}
}

func (rc *ResourceCache) onPodDelete(obj interface{}) {
	if pod, ok := obj.(*corev1.Pod); ok {
		rc.invalidateRelatedPodCaches(pod.Namespace)
		rc.logger.Debug("Pod deleted, invalidated related caches", "pod", pod.Name, "namespace", pod.Namespace)
	}
}

func (rc *ResourceCache) onDeploymentAdd(obj interface{}) {
	if deployment, ok := obj.(*v1.Deployment); ok {
		rc.invalidateRelatedDeploymentCaches(deployment.Namespace)
	}
}

func (rc *ResourceCache) onDeploymentUpdate(oldObj, newObj interface{}) {
	if deployment, ok := newObj.(*v1.Deployment); ok {
		rc.invalidateRelatedDeploymentCaches(deployment.Namespace)
	}
}

func (rc *ResourceCache) onDeploymentDelete(obj interface{}) {
	if deployment, ok := obj.(*v1.Deployment); ok {
		rc.invalidateRelatedDeploymentCaches(deployment.Namespace)
	}
}

func (rc *ResourceCache) onServiceAdd(obj interface{}) {
	if service, ok := obj.(*corev1.Service); ok {
		rc.invalidateRelatedServiceCaches(service.Namespace)
	}
}

func (rc *ResourceCache) onServiceUpdate(oldObj, newObj interface{}) {
	if service, ok := newObj.(*corev1.Service); ok {
		rc.invalidateRelatedServiceCaches(service.Namespace)
	}
}

func (rc *ResourceCache) onServiceDelete(obj interface{}) {
	if service, ok := obj.(*corev1.Service); ok {
		rc.invalidateRelatedServiceCaches(service.Namespace)
	}
}

func (rc *ResourceCache) onReplicaSetAdd(obj interface{}) {
	if rs, ok := obj.(*v1.ReplicaSet); ok {
		rc.invalidateRelatedPodCaches(rs.Namespace)
		rc.invalidateRelatedDeploymentCaches(rs.Namespace)
	}
}

func (rc *ResourceCache) onReplicaSetUpdate(oldObj, newObj interface{}) {
	if rs, ok := newObj.(*v1.ReplicaSet); ok {
		rc.invalidateRelatedPodCaches(rs.Namespace)
		rc.invalidateRelatedDeploymentCaches(rs.Namespace)
	}
}

func (rc *ResourceCache) onReplicaSetDelete(obj interface{}) {
	if rs, ok := obj.(*v1.ReplicaSet); ok {
		rc.invalidateRelatedPodCaches(rs.Namespace)
		rc.invalidateRelatedDeploymentCaches(rs.Namespace)
	}
}

// Helper methods for cache invalidation
func (rc *ResourceCache) invalidateRelatedPodCaches(namespace string) {
	rc.cacheMu.Lock()
	defer rc.cacheMu.Unlock()
	
	keysToDelete := make([]string, 0)
	for key := range rc.cache {
		if contains(key, "podlist:") && contains(key, namespace) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	
	for _, key := range keysToDelete {
		delete(rc.cache, key)
	}
}

func (rc *ResourceCache) invalidateRelatedDeploymentCaches(namespace string) {
	rc.cacheMu.Lock()
	defer rc.cacheMu.Unlock()
	
	keysToDelete := make([]string, 0)
	for key := range rc.cache {
		if contains(key, "deploymentlist:") && contains(key, namespace) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	
	for _, key := range keysToDelete {
		delete(rc.cache, key)
	}
}

func (rc *ResourceCache) invalidateRelatedServiceCaches(namespace string) {
	rc.cacheMu.Lock()
	defer rc.cacheMu.Unlock()
	
	keysToDelete := make([]string, 0)
	for key := range rc.cache {
		if contains(key, "servicelist:") && contains(key, namespace) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	
	for _, key := range keysToDelete {
		delete(rc.cache, key)
	}
}

// evictOldest removes the oldest cache entry (LRU eviction)
func (rc *ResourceCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time = time.Now()
	
	for key, entry := range rc.cache {
		if entry.CachedAt.Before(oldestTime) {
			oldestTime = entry.CachedAt
			oldestKey = key
		}
	}
	
	if oldestKey != "" {
		delete(rc.cache, oldestKey)
		rc.statsMu.Lock()
		rc.stats.Evictions++
		rc.stats.LastEviction = time.Now()
		rc.statsMu.Unlock()
		rc.logger.Debug("Evicted oldest cache entry", "key", oldestKey)
	}
}

// cleanupRoutine periodically removes expired entries
func (rc *ResourceCache) cleanupRoutine() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-rc.ctx.Done():
			return
		case <-ticker.C:
			rc.cleanup()
		}
	}
}

// cleanup removes expired entries
func (rc *ResourceCache) cleanup() {
	rc.cacheMu.Lock()
	defer rc.cacheMu.Unlock()
	
	now := time.Now()
	keysToDelete := make([]string, 0)
	
	for key, entry := range rc.cache {
		if now.After(entry.ExpiresAt) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	
	for _, key := range keysToDelete {
		delete(rc.cache, key)
	}
	
	if len(keysToDelete) > 0 {
		rc.logger.Debug("Cleaned up expired cache entries", "count", len(keysToDelete))
	}
}

// getResourceVersion extracts resource version from a Kubernetes object
func (rc *ResourceCache) getResourceVersion(obj interface{}) string {
	if metaObj, ok := obj.(metav1.Object); ok {
		return metaObj.GetResourceVersion()
	}
	if runtimeObj, ok := obj.(runtime.Object); ok {
		if gvk := runtimeObj.GetObjectKind().GroupVersionKind(); gvk.Version != "" {
			return ""
		}
	}
	return ""
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     func() bool {
		         for i := 0; i <= len(s)-len(substr); i++ {
		             if s[i:i+len(substr)] == substr {
		                 return true
		             }
		         }
		         return false
		     }()))
}