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
  Output, 
  EventEmitter,
  ChangeDetectorRef 
} from '@angular/core';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import * as d3 from 'd3';

import { YangDataService, YangNode } from '../../services/yang-data.service';

interface D3Node extends d3.HierarchyNode<YangNode> {
  x: number;
  y: number;
  children?: D3Node[];
  _children?: D3Node[];
  id: string;
}

@Component({
  selector: 'app-yang-tree-browser',
  templateUrl: './yang-tree-browser.component.html',
  styleUrls: ['./yang-tree-browser.component.scss']
})
export class YangTreeBrowserComponent implements OnInit, OnDestroy {
  @ViewChild('treeContainer', { static: true }) treeContainer!: ElementRef;
  @Output() nodeSelected = new EventEmitter<YangNode>();

  private destroy$ = new Subject<void>();
  private svg!: d3.Selection<SVGSVGElement, unknown, null, undefined>;
  private g!: d3.Selection<SVGGElement, unknown, null, undefined>;
  private tree!: d3.TreeLayout<YangNode>;
  private root!: D3Node;
  
  private width = 1200;
  private height = 800;
  private margin = { top: 40, right: 90, bottom: 50, left: 90 };
  private duration = 750;
  private nodeCounter = 0;

  // Search and filter functionality
  searchTerm = '';
  filterType = 'all';
  showMetadata = true;
  expandedNodes = new Set<string>();

  // Node type colors
  private nodeColors = {
    'container': '#4CAF50',
    'leaf': '#2196F3', 
    'leaf-list': '#FF9800',
    'list': '#9C27B0',
    'choice': '#F44336',
    'case': '#607D8B'
  };

  constructor(
    private yangDataService: YangDataService,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit(): void {
    this.initializeTree();
    this.loadData();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  /**
   * Initialize D3 tree structure
   */
  private initializeTree(): void {
    const container = this.treeContainer.nativeElement;
    const rect = container.getBoundingClientRect();
    this.width = rect.width || 1200;
    this.height = rect.height || 800;

    // Clear any existing SVG
    d3.select(container).selectAll('*').remove();

    this.svg = d3.select(container)
      .append('svg')
      .attr('width', this.width)
      .attr('height', this.height)
      .attr('class', 'yang-tree-svg');

    // Add zoom behavior
    const zoom = d3.zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.1, 3])
      .on('zoom', (event) => {
        this.g.attr('transform', event.transform);
      });

    this.svg.call(zoom);

    // Add main group
    this.g = this.svg.append('g')
      .attr('transform', `translate(${this.margin.left},${this.margin.top})`);

    // Add gradient definitions
    this.addGradientDefinitions();

    // Add legend
    this.addLegend();

    // Initialize tree layout
    this.tree = d3.tree<YangNode>()
      .size([this.height - this.margin.top - this.margin.bottom, 
             this.width - this.margin.left - this.margin.right]);
  }

  /**
   * Add gradient definitions for visual enhancement
   */
  private addGradientDefinitions(): void {
    const defs = this.svg.append('defs');

    // Node gradient
    const nodeGradient = defs.append('linearGradient')
      .attr('id', 'nodeGradient')
      .attr('x1', '0%').attr('y1', '0%')
      .attr('x2', '100%').attr('y2', '100%');

    nodeGradient.append('stop')
      .attr('offset', '0%')
      .attr('stop-color', '#ffffff')
      .attr('stop-opacity', 0.8);

    nodeGradient.append('stop')
      .attr('offset', '100%')
      .attr('stop-color', '#f0f0f0')
      .attr('stop-opacity', 0.3);

    // Hover effect filter
    const filter = defs.append('filter')
      .attr('id', 'glow');

    filter.append('feGaussianBlur')
      .attr('stdDeviation', '3')
      .attr('result', 'coloredBlur');

    const merge = filter.append('feMerge');
    merge.append('feMergeNode').attr('in', 'coloredBlur');
    merge.append('feMergeNode').attr('in', 'SourceGraphic');
  }

  /**
   * Add legend for node types
   */
  private addLegend(): void {
    const legend = this.svg.append('g')
      .attr('class', 'legend')
      .attr('transform', `translate(${this.width - 200}, 20)`);

    const legendItems = Object.entries(this.nodeColors);
    
    legendItems.forEach(([type, color], i) => {
      const legendItem = legend.append('g')
        .attr('class', 'legend-item')
        .attr('transform', `translate(0, ${i * 25})`);

      legendItem.append('circle')
        .attr('r', 8)
        .attr('fill', color)
        .attr('stroke', '#333')
        .attr('stroke-width', 1);

      legendItem.append('text')
        .attr('x', 15)
        .attr('y', 5)
        .style('font-size', '12px')
        .style('font-family', 'monospace')
        .text(type);
    });
  }

  /**
   * Load YANG data and initialize tree
   */
  private loadData(): void {
    this.yangDataService.loadYangData()
      .pipe(takeUntil(this.destroy$))
      .subscribe(data => {
        if (data && data.length > 0) {
          this.processData(data[0]); // Use first root node
          this.updateTree();
        }
      });
  }

