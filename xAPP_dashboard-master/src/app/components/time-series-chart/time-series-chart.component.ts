// Copyright 2024 The O-RAN Near-RT RIC Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

import { 
  Component, 
  OnInit, 
  OnDestroy, 
  ElementRef, 
  ViewChild, 
  Input,
  ChangeDetectorRef 
} from '@angular/core';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import * as echarts from 'echarts';

import { YangDataService, TimeSeriesData, XAppMetrics } from '../../services/yang-data.service';

interface ChartMetric {
  id: string;
  name: string;
  unit: string;
  color: string;
  enabled: boolean;
  data: { timestamp: Date; value: number }[];
}

interface AnomalyDetection {
  timestamp: Date;
  value: number;
  threshold: number;
  severity: 'low' | 'medium' | 'high';
  metric: string;
}

@Component({
  selector: 'app-time-series-chart',
  templateUrl: './time-series-chart.component.html',
  styleUrls: ['./time-series-chart.component.scss']
})
export class TimeSeriesChartComponent implements OnInit, OnDestroy {
  @ViewChild('chartContainer', { static: true }) chartContainer!: ElementRef;
  @Input() selectedNodeId?: string;

  private destroy$ = new Subject<void>();
  private chart!: echarts.ECharts;
  private resizeObserver!: ResizeObserver;

  // Chart configuration
  chartType: 'line' | 'area' | 'scatter' = 'line';
  timeRange: '1h' | '6h' | '24h' | '7d' = '1h';
  refreshInterval: 5 | 10 | 30 | 60 = 5; // seconds
  showAnomalies = true;
  showPredictions = false;
  autoRefresh = true;

  // Data management
  metrics: ChartMetric[] = [
    {
      id: 'e2-latency',
      name: 'E2 Interface Latency',
      unit: 'ms',
      color: '#ff6b6b',
      enabled: true,
      data: []
    },
    {
      id: 'cpu-usage',
      name: 'CPU Usage',
      unit: '%',
      color: '#4ecdc4',
      enabled: true,
      data: []
    },
    {
      id: 'memory-usage',
      name: 'Memory Usage',
      unit: '%',
      color: '#45b7d1',
      enabled: true,
      data: []
    },
    {
      id: 'throughput',
      name: 'Network Throughput',
      unit: 'Mbps',
      color: '#96ceb4',
      enabled: false,
      data: []
    },
    {
      id: 'handover-success',
      name: 'Handover Success Rate',
      unit: '%',
      color: '#feca57',
      enabled: false,
      data: []
    },
    {
      id: 'resource-utilization',
      name: 'Resource Utilization',
      unit: '%',
      color: '#ff9ff3',
      enabled: false,
      data: []
    }
  ];

  anomalies: AnomalyDetection[] = [];
  selectedMetrics: string[] = ['e2-latency', 'cpu-usage', 'memory-usage'];
  
  // Statistics
  statistics = {
    totalDataPoints: 0,
    anomaliesDetected: 0,
    averageLatency: 0,
    maxThroughput: 0,
    lastUpdate: new Date()
  };

  constructor(
    private yangDataService: YangDataService,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit(): void {
    this.initializeChart();
    this.subscribeToData();
    this.setupResizeObserver();
    this.startAutoRefresh();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    
    if (this.chart) {
      this.chart.dispose();
    }
    
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
    }
  }

  /**
   * Initialize ECharts instance
   */
  private initializeChart(): void {
    const container = this.chartContainer.nativeElement;
    this.chart = echarts.init(container, 'dark', {
      renderer: 'canvas',
      useDirtyRect: false,
      width: 'auto',
      height: 'auto'
    });

    this.updateChartOptions();
    this.chart.on('click', this.onChartClick.bind(this));
    this.chart.on('datazoom', this.onDataZoom.bind(this));
  }

