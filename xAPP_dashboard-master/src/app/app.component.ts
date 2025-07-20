import { Component, AfterViewInit } from '@angular/core';
import { MonitoringService } from './monitoring.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements AfterViewInit {
  title = 'cur';

  constructor(private monitoringService: MonitoringService) {}

  ngAfterViewInit() {
    const startTime = performance.now();
    // Simulate some rendering time
    setTimeout(() => {
      const endTime = performance.now();
      const renderTime = endTime - startTime;
      this.monitoringService.trackPerformance('app_component_render_time', renderTime);
    }, 100);
  }
}
