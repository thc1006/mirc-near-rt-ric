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

import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  OnDestroy,
  OnInit,
  QueryList,
  ViewChildren,
  Injectable,
  TrackByFunction
} from '@angular/core';
import {Observable, Subject, combineLatest, merge, BehaviorSubject} from 'rxjs';
import {takeUntil, debounceTime, distinctUntilChanged, map, startWith} from 'rxjs/operators';

import {ListGroupIdentifier} from '@common/components/resourcelist/groupids';
import {OptimizedGroupedResourceList} from '../shared/optimized-grouped-list';
import {ResourceCacheService} from '../shared/cache.service';
import {OnListChangeEvent, ResourcesRatio} from '@api/root.ui';
import {Metric} from '@api/root.api';

// Define interfaces for better type safety
interface WorkloadGroup {
  id: string;
  visible: boolean;
  components: WorkloadComponent[];
}

interface WorkloadComponent {
  selector: string;
  loaded: boolean;
  error?: Error;
  lastUpdate: Date;
}

// Optimized data structure for tracking visibility
interface VisibilityState {
  [groupId: string]: boolean;
}

@Injectable()
export class OverviewStateService {
  private visibilityState$ = new BehaviorSubject<VisibilityState>({
    [ListGroupIdentifier.cluster]: false,
    [ListGroupIdentifier.workloads]: false,
    [ListGroupIdentifier.discovery]: false,
    [ListGroupIdentifier.config]: false,
  });

  private resourcesRatio$ = new BehaviorSubject<ResourcesRatio>({});
  private cumulativeMetrics$ = new BehaviorSubject<Metric[]>([]);

  getVisibilityState(): Observable<VisibilityState> {
    return this.visibilityState$.asObservable();
  }

  updateGroupVisibility(groupId: string, visible: boolean): void {
    const current = this.visibilityState$.value;
    this.visibilityState$.next({
      ...current,
      [groupId]: visible
    });
  }

  getResourcesRatio(): Observable<ResourcesRatio> {
    return this.resourcesRatio$.asObservable();
  }

  updateResourcesRatio(ratio: ResourcesRatio): void {
    this.resourcesRatio$.next(ratio);
  }

  getCumulativeMetrics(): Observable<Metric[]> {
    return this.cumulativeMetrics$.asObservable();
  }

  updateCumulativeMetrics(metrics: Metric[]): void {
    this.cumulativeMetrics$.next(metrics);
  }
}

@Component({
  selector: 'kd-optimized-overview',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
  changeDetection: ChangeDetectionStrategy.OnPush,
  providers: [OverviewStateService]
})
export class OptimizedOverviewComponent extends OptimizedGroupedResourceList implements OnInit, OnDestroy {
  private readonly destroy$ = new Subject<void>();
  
  // Observables for reactive updates
  readonly visibilityState$ = this.stateService.getVisibilityState();
  readonly resourcesRatio$ = this.stateService.getResourcesRatio();
  readonly cumulativeMetrics$ = this.stateService.getCumulativeMetrics();
  
  // Computed observables
  readonly hasCluster$ = this.visibilityState$.pipe(
    map(state => state[ListGroupIdentifier.cluster]),
    distinctUntilChanged()
  );
  
  readonly hasWorkloads$ = this.visibilityState$.pipe(
    map(state => state[ListGroupIdentifier.workloads]),
    distinctUntilChanged()
  );
  
  readonly hasDiscovery$ = this.visibilityState$.pipe(
    map(state => state[ListGroupIdentifier.discovery]),
    distinctUntilChanged()
  );
  
  readonly hasConfig$ = this.visibilityState$.pipe(
    map(state => state[ListGroupIdentifier.config]),
    distinctUntilChanged()
  );
  
  readonly showWorkloadStatuses$ = this.resourcesRatio$.pipe(
    map(ratio => Object.values(ratio).reduce((sum, ratioItems) => sum + (ratioItems?.length || 0), 0) !== 0),
    distinctUntilChanged()
  );

  // Performance optimization: Track by functions for *ngFor
  readonly trackByGroupId: TrackByFunction<{id: string}> = (index, item) => item.id;
  readonly trackByWorkloadComponent: TrackByFunction<WorkloadComponent> = (index, item) => item.selector;

  // Child component references for manual change detection control
  @ViewChildren('workloadComponents') workloadComponents: QueryList<any>;

  // Debounced update stream to batch multiple changes
  private readonly updateStream$ = new Subject<OnListChangeEvent>();
  private readonly debouncedUpdates$ = this.updateStream$.pipe(
    debounceTime(50), // Batch updates within 50ms
    distinctUntilChanged((a, b) => 
      a.id === b.id && 
      a.items === b.items && 
      a.filtered === b.filtered
    )
  );

  constructor(
    private readonly cdr: ChangeDetectorRef,
    private readonly stateService: OverviewStateService,
    private readonly cacheService: ResourceCacheService
  ) {
    super();
  }