  /**
   * Update chart options and data
   */
  updateChartOptions(): void {
    const enabledMetrics = this.metrics.filter(m => m.enabled);
    
    const option: echarts.EChartsOption = {
      title: {
        text: 'O-RAN Near-RT RIC Metrics',
        subtext: 'Real-time performance monitoring',
        left: 'center',
        textStyle: {
          color: '#ffffff',
          fontSize: 18,
          fontWeight: 'bold'
        },
        subtextStyle: {
          color: '#cccccc',
          fontSize: 12
        }
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'cross',
          animation: true,
          crossStyle: {
            color: '#4ecdc4'
          }
        },
        backgroundColor: 'rgba(20, 20, 20, 0.9)',
        borderColor: '#4ecdc4',
        borderWidth: 1,
        textStyle: {
          color: '#ffffff',
          fontSize: 12
        },
        formatter: (params: any) => {
          if (!Array.isArray(params)) return '';
          
          let result = `<div style="font-weight: bold; margin-bottom: 5px;">
                          ${new Date(params[0].axisValue).toLocaleTimeString()}
                        </div>`;
          
          params.forEach((param: any) => {
            const metric = this.metrics.find(m => m.name === param.seriesName);
            const anomaly = this.findAnomalyAt(new Date(param.axisValue), param.seriesName);
            
            result += `<div style="margin: 2px 0;">
                        <span style="color: ${param.color};">●</span>
                        ${param.seriesName}: ${param.value}${metric?.unit || ''}
                        ${anomaly ? ' ⚠️' : ''}
                      </div>`;
          });
          
          return result;
        }
      },
      legend: {
        data: enabledMetrics.map(m => m.name),
        top: 40,
        textStyle: {
          color: '#ffffff'
        },
        selected: enabledMetrics.reduce((acc, m) => {
          acc[m.name] = this.selectedMetrics.includes(m.id);
          return acc;
        }, {} as Record<string, boolean>)
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        top: 80,
        containLabel: true
      },
      xAxis: {
        type: 'time',
        boundaryGap: false,
        axisLine: {
          lineStyle: { color: '#cccccc' }
        },
        axisLabel: {
          color: '#cccccc',
          formatter: (value: number) => {
            return new Date(value).toLocaleTimeString('en-US', {
              hour: '2-digit',
              minute: '2-digit',
              second: '2-digit'
            });
          }
        },
        splitLine: {
          show: true,
          lineStyle: {
            color: '#333333',
            type: 'dashed'
          }
        }
      },
      yAxis: [
        {
          type: 'value',
          name: 'Primary Metrics',
          position: 'left',
          axisLine: {
            lineStyle: { color: '#cccccc' }
          },
          axisLabel: {
            color: '#cccccc'
          },
          splitLine: {
            lineStyle: {
              color: '#333333',
              type: 'dashed'
            }
          }
        },
        {
          type: 'value',
          name: 'Secondary Metrics',
          position: 'right',
          axisLine: {
            lineStyle: { color: '#cccccc' }
          },
          axisLabel: {
            color: '#cccccc'
          },
          splitLine: { show: false }
        }
      ],
      dataZoom: [
        {
          type: 'inside',
          start: 70,
          end: 100,
          filterMode: 'none'
        },
        {
          type: 'slider',
          start: 70,
          end: 100,
          height: 30,
          bottom: 10,
          handleStyle: {
            color: '#4ecdc4'
          },
          textStyle: {
            color: '#ffffff'
          },
          borderColor: '#4ecdc4'
        }
      ],
      series: this.generateSeries(enabledMetrics),
      animation: true,
      animationDuration: 1000,
      animationEasing: 'cubicOut'
    };

