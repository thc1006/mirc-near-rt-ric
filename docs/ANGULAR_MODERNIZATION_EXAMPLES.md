# Angular Modernization Examples

This document provides concrete before/after examples for migrating the Near-RT RIC Dashboard from Angular 13.3.4 to Angular 18.x, showcasing the adoption of new features and patterns.

## 1. Standalone Components Migration

### Before: Module-Based Architecture

```typescript
// card/module.ts
@NgModule({
  declarations: [CardComponent],
  imports: [
    CommonModule,
    MatCardModule,
    MatIconModule,
    MatButtonModule,
    MatProgressSpinnerModule,
    MatTooltipModule,
  ],
  exports: [CardComponent],
})
export class CardModule {}

// card/component.ts
@Component({
  selector: 'kd-card',
  templateUrl: './template.html',
  styleUrls: ['./style.scss'],
  animations: [Animations.expandInOut],
})
export class CardComponent {
  @Input() initialized = true;
  @Input() role: 'inner' | 'table' | 'inner-content';
  @Input() withFooter = false;
  @Input() expandable = true;
  @Input() expanded = true;
  private classes_: string[] = [];

  constructor(@Inject(MESSAGES_DI_TOKEN) readonly message: IMessage) {}
}

// Usage in other modules
@NgModule({
  imports: [CardModule],
  // ...
})
export class SomeFeatureModule {}
```

### After: Standalone Components

```typescript
// card/component.ts
@Component({
  selector: 'kd-card',
  templateUrl: './template.html',
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
export class CardComponent {
  // Input signals (Angular 17.1+)
  initialized = input(true);
  role = input<'inner' | 'table' | 'inner-content'>();
  withFooter = input(false);
  expandable = input(true);
  expanded = input(true);

  // Injected services using inject function
  private message = inject(MESSAGES_DI_TOKEN);
}

// Direct usage in other components
@Component({
  standalone: true,
  imports: [CardComponent], // Direct import
  template: `<kd-card>Content</kd-card>`,
})
export class SomeFeatureComponent {}
```

## 2. New Control Flow Syntax

### Before: Traditional Directives

```html
<!-- userpanel/template.html -->
<ng-container *ngIf="isLoginStatusInitialized">
  <div *ngIf="hasUsername; else noUsername">
    <div class="kd-user-display-name">{{ username }}</div>
    <div class="kd-user-email mat-caption kd-muted">
      {{ loginStatus?.email || 'No email available' }}
    </div>
  </div>
  <ng-template #noUsername>
    <div class="kd-user-display-name">Anonymous</div>
    <div class="kd-user-email mat-caption kd-muted">
      No user logged in
    </div>
  </ng-template>
</ng-container>

<button mat-menu-item
        *ngIf="isLoggedIn()"
        (click)="logout()">
  <mat-icon>exit_to_app</mat-icon>
  <span>Sign Out</span>
</button>

<tr *ngFor="let item of items; trackBy: trackByFn; let i = index">
  <td>{{ i + 1 }}</td>
  <td>{{ item.name }}</td>
</tr>
```

### After: New Control Flow Syntax

```html
<!-- userpanel/modern-template.html -->
@if (isLoginStatusInitialized()) {
  @if (hasUsername()) {
    <div class="kd-user-display-name">{{ username() }}</div>
    <div class="kd-user-email mat-caption kd-muted">
      {{ loginStatus()?.email || 'No email available' }}
    </div>
  } @else {
    <div class="kd-user-display-name">Anonymous</div>
    <div class="kd-user-email mat-caption kd-muted">
      No user logged in
    </div>
  }
}

@if (isLoggedIn()) {
  <button mat-menu-item (click)="logout()">
    <mat-icon>exit_to_app</mat-icon>
    <span>Sign Out</span>
  </button>
}

@for (item of items; track item.id; let i = $index) {
  <tr>
    <td>{{ i + 1 }}</td>
    <td>{{ item.name }}</td>
  </tr>
}
```

## 3. Signals Implementation

### Before: Traditional Reactive Patterns

