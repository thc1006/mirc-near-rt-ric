// Copyright 2024 The O-RAN Near-RT RIC Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, BehaviorSubject, interval } from 'rxjs';
import { map, switchMap } from 'rxjs/operators';

export interface YangNode {
  id: string;
  name: string;
  type: 'container' | 'leaf' | 'leaf-list' | 'list' | 'choice' | 'case';
  path: string;
  value?: any;
  children?: YangNode[];
  metadata?: {
    namespace?: string;
    description?: string;
    status?: 'current' | 'deprecated' | 'obsolete';
    mandatory?: boolean;
    config?: boolean;
  };
  metrics?: {
    lastUpdated: Date;
    updateCount: number;
    errorCount: number;
  };
}

export interface TimeSeriesData {
  timestamp: Date;
  nodeId: string;
  value: number;
  unit?: string;
  quality?: 'good' | 'bad' | 'uncertain';
}

export interface XAppMetrics {
  id: string;
  name: string;
  status: 'running' | 'stopped' | 'error' | 'starting';
  cpu: number;
  memory: number;
  network: {
    bytesIn: number;
    bytesOut: number;
    packetsIn: number;
    packetsOut: number;
  };
  rrmMetrics: {
    e2Latency: number;
    throughput: number;
    resourceUtilization: number;
    handoverSuccess: number;
  };
  timestamp: Date;
}

@Injectable({
  providedIn: 'root'
})
export class YangDataService {
  private yangDataSubject = new BehaviorSubject<YangNode[]>([]);
  private timeSeriesSubject = new BehaviorSubject<TimeSeriesData[]>([]);
  private xAppMetricsSubject = new BehaviorSubject<XAppMetrics[]>([]);

  public yangData$ = this.yangDataSubject.asObservable();
  public timeSeriesData$ = this.timeSeriesSubject.asObservable();
  public xAppMetrics$ = this.xAppMetricsSubject.asObservable();

  constructor(private http: HttpClient) {
    this.initializeRealTimeUpdates();
  }

  /**
   * Load YANG model data from JSON files
   */
  loadYangData(): Observable<YangNode[]> {
    return this.http.get<any[]>('assets/data/sysrepo-netopeer1.json').pipe(
      map(data => this.transformToYangModel(data))
    );
  }

  /**
   * Transform raw JSON data to YANG model structure
   */
  private transformToYangModel(rawData: any[]): YangNode[] {
    const rootNodes: YangNode[] = [
      {
        id: 'o-ran-sc',
        name: 'o-ran-sc',
        type: 'container',
        path: '/o-ran-sc',
        metadata: {
          namespace: 'urn:o-ran:sc:yang:1.0',
          description: 'O-RAN Software Community YANG models',
          status: 'current'
        },
        children: [
          {
            id: 'near-rt-ric',
            name: 'near-rt-ric',
            type: 'container',
            path: '/o-ran-sc/near-rt-ric',
            metadata: {
              description: 'Near Real-Time RAN Intelligent Controller',
              status: 'current'
            },
            children: [
              {
                id: 'xapps',
                name: 'xapps',
                type: 'list',
                path: '/o-ran-sc/near-rt-ric/xapps',
                children: this.generateXAppNodes(rawData)
              },
              {
                id: 'e2-interface',
                name: 'e2-interface',
                type: 'container',
                path: '/o-ran-sc/near-rt-ric/e2-interface',
                children: this.generateE2InterfaceNodes()
              },
              {
                id: 'a1-interface',
                name: 'a1-interface', 
                type: 'container',
                path: '/o-ran-sc/near-rt-ric/a1-interface',
                children: this.generateA1InterfaceNodes()
              },
              {
                id: 'federated-learning',
                name: 'federated-learning',
                type: 'container',
                path: '/o-ran-sc/near-rt-ric/federated-learning',
                children: this.generateFederatedLearningNodes()
              }
            ]
          }
        ]
      }
    ];

    this.yangDataSubject.next(rootNodes);
    return rootNodes;
  }

