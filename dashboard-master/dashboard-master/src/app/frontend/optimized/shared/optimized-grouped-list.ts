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

import {Directive} from '@angular/core';
import {
  CronJobList,
  DaemonSetList,
  DeploymentList,
  JobList,
  Metric,
  PodList,
  ReplicaSetList,
  ReplicationControllerList,
  ResourceList,
  StatefulSetList,
} from '@api/root.api';
import {OnListChangeEvent, ResourcesRatio} from '@api/root.ui';
import {BehaviorSubject, Observable} from 'rxjs';
import {map, distinctUntilChanged} from 'rxjs/operators';

import {Helper, ResourceRatioModes} from '../../overview/helper';
import {ListGroupIdentifier, ListIdentifier} from '../../common/components/resourcelist/groupids';
import {emptyResourcesRatio} from '../../common/components/workloadstatus/component';

interface OptimizedGroupState {
  items: {[id: string]: number};
  groupItems: {[groupId: string]: {[id: string]: number}};
  resourcesRatio: ResourcesRatio;
  cumulativeMetrics: Metric[];
}

@Directive()
export abstract class OptimizedGroupedResourceList {
  // Observable state management for reactive updates
  private readonly state$ = new BehaviorSubject<OptimizedGroupState>({
    items: {},
    groupItems: {
      [ListGroupIdentifier.cluster]: {},
      [ListGroupIdentifier.workloads]: {},
      [ListGroupIdentifier.discovery]: {},
      [ListGroupIdentifier.config]: {},
    },
    resourcesRatio: emptyResourcesRatio,
    cumulativeMetrics: [],
  });

  // Public observables for components to subscribe to
  readonly items$ = this.state$.pipe(
    map(state => state.items),
    distinctUntilChanged((a, b) => JSON.stringify(a) === JSON.stringify(b))
  );

  readonly resourcesRatio$ = this.state$.pipe(
    map(state => state.resourcesRatio),
    distinctUntilChanged((a, b) => JSON.stringify(a) === JSON.stringify(b))
  );

  readonly cumulativeMetrics$ = this.state$.pipe(
    map(state => state.cumulativeMetrics),
    distinctUntilChanged((a, b) => JSON.stringify(a) === JSON.stringify(b))
  );

  // Computed observables for zero state and group visibility
  readonly shouldShowZeroState$ = this.items$.pipe(
    map(items => {
      let totalItems = 0;
      const ids = Object.keys(items);
      ids.forEach(id => (totalItems += items[id]));
      return totalItems === 0 && ids.length > 0;
    }),
    distinctUntilChanged()
  );

  // Getters for backward compatibility
  get resourcesRatio(): ResourcesRatio {
    return this.state$.value.resourcesRatio;
  }

  get cumulativeMetrics(): Metric[] {
    return this.state$.value.cumulativeMetrics;
  }

  shouldShowZeroState(): boolean {
    const state = this.state$.value;
    let totalItems = 0;
    const ids = Object.keys(state.items);
    ids.forEach(id => (totalItems += state.items[id]));
    return totalItems === 0 && ids.length > 0;
  }

  isGroupVisible(groupId: string): boolean {
    const state = this.state$.value;
    let totalItems = 0;
    const ids = Object.keys(state.groupItems[groupId] || {});
    ids.forEach(id => (totalItems += state.groupItems[groupId][id]));
    return totalItems > 0;
  }

  getGroupVisibility(groupId: string): Observable<boolean> {
    return this.state$.pipe(
      map(state => {
        let totalItems = 0;
        const ids = Object.keys(state.groupItems[groupId] || {});
        ids.forEach(id => (totalItems += state.groupItems[groupId][id]));
        return totalItems > 0;
      }),
      distinctUntilChanged()
    );
  }

  onListUpdate(listEvent: OnListChangeEvent): void {
    const currentState = this.state$.value;
    
    // Create new state object for immutability
    const newState: OptimizedGroupState = {
      items: {...currentState.items},
      groupItems: {
        ...currentState.groupItems,
        [listEvent.groupId]: {
          ...currentState.groupItems[listEvent.groupId],
          [listEvent.id]: listEvent.items
        }
      },
      resourcesRatio: {...currentState.resourcesRatio},
      cumulativeMetrics: [...currentState.cumulativeMetrics],
    };

    // Update items count
    newState.items[listEvent.id] = listEvent.items;

    // Handle filtered state
    if (listEvent.filtered) {
      newState.items[listEvent.id] = 1;
    }

    // Update resources ratio and metrics
    this.updateResourcesRatio(listEvent.id, listEvent.resourceList, newState);

    // Emit new state
    this.state$.next(newState);
  }

