import { NgModule } from '@angular/core';
import { ReactiveFormsModule, FormsModule } from '@angular/forms';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { HomeComponent } from './home/home.component';
import { TagsComponent } from './tags/tags.component';
import { HttpClientModule } from '@angular/common/http';
import { ImagehistoryComponent } from './imagehistory/imagehistory.component';
import { FrontPageComponent } from './front-page/front-page.component';

// New visualization components
import { YangTreeBrowserComponent } from './components/yang-tree-browser/yang-tree-browser.component';
import { TimeSeriesChartComponent } from './components/time-series-chart/time-series-chart.component';

// Services
import { YangDataService } from './services/yang-data.service';

@NgModule({
  declarations: [
    AppComponent,
    HomeComponent,
    TagsComponent,
    ImagehistoryComponent,
    FrontPageComponent,
    YangTreeBrowserComponent,
    TimeSeriesChartComponent
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    AppRoutingModule,
    ReactiveFormsModule,
    FormsModule,
    HttpClientModule
  ],
  providers: [
    YangDataService
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