```typescript
// userpanel/component.ts
export class UserPanelComponent implements OnInit, OnDestroy {
  loginStatus: LoginStatus | null = null;
  isLoginStatusInitialized = false;
  private destroy$ = new Subject<void>();

  constructor(
    private authService: AuthService,
    @Inject(MESSAGES_DI_TOKEN) readonly message: IMessage
  ) {}

  ngOnInit(): void {
    this.authService
      .getLoginStatus()
      .pipe(takeUntil(this.destroy$))
      .subscribe(status => {
        this.loginStatus = status;
        this.isLoginStatusInitialized = true;
      });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  get hasUsername(): boolean {
    return !!(this.loginStatus?.username && this.loginStatus.username.length > 0);
  }

  get username(): string {
    return this.loginStatus?.username || '';
  }
}
```

### After: Signals-Based

```typescript
// userpanel/modern-component.ts
export class ModernUserPanelComponent {
  // Injected services using modern inject function
  private authService = inject(AuthService);
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

  // Effects for reactive updates
  constructor() {
    effect(() => {
      this.authService
        .getLoginStatus()
        .pipe(takeUntilDestroyed())
        .subscribe(status => {
          this.loginStatus.set(status);
          this.isLoginStatusInitialized.set(true);
        });
    });
  }

  // Getter for template access
  get message(): IMessage {
    return this.messages;
  }
}
```

## 4. Flex Layout Migration

### Before: Angular Flex Layout

```html
<!-- card/template.html -->
<mat-card-title fxLayoutAlign=" center"
                (click)="onCardHeaderClick()">
  <div class="kd-card-title" fxFlex="100%">
    <ng-content select="[title]"></ng-content>
  </div>

  <div *ngIf="!expanded"
       class="kd-card-description kd-muted"
       fxLayoutAlign=" center"
       fxFlex="80">
    <ng-content select="[description]"></ng-content>
  </div>

  <div *ngIf="expanded && expandable"
       class="kd-card-actions kd-muted">
    <ng-content select="[actions]"></ng-content>
  </div>
</mat-card-title>
```

### After: CSS Grid/Flexbox

```html
<!-- card/modern-template.html -->
<mat-card-title class="layout-row align-center"
                (click)="onCardHeaderClick()">
  <div class="kd-card-title flex-auto">
    <ng-content select="[title]"></ng-content>
  </div>

  @if (!isExpanded()) {
    <div class="kd-card-description kd-muted layout-row align-center flex-80">
      <ng-content select="[description]"></ng-content>
    </div>
  }

  @if (isExpanded() && expandable()) {
    <div class="kd-card-actions kd-muted">
      <ng-content select="[actions]"></ng-content>
    </div>
  }
</mat-card-title>
```

```scss
// CSS utilities in modern-styles.scss
.layout-row { display: flex; flex-direction: row; }
.align-center { align-items: center; }
.flex-auto { flex: 1 1 auto; }
.flex-80 { flex: 0 1 80%; }
```

## 5. RxJS Migration

### Before: RxJS 6.x Patterns

```typescript
// search/component.ts
import { Subject } from 'rxjs';
import { debounceTime, distinctUntilChanged, takeUntil } from 'rxjs/operators';

export class SearchComponent implements OnInit, OnDestroy {
  private destroy$ = new Subject<void>();
  private searchSubject = new Subject<string>();
  
  ngOnInit(): void {
    this.searchSubject
      .pipe(
        debounceTime(300),
        distinctUntilChanged(),
        takeUntil(this.destroy$)
      )
      .subscribe(query => this.performSearch(query));
  }
  
  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }
  
  onSearchInput(query: string): void {
    this.searchSubject.next(query);
  }
}
```

### After: RxJS 7.x with takeUntilDestroyed

```typescript
// search/modern-component.ts
import { debounceTime, distinctUntilChanged } from 'rxjs';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

export class ModernSearchComponent {
  private destroyRef = inject(DestroyRef);
  private searchSubject = new Subject<string>();
  
  constructor() {
    this.searchSubject
      .pipe(
        debounceTime(300),
        distinctUntilChanged(),
        takeUntilDestroyed(this.destroyRef)
      )
      .subscribe(query => this.performSearch(query));
  }
  
  onSearchInput(query: string): void {
    this.searchSubject.next(query);
  }
}
```

