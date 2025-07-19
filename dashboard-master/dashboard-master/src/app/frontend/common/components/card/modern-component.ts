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

import {Component, computed, inject, input, signal} from '@angular/core';
import {CommonModule} from '@angular/common';
import {MatCardModule} from '@angular/material/card';
import {MatIconModule} from '@angular/material/icon';
import {MatButtonModule} from '@angular/material/button';
import {MatProgressSpinnerModule} from '@angular/material/progress-spinner';
import {MatTooltipModule} from '@angular/material/tooltip';
import {IMessage} from '@api/root.ui';
import {MESSAGES_DI_TOKEN} from '../../../index.messages';
import {Animations} from '../../animations/animations';

// Modern Angular 18 standalone component with signals
@Component({
  selector: 'kd-card',
  templateUrl: './modern-template.html',
  styleUrls: ['./style.scss'],
  animations: [Animations.expandInOut],
  standalone: true,
  imports: [
    CommonModule,
    MatCardModule,
    MatIconModule,
    MatButtonModule,
    MatProgressSpinnerModule,
    MatTooltipModule,
  ],
})
export class ModernCardComponent {
  // Input signals (Angular 17.1+)
  initialized = input(true);
  role = input<'inner' | 'table' | 'inner-content'>();
  withFooter = input(false);
  withTitle = input(true);
  expandable = input(true);
  expanded = input(true);
  graphMode = input(false);
  titleClasses = input('');

  // Internal signals
  private expandedState = signal(this.expanded());
  private classes = signal<string[]>([]);

  // Injected services using inject function
  private message = inject(MESSAGES_DI_TOKEN);

  // Computed signals
  isExpanded = computed(() => this.expandedState());
  
  titleClassMap = computed(() => {
    const ngCls: {[clsName: string]: boolean} = {};
    
    if (!this.expandedState()) {
      ngCls['kd-minimized-card-header'] = true;
    }

    if (this.expandable()) {
      ngCls['kd-card-header'] = true;
    }

    for (const cls of this.classes()) {
      ngCls[cls] = true;
    }

    return ngCls;
  });

  cardClasses = computed(() => ({
    'kd-minimized-card': !this.expandedState() && !this.graphMode(),
    'kd-graph': this.graphMode(),
    'kd-inner-table': this.role() === 'inner',
    'kd-inner-content': this.role() === 'inner-content'
  }));

  constructor() {
    // Initialize expanded state from input
    this.expandedState.set(this.expanded());
    
    // Parse title classes
    this.classes.set(this.titleClasses().split(/\s+/).filter(cls => cls.trim()));
  }

  expand(): void {
    if (this.expandable()) {
      this.expandedState.update(expanded => !expanded);
    }
  }

  onCardHeaderClick(): void {
    if (this.expandable() && !this.expandedState()) {
      this.expandedState.set(true);
    }
  }

  // Getter for template compatibility
  get messages(): IMessage {
    return this.message;
  }
}