  /**
   * Generate xApp nodes from container data
   */
  private generateXAppNodes(rawData: any[]): YangNode[] {
    return rawData.map((item, index) => ({
      id: `xapp-${item.Id || index}`,
      name: `xapp-${item.Id || index}`,
      type: 'list' as const,
      path: `/o-ran-sc/near-rt-ric/xapps/xapp-${item.Id || index}`,
      children: [
        {
          id: `${item.Id}-container-id`,
          name: 'container-id',
          type: 'leaf' as const,
          path: `/o-ran-sc/near-rt-ric/xapps/xapp-${item.Id || index}/container-id`,
          value: item.Id || item.Image,
          metadata: { description: 'Container identifier', mandatory: true }
        },
        {
          id: `${item.Id}-image-tag`,
          name: 'image-tag',
          type: 'leaf' as const,
          path: `/o-ran-sc/near-rt-ric/xapps/xapp-${item.Id || index}/image-tag`,
          value: item.Tag || 'latest',
          metadata: { description: 'Container image tag' }
        },
        {
          id: `${item.Id}-size`,
          name: 'size',
          type: 'leaf' as const,
          path: `/o-ran-sc/near-rt-ric/xapps/xapp-${item.Id || index}/size`,
          value: item.Size,
          metadata: { description: 'Container size' }
        },
        {
          id: `${item.Id}-command`,
          name: 'command',
          type: 'leaf' as const,
          path: `/o-ran-sc/near-rt-ric/xapps/xapp-${item.Id || index}/command`,
          value: item.Cmd,
          metadata: { description: 'Container startup command' }
        },
        {
          id: `${item.Id}-metrics`,
          name: 'metrics',
          type: 'container' as const,
          path: `/o-ran-sc/near-rt-ric/xapps/xapp-${item.Id || index}/metrics`,
          children: this.generateMetricsNodes(item.Id || index)
        }
      ],
      metrics: {
        lastUpdated: new Date(),
        updateCount: Math.floor(Math.random() * 100),
        errorCount: Math.floor(Math.random() * 5)
      }
    }));
  }

  /**
   * Generate E2 interface nodes
   */
  private generateE2InterfaceNodes(): YangNode[] {
    return [
      {
        id: 'e2-connections',
        name: 'connections',
        type: 'list',
        path: '/o-ran-sc/near-rt-ric/e2-interface/connections',
        children: [
          {
            id: 'e2-node-1',
            name: 'e2-node-1',
            type: 'list',
            path: '/o-ran-sc/near-rt-ric/e2-interface/connections/e2-node-1',
            children: [
              {
                id: 'e2-node-id',
                name: 'node-id',
                type: 'leaf',
                path: '/o-ran-sc/near-rt-ric/e2-interface/connections/e2-node-1/node-id',
                value: 'gnb-001',
                metadata: { description: 'E2 node identifier' }
              },
              {
                id: 'e2-latency',
                name: 'latency-ms',
                type: 'leaf',
                path: '/o-ran-sc/near-rt-ric/e2-interface/connections/e2-node-1/latency-ms',
                value: 15.5,
                metadata: { description: 'E2 interface latency in milliseconds' }
              }
            ]
          }
        ]
      }
    ];
  }

  /**
   * Generate A1 interface nodes
   */
  private generateA1InterfaceNodes(): YangNode[] {
    return [
      {
        id: 'a1-policies',
        name: 'policies',
        type: 'list',
        path: '/o-ran-sc/near-rt-ric/a1-interface/policies',
        children: [
          {
            id: 'policy-1',
            name: 'policy-1',
            type: 'list',
            path: '/o-ran-sc/near-rt-ric/a1-interface/policies/policy-1',
            children: [
              {
                id: 'policy-type',
                name: 'policy-type',
                type: 'leaf',
                path: '/o-ran-sc/near-rt-ric/a1-interface/policies/policy-1/policy-type',
                value: 'QoS',
                metadata: { description: 'Policy type identifier' }
              }
            ]
          }
        ]
      }
    ];
  }

  /**
   * Generate federated learning nodes
   */
  private generateFederatedLearningNodes(): YangNode[] {
    return [
      {
        id: 'fl-coordinator',
        name: 'coordinator',
        type: 'container',
        path: '/o-ran-sc/near-rt-ric/federated-learning/coordinator',
        children: [
          {
            id: 'fl-status',
            name: 'status',
            type: 'leaf',
            path: '/o-ran-sc/near-rt-ric/federated-learning/coordinator/status',
            value: 'active',
            metadata: { description: 'FL coordinator status' }
          },
          {
            id: 'fl-clients',
            name: 'registered-clients',
            type: 'leaf',
            path: '/o-ran-sc/near-rt-ric/federated-learning/coordinator/registered-clients',
            value: 5,
            metadata: { description: 'Number of registered FL clients' }
          }
        ]
      }
    ];
  }

