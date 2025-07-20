// Copyright 2024 The O-RAN Near-RT RIC Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

import { Component, OnInit } from '@angular/core';
import { YangNode } from '../services/yang-data.service';

@Component({
  selector: 'app-visualization-dashboard',
  templateUrl: './visualization-dashboard.component.html',
  styleUrls: ['./visualization-dashboard.component.scss']
})
export class VisualizationDashboardComponent implements OnInit {
  selectedNode: YangNode | null = null;
  selectedNodeId: string | undefined;

  constructor() { }

  ngOnInit(): void {
    // Initialize visualization dashboard
    console.log('Visualization Dashboard initialized');
  }

  onNodeSelected(node: YangNode): void {
    this.selectedNode = node;
    this.selectedNodeId = node.id;
    console.log('Selected YANG node:', node);
  }
}