// xAPP_dashboard-master/src/app/services/xapp-service.ts
import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class XAppService {
  private baseUrl = '/api/xapps';

  constructor(private http: HttpClient) { }

  getSubscriptions(): Observable<any> {
    return this.http.get(`${this.baseUrl}/subscriptions`);
  }

  sendControlMessage(message: string): Observable<any> {
    return this.http.post(`${this.baseUrl}/control`, { message });
  }
}