  /**
   * Process YANG data for D3 tree
   */
  private processData(rootNode: YangNode): void {
    const hierarchy = d3.hierarchy(rootNode, (d: YangNode) => d.children);
    
    this.root = this.tree(hierarchy) as D3Node;
    this.root.x0 = this.height / 2;
    this.root.y0 = 0;

    // Collapse all nodes initially except root
    if (this.root.children) {
      this.root.children.forEach(d => this.collapse(d));
    }

    // Assign unique IDs
    this.assignNodeIds(this.root);
  }

  /**
   * Assign unique IDs to nodes
   */
  private assignNodeIds(node: D3Node): void {
    node.id = node.data.id || `node-${this.nodeCounter++}`;
    if (node.children) {
      node.children.forEach(child => this.assignNodeIds(child));
    }
    if (node._children) {
      node._children.forEach(child => this.assignNodeIds(child));
    }
  }

  /**
   * Collapse node and its children
   */
  private collapse(d: D3Node): void {
    if (d.children) {
      d._children = d.children;
      d._children.forEach(child => this.collapse(child));
      d.children = undefined;
    }
  }

  /**
   * Update tree visualization
   */
  private updateTree(source?: D3Node): void {
    if (!this.root) return;

    const treeData = this.tree(this.root);
    const nodes = treeData.descendants();
    const links = treeData.descendants().slice(1);

    // Normalize for fixed-depth
    nodes.forEach(d => d.y = d.depth * 180);

    // Update nodes
    this.updateNodes(nodes, source || this.root);

    // Update links
    this.updateLinks(links, source || this.root);
  }

  /**
   * Update node elements
   */
  private updateNodes(nodes: D3Node[], source: D3Node): void {
    const node = this.g.selectAll('g.node')
      .data(nodes, (d: any) => d.id);

    // Enter new nodes
    const nodeEnter = node.enter().append('g')
      .attr('class', 'node')
      .attr('transform', () => `translate(${source.y0},${source.x0})`)
      .style('cursor', 'pointer')
      .on('click', (event, d) => this.toggleNode(d));

    // Add circles for nodes
    nodeEnter.append('circle')
      .attr('r', 1e-6)
      .style('fill', d => d._children ? 'lightsteelblue' : '#fff')
      .style('stroke', d => this.nodeColors[d.data.type as keyof typeof this.nodeColors] || '#999')
      .style('stroke-width', '2px')
      .style('filter', 'url(#nodeGradient)');

    // Add labels
    nodeEnter.append('text')
      .attr('dy', '.35em')
      .attr('x', d => d.children || d._children ? -13 : 13)
      .attr('text-anchor', d => d.children || d._children ? 'end' : 'start')
      .style('font-family', 'monospace')
      .style('font-size', '12px')
      .style('fill-opacity', 1e-6)
      .text(d => d.data.name);

    // Add value display for leaf nodes
    nodeEnter.filter(d => d.data.type === 'leaf' && d.data.value !== undefined)
      .append('text')
      .attr('dy', '1.5em')
      .attr('x', 13)
      .attr('text-anchor', 'start')
      .style('font-family', 'monospace')
      .style('font-size', '10px')
      .style('fill', '#666')
      .style('fill-opacity', 1e-6)
      .text(d => `= ${d.data.value}`);

    // Add metadata tooltip
    nodeEnter.filter(d => d.data.metadata?.description)
      .append('title')
      .text(d => `${d.data.name}\nType: ${d.data.type}\nPath: ${d.data.path}\n${d.data.metadata?.description || ''}`);

    // Transition nodes to their new position
    const nodeUpdate = nodeEnter.merge(node as any);

    nodeUpdate.transition()
      .duration(this.duration)
      .attr('transform', d => `translate(${d.y},${d.x})`);

    nodeUpdate.select('circle')
      .attr('r', d => this.getNodeRadius(d))
      .style('fill', d => this.getNodeColor(d))
      .style('stroke', d => this.nodeColors[d.data.type as keyof typeof this.nodeColors] || '#999')
      .style('stroke-width', d => d.data.metrics?.errorCount ? '3px' : '2px');

    nodeUpdate.select('text')
      .style('fill-opacity', 1)
      .style('font-weight', d => d.data.type === 'container' ? 'bold' : 'normal');

    nodeUpdate.selectAll('text:nth-child(3)')
      .style('fill-opacity', 1);

    // Remove exiting nodes
    const nodeExit = node.exit().transition()
      .duration(this.duration)
      .attr('transform', () => `translate(${source.y},${source.x})`)
      .remove();

    nodeExit.select('circle')
      .attr('r', 1e-6);

    nodeExit.select('text')
      .style('fill-opacity', 1e-6);
  }