    this.chart.setOption(option, true);
  }

  /**
   * Generate chart series data
   */
  private generateSeries(metrics: ChartMetric[]): echarts.SeriesOption[] {
    const series: echarts.SeriesOption[] = [];

    metrics.forEach((metric, index) => {
      if (!this.selectedMetrics.includes(metric.id)) return;

      const seriesData = metric.data.map(point => [
        point.timestamp.getTime(),
        point.value
      ]);

      // Main metric series
      if (this.chartType === 'scatter') {
        series.push({
          name: metric.name,
          type: 'scatter',
          data: seriesData,
          yAxisIndex: index > 2 ? 1 : 0,
          symbol: 'circle',
          symbolSize: 6,
          itemStyle: {
            color: metric.color
          },
          emphasis: {
            focus: 'series'
          },
          markPoint: this.getAnomalyPoints(metric)
        });
      } else {
        series.push({
          name: metric.name,
          type: 'line',
          data: seriesData,
          yAxisIndex: index > 2 ? 1 : 0,
          smooth: true,
          symbol: 'none',
          symbolSize: 4,
          lineStyle: {
            color: metric.color,
            width: 2
          },
          itemStyle: {
            color: metric.color
          },
          areaStyle: this.chartType === 'area' ? {
            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
              { offset: 0, color: metric.color + '80' },
              { offset: 1, color: metric.color + '10' }
            ])
          } : undefined,
          emphasis: {
            focus: 'series',
            lineStyle: {
              width: 3
            }
          },
          markLine: this.getThresholdLines(metric),
          markPoint: this.getAnomalyPoints(metric)
        });
      }

      // Add prediction series if enabled
      if (this.showPredictions) {
        const predictions = this.generatePredictions(metric);
        if (predictions.length > 0) {
          series.push({
            name: `${metric.name} (Predicted)`,
            type: 'line',
            data: predictions,
            yAxisIndex: index > 2 ? 1 : 0,
            lineStyle: {
              color: metric.color,
              type: 'dashed',
              width: 1,
              opacity: 0.6
            },
            symbol: 'none',
            silent: true
          });
        }
      }
    });

    return series;
  }

  /**
   * Get threshold lines for metrics
   */
  private getThresholdLines(metric: ChartMetric): echarts.MarkLineComponentOption {
    const thresholds: Record<string, number> = {
      'e2-latency': 100, // 100ms threshold
      'cpu-usage': 80,   // 80% threshold
      'memory-usage': 85, // 85% threshold
      'throughput': 500,  // 500 Mbps threshold
    };

    const threshold = thresholds[metric.id];
    if (!threshold) return {};

    return {
      silent: true,
      lineStyle: {
        color: '#ff6b6b',
        type: 'dashed',
        width: 1,
        opacity: 0.7
      },
      label: {
        show: true,
        position: 'end',
        formatter: `Threshold: ${threshold}${metric.unit}`
      },
      data: [
        {
          yAxis: threshold,
          name: 'Threshold'
        }
      ]
    };
  }

  /**
   * Get anomaly points for metrics
   */
  private getAnomalyPoints(metric: ChartMetric): echarts.MarkPointComponentOption {
    const anomalies = this.anomalies.filter(a => a.metric === metric.name);
    if (!this.showAnomalies || anomalies.length === 0) return {};

    return {
      symbol: 'triangle',
      symbolSize: 8,
      itemStyle: {
        color: '#ff6b6b',
        borderColor: '#ffffff',
        borderWidth: 1
      },
      label: {
        show: true,
        position: 'top',
        formatter: '⚠️'
      },
      data: anomalies.map(anomaly => ({
        name: `Anomaly: ${anomaly.severity}`,
        coord: [anomaly.timestamp.getTime(), anomaly.value],
        value: `Anomaly: ${anomaly.severity}`
      }))
    };
  }

  /**
   * Generate prediction data points
   */
  private generatePredictions(metric: ChartMetric): Array<[number, number]> {
    if (metric.data.length < 10) return [];

    const lastPoints = metric.data.slice(-10);
    const predictions: Array<[number, number]> = [];
    
    // Simple linear regression for prediction
    const trend = this.calculateTrend(lastPoints);
    const lastPoint = lastPoints[lastPoints.length - 1];
    
    for (let i = 1; i <= 10; i++) {
      const futureTime = new Date(lastPoint.timestamp.getTime() + (i * 5000));
      const predictedValue = lastPoint.value + (trend * i);
      predictions.push([futureTime.getTime(), predictedValue]);
    }

    return predictions;
  }

  /**
   * Calculate trend from data points
   */
  private calculateTrend(points: { timestamp: Date; value: number }[]): number {
    if (points.length < 2) return 0;

    const n = points.length;
    let sumX = 0, sumY = 0, sumXY = 0, sumXX = 0;

    points.forEach((point, index) => {
      sumX += index;
      sumY += point.value;
      sumXY += index * point.value;
      sumXX += index * index;
    });

    return (n * sumXY - sumX * sumY) / (n * sumXX - sumX * sumX);
  }

  /**
   * Subscribe to real-time data updates
   */
  private subscribeToData(): void {
    // Subscribe to time series data
    this.yangDataService.timeSeriesData$
      .pipe(takeUntil(this.destroy$))
      .subscribe(data => {
        this.updateMetricsData(data);
        this.detectAnomalies(data);
        this.updateStatistics();
        this.updateChartOptions();
        this.cdr.detectChanges();
      });

    // Subscribe to xApp metrics
    this.yangDataService.xAppMetrics$
      .pipe(takeUntil(this.destroy$))
      .subscribe(metrics => {
        this.updateXAppMetrics(metrics);
      });
  }

  /**
   * Update metrics data with new time series points
   */
  private updateMetricsData(data: TimeSeriesData[]): void {
    data.forEach(point => {
      const metric = this.metrics.find(m => m.id === point.nodeId);
      if (metric) {
        metric.data.push({
          timestamp: point.timestamp,
          value: point.value
        });

        // Keep only last 1000 points for performance
        if (metric.data.length > 1000) {
          metric.data = metric.data.slice(-1000);
        }
      }
    });
  }

  /**
   * Update metrics from xApp data
   */
  private updateXAppMetrics(xAppMetrics: XAppMetrics[]): void {
    xAppMetrics.forEach(xApp => {
      // Update CPU usage
      const cpuMetric = this.metrics.find(m => m.id === 'cpu-usage');
      if (cpuMetric) {
        cpuMetric.data.push({
          timestamp: xApp.timestamp,
          value: xApp.cpu
        });
      }

      // Update memory usage
      const memoryMetric = this.metrics.find(m => m.id === 'memory-usage');
      if (memoryMetric) {
        memoryMetric.data.push({
          timestamp: xApp.timestamp,
          value: xApp.memory
        });
      }

      // Update RRM metrics
      const throughputMetric = this.metrics.find(m => m.id === 'throughput');
      if (throughputMetric) {
        throughputMetric.data.push({
          timestamp: xApp.timestamp,
          value: xApp.rrmMetrics.throughput
        });
      }

      const handoverMetric = this.metrics.find(m => m.id === 'handover-success');
      if (handoverMetric) {
        handoverMetric.data.push({
          timestamp: xApp.timestamp,
          value: xApp.rrmMetrics.handoverSuccess
        });
      }

      const resourceMetric = this.metrics.find(m => m.id === 'resource-utilization');
      if (resourceMetric) {
        resourceMetric.data.push({
          timestamp: xApp.timestamp,
          value: xApp.rrmMetrics.resourceUtilization
        });
      }
    });
  }

  /**
   * Detect anomalies in the data
   */
  private detectAnomalies(data: TimeSeriesData[]): void {
    const thresholds: Record<string, { high: number; critical: number }> = {
      'e2-latency': { high: 50, critical: 100 },
      'cpu-usage': { high: 70, critical: 90 },
      'memory-usage': { high: 75, critical: 95 },
      'throughput': { high: 800, critical: 1000 }
    };

    data.forEach(point => {
      const threshold = thresholds[point.nodeId];
      if (!threshold) return;

      let severity: 'low' | 'medium' | 'high' = 'low';
      let isAnomaly = false;

      if (point.value > threshold.critical) {
        severity = 'high';
        isAnomaly = true;
      } else if (point.value > threshold.high) {
        severity = 'medium';
        isAnomaly = true;
      }

      if (isAnomaly) {
        const metric = this.metrics.find(m => m.id === point.nodeId);
        this.anomalies.push({
          timestamp: point.timestamp,
          value: point.value,
          threshold: severity === 'high' ? threshold.critical : threshold.high,
          severity,
          metric: metric?.name || point.nodeId
        });

        // Keep only last 100 anomalies
        if (this.anomalies.length > 100) {
          this.anomalies = this.anomalies.slice(-100);
        }
      }
    });
  }

  /**
   * Update statistics
   */
  private updateStatistics(): void {
    this.statistics.totalDataPoints = this.metrics.reduce(
      (sum, metric) => sum + metric.data.length, 0
    );
    
    this.statistics.anomaliesDetected = this.anomalies.length;
    
    const latencyMetric = this.metrics.find(m => m.id === 'e2-latency');
    if (latencyMetric && latencyMetric.data.length > 0) {
      const sum = latencyMetric.data.reduce((s, d) => s + d.value, 0);
      this.statistics.averageLatency = sum / latencyMetric.data.length;
    }

    const throughputMetric = this.metrics.find(m => m.id === 'throughput');
    if (throughputMetric && throughputMetric.data.length > 0) {
      this.statistics.maxThroughput = Math.max(
        ...throughputMetric.data.map(d => d.value)
      );
    }

    this.statistics.lastUpdate = new Date();
  }

  /**
   * Setup resize observer for responsive charts
   */
  private setupResizeObserver(): void {
    if (!window.ResizeObserver) return;

    this.resizeObserver = new ResizeObserver(() => {
      if (this.chart) {
        this.chart.resize();
      }
    });

    this.resizeObserver.observe(this.chartContainer.nativeElement);
  }

  /**
   * Start auto-refresh timer
   */
  private startAutoRefresh(): void {
    if (!this.autoRefresh) return;

    setInterval(() => {
      if (this.autoRefresh && this.chart) {
        this.updateChartOptions();
      }
    }, this.refreshInterval * 1000);
  }

  /**
   * Handle chart click events
   */
  private onChartClick(params: any): void {
    console.log('Chart clicked:', params);
    
    if (params.componentType === 'series') {
      const timestamp = new Date(params.value[0]);
      const metric = this.metrics.find(m => m.name === params.seriesName);
      
      // Emit event or show details
      console.log('Clicked data point:', {
        metric: metric?.name,
        timestamp,
        value: params.value[1]
      });
    }
  }

  /**
   * Handle data zoom events
   */
  private onDataZoom(params: any): void {
    console.log('Data zoom:', params);
  }

  /**
   * Find anomaly at specific timestamp
   */
  private findAnomalyAt(timestamp: Date, metricName: string): AnomalyDetection | undefined {
    return this.anomalies.find(a => 
      Math.abs(a.timestamp.getTime() - timestamp.getTime()) < 5000 &&
      a.metric === metricName
    );
  }

  // Public methods for component interaction

  /**
   * Toggle metric visibility
   */
  toggleMetric(metricId: string): void {
    const index = this.selectedMetrics.indexOf(metricId);
    if (index > -1) {
      this.selectedMetrics.splice(index, 1);
    } else {
      this.selectedMetrics.push(metricId);
    }
    this.updateChartOptions();
  }

  /**
   * Change chart type
   */
  setChartType(type: 'line' | 'area' | 'scatter'): void {
    this.chartType = type;
    this.updateChartOptions();
  }

  /**
   * Change time range
   */
  setTimeRange(range: '1h' | '6h' | '24h' | '7d'): void {
    this.timeRange = range;
    // Filter data based on time range
    const now = new Date();
    const hours = range === '1h' ? 1 : range === '6h' ? 6 : range === '24h' ? 24 : 168;
    const cutoff = new Date(now.getTime() - (hours * 60 * 60 * 1000));

    this.metrics.forEach(metric => {
      metric.data = metric.data.filter(d => d.timestamp >= cutoff);
    });

    this.updateChartOptions();
  }

  /**
   * Toggle auto refresh
   */
  toggleAutoRefresh(): void {
    this.autoRefresh = !this.autoRefresh;
    if (this.autoRefresh) {
      this.startAutoRefresh();
    }
  }

  /**
   * Export chart as image
   */
  exportChart(): void {
    if (this.chart) {
      const dataURL = this.chart.getDataURL({
        type: 'png',
        backgroundColor: '#1a1a1a',
        excludeComponents: ['toolbox']
      });
      
      const link = document.createElement('a');
      link.download = `oran-metrics-${new Date().toISOString().slice(0, 19)}.png`;
      link.href = dataURL;
      link.click();
    }
  }

  /**
   * Clear all data
   */
  clearData(): void {
    this.metrics.forEach(metric => {
      metric.data = [];
    });
    this.anomalies = [];
    this.updateChartOptions();
    this.updateStatistics();
  }
}