  private updateResourcesRatio(identifier: ListIdentifier, list: ResourceList, state: OptimizedGroupState): void {
    // Optimize switch statement with early returns and efficient processing
    switch (identifier) {
      case ListIdentifier.cronJob:
        this.updateCronJobRatio(list as CronJobList, state);
        break;
      case ListIdentifier.daemonSet:
        this.updateDaemonSetRatio(list as DaemonSetList, state);
        break;
      case ListIdentifier.deployment:
        this.updateDeploymentRatio(list as DeploymentList, state);
        break;
      case ListIdentifier.job:
        this.updateJobRatio(list as JobList, state);
        break;
      case ListIdentifier.pod:
        this.updatePodRatio(list as PodList, state);
        break;
      case ListIdentifier.replicaSet:
        this.updateReplicaSetRatio(list as ReplicaSetList, state);
        break;
      case ListIdentifier.replicationController:
        this.updateReplicationControllerRatio(list as ReplicationControllerList, state);
        break;
      case ListIdentifier.statefulSet:
        this.updateStatefulSetRatio(list as StatefulSetList, state);
        break;
    }
  }

  private updateCronJobRatio(cronJobs: CronJobList, state: OptimizedGroupState): void {
    state.resourcesRatio.cronJobRatio = Helper.getResourceRatio(
      cronJobs.status,
      cronJobs.listMeta.totalItems,
      ResourceRatioModes.Suspendable
    );
  }

  private updateDaemonSetRatio(daemonSets: DaemonSetList, state: OptimizedGroupState): void {
    state.resourcesRatio.daemonSetRatio = Helper.getResourceRatio(
      daemonSets.status,
      daemonSets.listMeta.totalItems
    );
  }

  private updateDeploymentRatio(deployments: DeploymentList, state: OptimizedGroupState): void {
    state.resourcesRatio.deploymentRatio = Helper.getResourceRatio(
      deployments.status,
      deployments.listMeta.totalItems
    );
  }

  private updateJobRatio(jobs: JobList, state: OptimizedGroupState): void {
    state.resourcesRatio.jobRatio = Helper.getResourceRatio(
      jobs.status,
      jobs.listMeta.totalItems,
      ResourceRatioModes.Completable
    );
  }

  private updatePodRatio(pods: PodList, state: OptimizedGroupState): void {
    state.resourcesRatio.podRatio = Helper.getResourceRatio(
      pods.status,
      pods.listMeta.totalItems,
      ResourceRatioModes.Completable
    );
    // Update cumulative metrics for pods
    state.cumulativeMetrics = pods.cumulativeMetrics || [];
  }

  private updateReplicaSetRatio(replicaSets: ReplicaSetList, state: OptimizedGroupState): void {
    state.resourcesRatio.replicaSetRatio = Helper.getResourceRatio(
      replicaSets.status,
      replicaSets.listMeta.totalItems
    );
  }

  private updateReplicationControllerRatio(replicationControllers: ReplicationControllerList, state: OptimizedGroupState): void {
    state.resourcesRatio.replicationControllerRatio = Helper.getResourceRatio(
      replicationControllers.status,
      replicationControllers.listMeta.totalItems
    );
  }

  private updateStatefulSetRatio(statefulSets: StatefulSetList, state: OptimizedGroupState): void {
    state.resourcesRatio.statefulSetRatio = Helper.getResourceRatio(
      statefulSets.status,
      statefulSets.listMeta.totalItems
    );
  }

  // Performance monitoring helpers
  getStateSnapshot(): OptimizedGroupState {
    return this.state$.value;
  }

  // Memory cleanup helper
  cleanup(): void {
    // Reset state to initial values for garbage collection
    this.state$.next({
      items: {},
      groupItems: {
        [ListGroupIdentifier.cluster]: {},
        [ListGroupIdentifier.workloads]: {},
        [ListGroupIdentifier.discovery]: {},
        [ListGroupIdentifier.config]: {},
      },
      resourcesRatio: emptyResourcesRatio,
      cumulativeMetrics: [],
    });
  }
}