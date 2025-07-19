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

import {Component, computed, effect, inject, signal} from '@angular/core';
import {CommonModule} from '@angular/common';
import {MatButtonModule} from '@angular/material/button';
import {MatMenuModule} from '@angular/material/menu';
import {MatIconModule} from '@angular/material/icon';
import {MatDividerModule} from '@angular/material/divider';
import {MatTooltipModule} from '@angular/material/tooltip';
import {Router} from '@angular/router';
import {takeUntilDestroyed} from '@angular/core/rxjs-interop';

import {AuthService} from '@common/services/global/authentication';
import {AssetsService} from '@common/services/global/assets';
import {KdStateService} from '@common/services/global/state';
import {LocalConfigLoaderService} from '@common/services/global/loader';
import {MESSAGES_DI_TOKEN} from '../../index.messages';
import {IMessage} from '@api/root.ui';
import {LoginStatus} from '@api/root.ui';

// Modern Angular 18 standalone component with signals and new patterns
@Component({
  selector: 'kd-user-panel',
  templateUrl: './modern-template.html',
  styleUrls: ['./style.scss'],
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule,
    MatMenuModule,
    MatIconModule,
    MatDividerModule,
    MatTooltipModule,
  ],
})
export class ModernUserPanelComponent {
  // Injected services using modern inject function
  private authService = inject(AuthService);
  private assetsService = inject(AssetsService);
  private stateService = inject(KdStateService);
  private configLoader = inject(LocalConfigLoaderService);
  private router = inject(Router);
  private messages = inject(MESSAGES_DI_TOKEN);

  // Reactive state using signals
  loginStatus = signal<LoginStatus | null>(null);
  isLoginStatusInitialized = signal(false);
  
  // Computed properties
  hasUsername = computed(() => {
    const status = this.loginStatus();
    return !!(status?.username && status.username.length > 0);
  });

  username = computed(() => this.loginStatus()?.username || '');
  
  userIconSrc = computed(() => 
    this.assetsService.getAssetPath('images/kubernetes-logo.svg')
  );

  // Effects for reactive updates
  constructor() {
    // Initialize login status
    effect(() => {
      this.authService
        .getLoginStatus()
        .pipe(takeUntilDestroyed())
        .subscribe(status => {
          this.loginStatus.set(status);
          this.isLoginStatusInitialized.set(true);
        });
    });

    // React to login status changes
    effect(() => {
      const status = this.loginStatus();
      if (status) {
        console.log('Login status updated:', status);
      }
    });
  }

  isLoggedIn(): boolean {
    return !!this.loginStatus() && !this.isAuthSkipped();
  }

  isAuthSkipped(): boolean {
    return this.authService.isAuthenticationSkipped(this.loginStatus());
  }

  logout(): void {
    this.authService.logout();
  }

  getSettingsRoute(): string {
    return this.stateService.href('settings');
  }

  getAboutRoute(): string {
    return this.stateService.href('about');
  }

  navigateToSettings(): void {
    this.router.navigate([this.getSettingsRoute()]);
  }

  navigateToAbout(): void {
    this.router.navigate([this.getAboutRoute()]);
  }

  getLoginUrl(): string {
    return this.configLoader.getLoginUrl();
  }

  // Getter for template access
  get message(): IMessage {
    return this.messages;
  }
}