  /**
   * Update link elements
   */
  private updateLinks(links: D3Node[], source: D3Node): void {
    const link = this.g.selectAll('path.link')
      .data(links, (d: any) => d.id);

    // Enter new links
    const linkEnter = link.enter().insert('path', 'g')
      .attr('class', 'link')
      .attr('d', () => {
        const o = { x: source.x0, y: source.y0 };
        return this.diagonal(o, o);
      })
      .style('fill', 'none')
      .style('stroke', '#ccc')
      .style('stroke-width', '2px');

    // Transition links to their new position
    const linkUpdate = linkEnter.merge(link as any);

    linkUpdate.transition()
      .duration(this.duration)
      .attr('d', d => this.diagonal(d, d.parent!))
      .style('stroke', d => d.data.metrics?.errorCount ? '#f44336' : '#ccc')
      .style('stroke-width', d => d.data.metrics?.errorCount ? '3px' : '2px');

    // Remove exiting links
    link.exit().transition()
      .duration(this.duration)
      .attr('d', () => {
        const o = { x: source.x, y: source.y };
        return this.diagonal(o, o);
      })
      .remove();
  }

  /**
   * Generate diagonal path for links
   */
  private diagonal(s: { x: number; y: number }, d: { x: number; y: number }): string {
    return `M ${s.y} ${s.x}
            C ${(s.y + d.y) / 2} ${s.x},
              ${(s.y + d.y) / 2} ${d.x},
              ${d.y} ${d.x}`;
  }

  /**
   * Toggle node expansion/collapse
   */
  private toggleNode(d: D3Node): void {
    if (d.children) {
      d._children = d.children;
      d.children = undefined;
      this.expandedNodes.delete(d.id);
    } else {
      d.children = d._children;
      d._children = undefined;
      this.expandedNodes.add(d.id);
    }

    // Emit selection event
    this.nodeSelected.emit(d.data);

    this.updateTree(d);
  }

  /**
   * Get node radius based on type and data
   */
  private getNodeRadius(d: D3Node): number {
    const baseRadius = 6;
    const multiplier = d.data.type === 'container' ? 1.5 : 
                     d.data.type === 'list' ? 1.3 : 1.0;
    
    return baseRadius * multiplier;
  }

  /**
   * Get node color based on status and type
   */
  private getNodeColor(d: D3Node): string {
    if (d.data.metrics?.errorCount) {
      return '#ffcdd2'; // Light red for errors
    }
    
    if (d._children) {
      return 'lightsteelblue'; // Collapsed nodes
    }
    
    return '#fff'; // Default
  }

  /**
   * Search functionality
   */
  onSearch(): void {
    if (!this.searchTerm.trim()) {
      this.clearHighlights();
      return;
    }

    this.highlightSearchResults(this.searchTerm.toLowerCase());
  }

  /**
   * Highlight search results
   */
  private highlightSearchResults(term: string): void {
    this.g.selectAll('g.node')
      .style('opacity', (d: any) => {
        const matches = d.data.name.toLowerCase().includes(term) ||
                       d.data.path.toLowerCase().includes(term) ||
                       (d.data.metadata?.description || '').toLowerCase().includes(term);
        return matches ? 1 : 0.3;
      });

    this.g.selectAll('g.node circle')
      .style('stroke-width', (d: any) => {
        const matches = d.data.name.toLowerCase().includes(term) ||
                       d.data.path.toLowerCase().includes(term) ||
                       (d.data.metadata?.description || '').toLowerCase().includes(term);
        return matches ? '4px' : '2px';
      })
      .style('filter', (d: any) => {
        const matches = d.data.name.toLowerCase().includes(term) ||
                       d.data.path.toLowerCase().includes(term) ||
                       (d.data.metadata?.description || '').toLowerCase().includes(term);
        return matches ? 'url(#glow)' : 'none';
      });
  }

  /**
   * Clear search highlights
   */
  private clearHighlights(): void {
    this.g.selectAll('g.node')
      .style('opacity', 1);

    this.g.selectAll('g.node circle')
      .style('stroke-width', '2px')
      .style('filter', 'none');
  }

  /**
   * Filter by node type
   */
  onFilterChange(): void {
    if (this.filterType === 'all') {
      this.g.selectAll('g.node')
        .style('display', 'block');
    } else {
      this.g.selectAll('g.node')
        .style('display', (d: any) => 
          d.data.type === this.filterType ? 'block' : 'none'
        );
    }
  }

  /**
   * Expand all nodes
   */
  expandAll(): void {
    this.expandAllNodes(this.root);
    this.updateTree();
  }

  /**
   * Collapse all nodes
   */
  collapseAll(): void {
    if (this.root.children) {
      this.root.children.forEach(d => this.collapse(d));
    }
    this.updateTree();
  }

  /**
   * Recursively expand all nodes
   */
  private expandAllNodes(d: D3Node): void {
    if (d._children) {
      d.children = d._children;
      d._children = undefined;
      this.expandedNodes.add(d.id);
    }
    if (d.children) {
      d.children.forEach(child => this.expandAllNodes(child));
    }
  }

  /**
   * Center tree on root
   */
  centerTree(): void {
    const transform = d3.zoomIdentity
      .translate(this.width / 2, this.height / 2)
      .scale(1);
    
    this.svg.transition()
      .duration(750)
      .call(d3.zoom<SVGSVGElement, unknown>().transform as any, transform);
  }

  /**
   * Get node type options for filter
   */
  getNodeTypes(): string[] {
    return ['all', ...Object.keys(this.nodeColors)];
  }
}