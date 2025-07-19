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

// Modern Angular 18 main.ts with standalone bootstrap
import {bootstrapApplication} from '@angular/platform-browser';
import {provideAnimations} from '@angular/platform-browser/animations';
import {provideHttpClient, withInterceptors} from '@angular/common/http';
import {provideRouter, withHashLocation, withRouterConfig} from '@angular/router';
import {ErrorHandler, importProvidersFrom, LOCALE_ID} from '@angular/core';
import {MatDialogModule} from '@angular/material/dialog';
import {MatSnackBarModule} from '@angular/material/snack-bar';

// Import the modern root component
import {ModernRootComponent} from './modern-root.component';
import {routes} from './index.routing';

// Import services and interceptors
import {GlobalErrorHandler} from './error/handler';
import {authInterceptor} from './common/services/global/interceptor';
import {MESSAGES_DI_TOKEN} from './index.messages';

// Import configuration
import {LocalConfigLoaderService} from './common/services/global/loader';

// Modern standalone bootstrap
bootstrapApplication(ModernRootComponent, {
  providers: [
    // Router configuration
    provideRouter(
      routes,
      withHashLocation(),
      withRouterConfig({
        onSameUrlNavigation: 'reload',
      })
    ),

    // HTTP client with interceptors
    provideHttpClient(
      withInterceptors([authInterceptor])
    ),

    // Animations
    provideAnimations(),

    // Error handling
    {
      provide: ErrorHandler,
      useClass: GlobalErrorHandler,
    },

    // Locale support
    {
      provide: LOCALE_ID,
      useValue: 'en-US',
    },

    // Material modules that need to be provided
    importProvidersFrom(
      MatDialogModule,
      MatSnackBarModule
    ),

    // Global services
    LocalConfigLoaderService,

    // Messages token
    {
      provide: MESSAGES_DI_TOKEN,
      useValue: {
        // Default messages - these would normally be loaded from i18n
        Minimize: 'Minimize',
        Expand: 'Expand',
        User: 'User',
        Settings: 'Settings',
        About: 'About',
        SignOut: 'Sign Out',
        Anonymous: 'Anonymous',
        NoUserLoggedIn: 'No user logged in',
        NoEmailAvailable: 'No email available',
        AuthSkipped: 'Authentication skipped',
      },
    },
  ],
}).catch(err => console.error(err));