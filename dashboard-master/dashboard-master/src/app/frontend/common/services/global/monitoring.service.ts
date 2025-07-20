import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class MonitoringService {

  constructor(private http: HttpClient) { }

  trackPerformance(metricName: string, value: number) {
    this.http.post('/api/metrics', { metricName, value }).subscribe();
  }
}