## 6. Modern Bootstrap

### Before: NgModule Bootstrap

```typescript
// main.ts
import { platformBrowserDynamic } from '@angular/platform-browser-dynamic';
import { RootModule } from './app/index.module';

platformBrowserDynamic()
  .bootstrapModule(RootModule)
  .catch(err => console.error(err));

// index.module.ts
@NgModule({
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    CoreModule,
    ChromeModule,
    LoginModule,
    RouterModule.forRoot(routes, {
      useHash: true,
      onSameUrlNavigation: 'reload',
    }),
  ],
  providers: [
    { provide: ErrorHandler, useClass: GlobalErrorHandler },
    { provide: HTTP_INTERCEPTORS, useClass: AuthInterceptor, multi: true },
  ],
  declarations: [RootComponent],
  bootstrap: [RootComponent],
})
export class RootModule {}
```

### After: Standalone Bootstrap

```typescript
// modern-main.ts
import { bootstrapApplication } from '@angular/platform-browser';
import { provideAnimations } from '@angular/platform-browser/animations';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { provideRouter, withHashLocation } from '@angular/router';

import { ModernRootComponent } from './modern-root.component';
import { routes } from './index.routing';

bootstrapApplication(ModernRootComponent, {
  providers: [
    provideRouter(routes, withHashLocation()),
    provideHttpClient(withInterceptors([authInterceptor])),
    provideAnimations(),
    { provide: ErrorHandler, useClass: GlobalErrorHandler },
    // ... other providers
  ],
});

// modern-root.component.ts
@Component({
  selector: 'kd-root',
  template: `
    <div class="kd-app-container">
      <router-outlet></router-outlet>
    </div>
  `,
  standalone: true,
  imports: [RouterOutlet],
})
export class ModernRootComponent {}
```

## 7. Modern Route Guards

### Before: Class-Based Guards

```typescript
// auth.guard.ts
@Injectable()
export class AuthGuard implements CanActivate {
  constructor(
    private authService: AuthService,
    private router: Router
  ) {}

  canActivate(
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot
  ): boolean | Observable<boolean> {
    return this.authService.isAuthenticated().pipe(
      map(isAuth => {
        if (!isAuth) {
          this.router.navigate(['/login']);
          return false;
        }
        return true;
      })
    );
  }
}

// routing.ts
const routes: Routes = [
  {
    path: 'dashboard',
    component: DashboardComponent,
    canActivate: [AuthGuard],
  },
];
```

### After: Functional Guards

```typescript
// auth.guard.ts
export const authGuard = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  return authService.isAuthenticated().pipe(
    map(isAuth => {
      if (!isAuth) {
        router.navigate(['/login']);
        return false;
      }
      return true;
    })
  );
};

// routing.ts
const routes: Routes = [
  {
    path: 'dashboard',
    component: DashboardComponent,
    canActivate: [authGuard],
  },
];
```

## 8. Material 3 Design System

### Before: Material 2 Theming

```scss
// _theming.scss
@import '~@angular/material/theming';

$primary: mat-palette($mat-indigo);
$accent: mat-palette($mat-pink, A200, A100, A400);
$warn: mat-palette($mat-red);

$theme: mat-light-theme($primary, $accent, $warn);

@include angular-material-theme($theme);

$dark-primary: mat-palette($mat-blue-grey);
$dark-accent: mat-palette($mat-amber, A200, A100, A400);
$dark-warn: mat-palette($mat-deep-orange);

$dark-theme: mat-dark-theme($dark-primary, $dark-accent, $dark-warn);

.kd-dark-theme {
  @include angular-material-theme($dark-theme);
}
```

### After: Material 3 Theming