  ngOnInit(): void {
    this.setupReactiveUpdates();
    this.setupPerformanceOptimizations();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  /**
   * Optimized list update handler with batching and caching
   */
  onListUpdate(event: OnListChangeEvent): void {
    // Cache the event for performance
    this.cacheService.cacheListEvent(event);
    
    // Emit to debounced stream
    this.updateStream$.next(event);
  }

  /**
   * Setup reactive data streams for optimal performance
   */
  private setupReactiveUpdates(): void {
    // Subscribe to debounced updates
    this.debouncedUpdates$.pipe(
      takeUntil(this.destroy$)
    ).subscribe(event => {
      this.processListUpdate(event);
      
      // Only trigger change detection if component is visible
      if (this.isComponentVisible(event.groupId)) {
        this.cdr.markForCheck();
      }
    });

    // Setup visibility state tracking
    this.setupVisibilityTracking();
  }

  /**
   * Setup performance optimizations
   */
  private setupPerformanceOptimizations(): void {
    // Implement intersection observer for lazy loading
    this.setupIntersectionObserver();
    
    // Setup memory management
    this.setupMemoryManagement();
  }

  /**
   * Setup visibility tracking for groups
   */
  private setupVisibilityTracking(): void {
    // Combine multiple visibility streams efficiently
    const visibilityStreams = [
      this.hasCluster$,
      this.hasWorkloads$,
      this.hasDiscovery$,
      this.hasConfig$
    ];

    combineLatest(visibilityStreams).pipe(
      takeUntil(this.destroy$),
      debounceTime(10) // Prevent excessive updates
    ).subscribe(([cluster, workloads, discovery, config]) => {
      // Update visibility state efficiently
      this.stateService.updateGroupVisibility(ListGroupIdentifier.cluster, cluster);
      this.stateService.updateGroupVisibility(ListGroupIdentifier.workloads, workloads);
      this.stateService.updateGroupVisibility(ListGroupIdentifier.discovery, discovery);
      this.stateService.updateGroupVisibility(ListGroupIdentifier.config, config);
    });
  }

  /**
   * Setup intersection observer for lazy loading components
   */
  private setupIntersectionObserver(): void {
    if (typeof IntersectionObserver !== 'undefined') {
      const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            // Load component data when it becomes visible
            const componentId = entry.target.getAttribute('data-component-id');
            if (componentId) {
              this.loadComponentData(componentId);
            }
          }
        });
      }, {
        rootMargin: '50px', // Load 50px before visible
        threshold: 0.1
      });

      // Observe component containers after view init
      setTimeout(() => {
        document.querySelectorAll('[data-component-id]').forEach(el => {
          observer.observe(el);
        });
      });
    }
  }

  /**
   * Setup memory management for large lists
   */
  private setupMemoryManagement(): void {
    // Clean up cached data periodically
    const cleanupInterval = setInterval(() => {
      this.cacheService.cleanup();
    }, 60000); // Every minute

    // Clean up on destroy
    this.destroy$.subscribe(() => {
      clearInterval(cleanupInterval);
    });
  }

  /**
   * Process list update with optimizations
   */
  private processListUpdate(event: OnListChangeEvent): void {
    // Update parent class state
    super.onListUpdate(event);

    // Update reactive state
    this.stateService.updateResourcesRatio(this.resourcesRatio);
    this.stateService.updateCumulativeMetrics(this.cumulativeMetrics);

    // Update group visibility
    const groupVisible = this.isGroupVisible(event.groupId);
    this.stateService.updateGroupVisibility(event.groupId, groupVisible);
  }

  /**
   * Check if component is currently visible
   */
  private isComponentVisible(groupId: string): boolean {
    const element = document.querySelector(`[data-group-id="${groupId}"]`);
    if (!element) return true; // Assume visible if element not found

    const rect = element.getBoundingClientRect();
    return rect.top < window.innerHeight && rect.bottom > 0;
  }

  /**
   * Load component data lazily
   */
  private loadComponentData(componentId: string): void {
    // Implementation would trigger data loading for specific component
    // This prevents loading all data upfront
    console.log(`Loading data for component: ${componentId}`);
  }

  /**
   * Optimized method to check zero state with caching
   */
  shouldShowZeroState(): Observable<boolean> {
    return this.visibilityState$.pipe(
      map(state => {
        const hasAnyVisible = Object.values(state).some(visible => visible);
        return !hasAnyVisible;
      }),
      distinctUntilChanged()
    );
  }

  /**
   * Get workload groups for template iteration
   */
  getWorkloadGroups(): WorkloadGroup[] {
    return [
      {
        id: 'cronjob',
        visible: true,
        components: [{ selector: 'kd-cron-job-list', loaded: false, lastUpdate: new Date() }]
      },
      {
        id: 'daemonset',
        visible: true,
        components: [{ selector: 'kd-daemon-set-list', loaded: false, lastUpdate: new Date() }]
      },
      {
        id: 'deployment',
        visible: true,
        components: [{ selector: 'kd-deployment-list', loaded: false, lastUpdate: new Date() }]
      },
      {
        id: 'job',
        visible: true,
        components: [{ selector: 'kd-job-list', loaded: false, lastUpdate: new Date() }]
      },
      {
        id: 'pod',
        visible: true,
        components: [{ selector: 'kd-pod-list', loaded: false, lastUpdate: new Date() }]
      },
      {
        id: 'replicaset',
        visible: true,
        components: [{ selector: 'kd-replica-set-list', loaded: false, lastUpdate: new Date() }]
      },
      {
        id: 'replicationcontroller',
        visible: true,
        components: [{ selector: 'kd-replication-controller-list', loaded: false, lastUpdate: new Date() }]
      },
      {
        id: 'statefulset',
        visible: true,
        components: [{ selector: 'kd-stateful-set-list', loaded: false, lastUpdate: new Date() }]
      }
    ];
  }

  /**
   * Performance monitoring helper
   */
  trackPerformance(eventName: string, startTime: number): void {
    const duration = performance.now() - startTime;
    console.log(`[Performance] ${eventName}: ${duration.toFixed(2)}ms`);
  }
}