  /**
   * Generate metrics nodes for containers
   */
  private generateMetricsNodes(containerId: string | number): YangNode[] {
    return [
      {
        id: `${containerId}-cpu`,
        name: 'cpu-usage',
        type: 'leaf',
        path: `/metrics/cpu-usage`,
        value: Math.random() * 100,
        metadata: { description: 'CPU usage percentage' }
      },
      {
        id: `${containerId}-memory`,
        name: 'memory-usage',
        type: 'leaf',
        path: `/metrics/memory-usage`,
        value: Math.random() * 100,
        metadata: { description: 'Memory usage percentage' }
      },
      {
        id: `${containerId}-network`,
        name: 'network-io',
        type: 'container',
        path: `/metrics/network-io`,
        children: [
          {
            id: `${containerId}-bytes-in`,
            name: 'bytes-in',
            type: 'leaf',
            path: `/metrics/network-io/bytes-in`,
            value: Math.floor(Math.random() * 1000000),
            metadata: { description: 'Network bytes received' }
          },
          {
            id: `${containerId}-bytes-out`,
            name: 'bytes-out',
            type: 'leaf',
            path: `/metrics/network-io/bytes-out`,
            value: Math.floor(Math.random() * 1000000),
            metadata: { description: 'Network bytes transmitted' }
          }
        ]
      }
    ];
  }

  /**
   * Initialize real-time data updates
   */
  private initializeRealTimeUpdates(): void {
    // Update time series data every 5 seconds
    interval(5000).pipe(
      switchMap(() => this.generateTimeSeriesData())
    ).subscribe(data => {
      const currentData = this.timeSeriesSubject.value;
      const updatedData = [...currentData, ...data].slice(-1000); // Keep last 1000 points
      this.timeSeriesSubject.next(updatedData);
    });

    // Update xApp metrics every 10 seconds
    interval(10000).subscribe(() => {
      const metrics = this.generateXAppMetrics();
      this.xAppMetricsSubject.next(metrics);
    });
  }

  /**
   * Generate time series data for visualization
   */
  private generateTimeSeriesData(): Observable<TimeSeriesData[]> {
    const data: TimeSeriesData[] = [
      {
        timestamp: new Date(),
        nodeId: 'e2-latency',
        value: 10 + Math.random() * 20,
        unit: 'ms',
        quality: 'good'
      },
      {
        timestamp: new Date(),
        nodeId: 'cpu-usage',
        value: 30 + Math.random() * 40,
        unit: '%',
        quality: 'good'
      },
      {
        timestamp: new Date(),
        nodeId: 'memory-usage',
        value: 40 + Math.random() * 30,
        unit: '%',
        quality: 'good'
      },
      {
        timestamp: new Date(),
        nodeId: 'throughput',
        value: 100 + Math.random() * 200,
        unit: 'Mbps',
        quality: 'good'
      }
    ];

    return new Observable(observer => {
      observer.next(data);
      observer.complete();
    });
  }

  /**
   * Generate xApp metrics data
   */
  private generateXAppMetrics(): XAppMetrics[] {
    return [
      {
        id: 'xapp-1',
        name: 'Traffic Prediction xApp',
        status: 'running',
        cpu: 30 + Math.random() * 40,
        memory: 40 + Math.random() * 30,
        network: {
          bytesIn: Math.floor(Math.random() * 1000000),
          bytesOut: Math.floor(Math.random() * 1000000),
          packetsIn: Math.floor(Math.random() * 10000),
          packetsOut: Math.floor(Math.random() * 10000)
        },
        rrmMetrics: {
          e2Latency: 10 + Math.random() * 20,
          throughput: 100 + Math.random() * 200,
          resourceUtilization: 60 + Math.random() * 30,
          handoverSuccess: 85 + Math.random() * 10
        },
        timestamp: new Date()
      },
      {
        id: 'xapp-2',
        name: 'QoS Optimization xApp',
        status: 'running',
        cpu: 25 + Math.random() * 35,
        memory: 35 + Math.random() * 25,
        network: {
          bytesIn: Math.floor(Math.random() * 800000),
          bytesOut: Math.floor(Math.random() * 800000),
          packetsIn: Math.floor(Math.random() * 8000),
          packetsOut: Math.floor(Math.random() * 8000)
        },
        rrmMetrics: {
          e2Latency: 8 + Math.random() * 15,
          throughput: 120 + Math.random() * 180,
          resourceUtilization: 55 + Math.random() * 35,
          handoverSuccess: 88 + Math.random() * 8
        },
        timestamp: new Date()
      }
    ];
  }

  /**
   * Get node by path
   */
  getNodeByPath(path: string): Observable<YangNode | null> {
    return this.yangData$.pipe(
      map(nodes => this.findNodeByPath(nodes, path))
    );
  }

  /**
   * Find node by path recursively
   */
  private findNodeByPath(nodes: YangNode[], path: string): YangNode | null {
    for (const node of nodes) {
      if (node.path === path) {
        return node;
      }
      if (node.children) {
        const found = this.findNodeByPath(node.children, path);
        if (found) return found;
      }
    }
    return null;
  }

  /**
   * Get time series data for specific node
   */
  getTimeSeriesForNode(nodeId: string): Observable<TimeSeriesData[]> {
    return this.timeSeriesData$.pipe(
      map(data => data.filter(point => point.nodeId === nodeId))
    );
  }
}