```scss
// modern-theming.scss
@use '@angular/material' as mat;

$light-theme: mat.define-theme((
  color: (
    theme-type: light,
    primary: mat.$azure-palette,
    tertiary: mat.$blue-palette,
  ),
  typography: (
    brand-family: 'Roboto',
    plain-family: 'Roboto',
  ),
  density: (
    scale: 0,
  )
));

$dark-theme: mat.define-theme((
  color: (
    theme-type: dark,
    primary: mat.$azure-palette,
    tertiary: mat.$blue-palette,
  ),
  typography: (
    brand-family: 'Roboto',
    plain-family: 'Roboto',
  ),
  density: (
    scale: 0,
  )
));

html {
  @include mat.all-component-themes($light-theme);
}

.kd-dark-theme {
  @include mat.all-component-themes($dark-theme);
}
```

## 9. Modern Form Handling

### Before: Traditional Forms

```typescript
// create-form.component.ts
export class CreateFormComponent implements OnInit {
  form: FormGroup;
  
  constructor(private fb: FormBuilder) {
    this.form = this.fb.group({
      name: ['', [Validators.required, Validators.minLength(3)]],
      description: [''],
      labels: this.fb.array([]),
    });
  }
  
  get nameControl(): AbstractControl {
    return this.form.get('name')!;
  }
  
  get nameErrors(): string[] {
    const control = this.nameControl;
    const errors: string[] = [];
    
    if (control.hasError('required')) {
      errors.push('Name is required');
    }
    if (control.hasError('minlength')) {
      errors.push('Name must be at least 3 characters');
    }
    
    return errors;
  }
}
```

### After: Modern Forms with Signals

```typescript
// modern-create-form.component.ts
export class ModernCreateFormComponent {
  private fb = inject(FormBuilder);
  
  form = this.fb.group({
    name: ['', [Validators.required, Validators.minLength(3)]],
    description: [''],
    labels: this.fb.array([]),
  });
  
  // Signal-based form state
  nameControl = signal(this.form.get('name')!);
  
  nameErrors = computed(() => {
    const control = this.nameControl();
    const errors: string[] = [];
    
    if (control.hasError('required')) {
      errors.push('Name is required');
    }
    if (control.hasError('minlength')) {
      errors.push('Name must be at least 3 characters');
    }
    
    return errors;
  });
  
  isFormValid = computed(() => this.form.valid);
}
```

## 10. Modern HTTP Client

### Before: Traditional HTTP

```typescript
// data.service.ts
@Injectable()
export class DataService {
  constructor(private http: HttpClient) {}
  
  getData(): Observable<any[]> {
    return this.http.get<any[]>('/api/data');
  }
  
  postData(data: any): Observable<any> {
    return this.http.post<any>('/api/data', data);
  }
}
```

### After: Modern HTTP with Signals

```typescript
// modern-data.service.ts
@Injectable()
export class ModernDataService {
  private http = inject(HttpClient);
  
  // Signal-based resource loading
  data = signal<any[]>([]);
  loading = signal(false);
  error = signal<string | null>(null);
  
  async loadData(): Promise<void> {
    this.loading.set(true);
    this.error.set(null);
    
    try {
      const data = await firstValueFrom(this.http.get<any[]>('/api/data'));
      this.data.set(data);
    } catch (error) {
      this.error.set(error instanceof Error ? error.message : 'Unknown error');
    } finally {
      this.loading.set(false);
    }
  }
  
  async saveData(data: any): Promise<any> {
    return firstValueFrom(this.http.post<any>('/api/data', data));
  }
}
```

## Summary of Benefits

### Performance Improvements
- **Bundle Size**: 15-25% reduction with standalone components
- **Runtime Performance**: Up to 90% faster template rendering with new control flow
- **Memory Usage**: Reduced with signals and fine-grained reactivity

### Developer Experience
- **Type Safety**: Better type inference with signals and new APIs
- **Less Boilerplate**: Reduced code with standalone components and functional guards
- **Better Tooling**: Enhanced IDE support and debugging

### Maintainability
- **Modern Patterns**: Contemporary Angular best practices
- **Future-Proofing**: Ready for upcoming Angular features
- **Simplified Testing**: Easier unit testing with standalone components

This modernization will transform the codebase into a contemporary Angular application while maintaining functionality and improving performance and developer experience.