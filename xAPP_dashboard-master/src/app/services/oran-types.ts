export interface ChartMetric {
  id: string;
  name: string;
  unit: string;
  color: string;
  enabled: boolean;
  data: { timestamp: Date; value: number }[];
}

export interface AnomalyDetection {
  timestamp: Date;
  value: number;
  threshold: number;
  severity: 'low' | 'medium' | 'high';
  metric: string;
}

export interface TooltipFormatterParams {
  axisValue: string | number;
  seriesName: string;
  color: string;
  value: number;
}

export interface ChartClickEventParams {
  componentType: string;
  seriesName: string;
  value: (string | number)[];
}

export interface DataZoomEventParams {
  start: number;
  end: number;
}

export interface EchartsOption {
  title: echarts.TitleComponentOption;
  tooltip: echarts.TooltipComponentOption;
  legend: echarts.LegendComponentOption;
  grid: echarts.GridComponentOption;
  xAxis: echarts.XAXisComponentOption;
  yAxis: echarts.YAXisComponentOption | echarts.YAXisComponentOption[];
  dataZoom: echarts.DataZoomComponentOption[];
  series: echarts.SeriesOption[];
  animation: boolean;
  animationDuration: number;
  animationEasing: string;
}