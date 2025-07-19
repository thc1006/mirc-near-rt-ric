# Go Modernization: Before & After Examples

This document provides concrete before/after examples showing how to modernize the dashboard backend codebase using contemporary Go patterns and features.

## 1. Receiver Naming Convention

### Before (Outdated - uses `self`)
```go
func (self authManager) Login(spec *authApi.LoginSpec) (*authApi.AuthResponse, error) {
    authenticator, err := self.getAuthenticator(spec)
    if err != nil {
        return nil, err
    }
    
    authInfo, err := authenticator.GetAuthInfo()
    if err != nil {
        return nil, err
    }
    
    username, err := self.healthCheck(authInfo)
    // ...
}

func (self DataSelector) Len() int { 
    return len(self.GenericDataList) 
}
```

### After (Modern - uses type abbreviations)
```go
func (am authManager) Login(spec *authApi.LoginSpec) (*authApi.AuthResponse, error) {
    authenticator, err := am.getAuthenticator(spec)
    if err != nil {
        return nil, err
    }
    
    authInfo, err := authenticator.GetAuthInfo()
    if err != nil {
        return nil, err
    }
    
    username, err := am.healthCheck(authInfo)
    // ...
}

func (ds DataSelector) Len() int { 
    return len(ds.GenericDataList) 
}
```

## 2. Error Handling Modernization

### Before (Custom error package)
```go
// From auth/manager.go
func (self authManager) getAuthenticator(spec *authApi.LoginSpec) (authApi.Authenticator, error) {
    if len(self.authenticationModes) == 0 {
        return nil, errors.NewInvalid("All authentication options disabled. Check --authentication-modes argument for more information.")
    }
    
    switch {
    case len(spec.Token) > 0 && self.authenticationModes.IsEnabled(authApi.Token):
        return NewTokenAuthenticator(spec), nil
    case len(spec.Username) > 0 && len(spec.Password) > 0 && self.authenticationModes.IsEnabled(authApi.Basic):
        return NewBasicAuthenticator(spec), nil
    case len(spec.KubeConfig) > 0:
        return NewKubeConfigAuthenticator(spec, self.authenticationModes), nil
    }
    
    return nil, errors.NewInvalid("Not enough data to create authenticator.")
}
```

### After (Modern error wrapping and sentinel errors)
```go
import (
    "errors"
    "fmt"
)

var (
    ErrAllAuthDisabled = errors.New("all authentication options disabled")
    ErrAuthDataInsufficient = errors.New("insufficient authentication data")
)

func (am authManager) getAuthenticator(spec *authApi.LoginSpec) (authApi.Authenticator, error) {
    if len(am.authenticationModes) == 0 {
        return nil, fmt.Errorf("%w: check --authentication-modes argument for more information",
            ErrAllAuthDisabled)
    }
    
    switch {
    case len(spec.Token) > 0 && am.authenticationModes.IsEnabled(authApi.Token):
        return NewTokenAuthenticator(spec), nil
    case len(spec.Username) > 0 && len(spec.Password) > 0 && am.authenticationModes.IsEnabled(authApi.Basic):
        return NewBasicAuthenticator(spec), nil
    case len(spec.KubeConfig) > 0:
        return NewKubeConfigAuthenticator(spec, am.authenticationModes), nil
    }
    
    return nil, fmt.Errorf("%w: missing token, credentials, or kubeconfig", ErrAuthDataInsufficient)
}

// Error checking with errors.Is
func handleAuthError(err error) {
    if errors.Is(err, ErrAllAuthDisabled) {
        // Handle disabled auth
    } else if errors.Is(err, ErrAuthDataInsufficient) {
        // Handle insufficient data
    }
}
```

## 3. Interface{} to Generics Migration

### Before (Type-unsafe with interface{})
```go
// From cert/api/types.go
type Creator interface {
    GenerateKey() interface{}
    GenerateCertificate(key interface{}) []byte
    StoreCertificates(path string, key interface{}, certBytes []byte)
    KeyCertPEMBytes(key interface{}, certBytes []byte) (keyPEM []byte, certPEM []byte, err error)
}

// From cert/ecdsa/creator.go
func (self *ecdsaCreator) GenerateKey() interface{} {
    key, err := ecdsa.GenerateKey(self.curve, rand.Reader)
    if err != nil {
        log.Fatalf("Failed to generate key: %s", err)
    }
    return key
}

func (self *ecdsaCreator) GenerateCertificate(key interface{}) []byte {
    ecdsaKey := self.getKey(key)
    // ... implementation
}
```

### After (Type-safe with generics)
```go
// Generic certificate creator
type Creator[K any] interface {
    GenerateKey() K
    GenerateCertificate(key K) []byte
    StoreCertificates(path string, key K, certBytes []byte)
    KeyCertPEMBytes(key K, certBytes []byte) (keyPEM []byte, certPEM []byte, err error)
}

// Specific type aliases
type ECDSACreator = Creator[*ecdsa.PrivateKey]
type RSACreator = Creator[*rsa.PrivateKey]

// Type-safe implementation
type ecdsaCreator struct {
    curve     elliptic.Curve
    keyFile   string
    certFile  string
}

func (ec *ecdsaCreator) GenerateKey() *ecdsa.PrivateKey {
    key, err := ecdsa.GenerateKey(ec.curve, rand.Reader)
    if err != nil {
        log.Fatalf("Failed to generate key: %s", err)
    }
    return key
}

func (ec *ecdsaCreator) GenerateCertificate(key *ecdsa.PrivateKey) []byte {
    // No type assertion needed - key is already the correct type
    // ... implementation
}
```

## 4. Data Selection Framework Modernization

### Before (Type-unsafe data selection)
```go
// From resource/dataselect/dataselect.go
type DataCell interface {
    GetProperty(PropertyName) ComparableValue
}

type DataSelector struct {
    GenericDataList []DataCell
    DataSelectQuery *DataSelectQuery
    // ...
}

func (self DataSelector) Less(i, j int) bool {
    for _, sortBy := range self.DataSelectQuery.SortQuery.SortByList {
        a := self.GenericDataList[i].GetProperty(sortBy.Property)
        b := self.GenericDataList[j].GetProperty(sortBy.Property)
        // ignore sort completely if property name not found
        if a == nil || b == nil {
            break
        }
        cmp := a.Compare(b)
        if cmp == 0 {
            continue
        } else {
            return (cmp == -1 && sortBy.Ascending) || (cmp == 1 && !sortBy.Ascending)
        }
    }
    return false
}
```

### After (Type-safe with generics and modern patterns)
```go
type ModernDataCell[T any] interface {
    GetProperty(name PropertyName) ModernComparableValue[T]
    GetData() T
}

type ModernDataSelector[T any, C ModernDataCell[T]] struct {
    dataList        []C
    query          *DataSelectQuery
    logger         *slog.Logger
}

func (mds *ModernDataSelector[T, C]) Sort() *ModernDataSelector[T, C] {
    if mds.query == nil || mds.query.SortQuery == nil {
        return mds
    }

    slices.SortFunc(mds.dataList, func(a, b C) int {
        for _, sortBy := range mds.query.SortQuery.SortByList {
            propA := a.GetProperty(sortBy.Property)
            propB := b.GetProperty(sortBy.Property)

            if propA == nil || propB == nil {
                mds.logger.Warn("sort property not found",
                    "property", sortBy.Property,
                    "item_type", fmt.Sprintf("%T", a))
                continue
            }

            comparison := propA.Compare(propB.Value())
            if comparison != 0 {
                if sortBy.Ascending {
                    return comparison
                }
                return -comparison
            }
        }
        return 0
    })

    return mds
}
```

## 5. Channel Management Modernization

### Before (Repetitive channel definitions)
```go
// From resource/common/resourcechannels.go
type ServiceListChannel struct {
    List  chan *v1.ServiceList
    Error chan error
}

type PodListChannel struct {
    List  chan *v1.PodList
    Error chan error
}

type DeploymentListChannel struct {
    List  chan *appsv1.DeploymentList
    Error chan error
}

// Manual sending
func GetServiceListChannel(client kubernetes.Interface, nsQuery *NamespaceQuery, numReads int) ServiceListChannel {
    channel := ServiceListChannel{
        List:  make(chan *v1.ServiceList, numReads),
        Error: make(chan error, numReads),
    }
    
    go func() {
        list, err := client.CoreV1().Services(nsQuery.ToRequestParam()).List(context.TODO(), metav1.ListOptions{})
        if err != nil {
            channel.Error <- err
            return
        }
        channel.List <- list
    }()
    
    return channel
}
```

### After (Generic channel management)
```go
type ResourceChannel[T any] struct {
    List  chan *T
    Error chan error
    once  sync.Once
}

func NewResourceChannel[T any]() ResourceChannel[T] {
    return ResourceChannel[T]{
        List:  make(chan *T, 1),
        Error: make(chan error, 1),
    }
}

func (rc *ResourceChannel[T]) SendList(list *T) {
    rc.once.Do(func() {
        rc.List <- list
        close(rc.List)
        close(rc.Error)
    })
}

func (rc *ResourceChannel[T]) SendError(err error) {
    rc.once.Do(func() {
        rc.Error <- err
        close(rc.List)
        close(rc.Error)
    })
}

func (rc *ResourceChannel[T]) Receive(ctx context.Context) (*T, error) {
    select {
    case list := <-rc.List:
        return list, nil
    case err := <-rc.Error:
        return nil, err
    case <-ctx.Done():
        return nil, fmt.Errorf("resource channel receive cancelled: %w", ctx.Err())
    }
}

// Usage
type ResourceChannels struct {
    ServiceList    ResourceChannel[v1.ServiceList]
    PodList        ResourceChannel[v1.PodList]
    DeploymentList ResourceChannel[appsv1.DeploymentList]
}
```

## 6. Logging Modernization

### Before (Basic logging)
```go
import "log"

log.Printf("Using apiserver-host location: %s", args.Holder.GetApiServerHost())
log.Printf("Successful initial request to the apiserver, version: %s", versionInfo.String())
log.Print("no metrics provider selected, will not check metrics.")
```

### After (Structured logging with slog)
```go
import "log/slog"

// Setup structured logger
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Usage with structured attributes
logger.Info("apiserver configuration",
    "host", args.Holder.GetApiServerHost(),
    "kubeconfig", args.Holder.GetKubeConfigFile(),
    "namespace", args.Holder.GetNamespace())

logger.Info("apiserver connection established",
    "version", versionInfo.String(),
    "git_version", versionInfo.GitVersion)

logger.Warn("no metrics provider configured",
    "available_providers", []string{"sidecar", "heapster"},
    "recommendation", "configure a metrics provider for monitoring")
```

## 7. Context-Aware Operations

### Before (No context usage)
```go
func (self *clientManager) HasAccess(authInfo api.AuthInfo) (string, error) {
    client, err := self.ClientFromAuthInfo(authInfo)
    if err != nil {
        return "", err
    }
    
    user, err := client.AuthenticationV1().SelfSubjectReviews().Create(&authv1.SelfSubjectReview{}, metav1.CreateOptions{})
    if err != nil {
        return "", err
    }
    
    return user.Status.UserInfo.Username, nil
}
```

### After (Context-aware with timeouts)
```go
func (cm *clientManager) HasAccess(authInfo api.AuthInfo) (string, error) {
    client, err := cm.ClientFromAuthInfo(authInfo)
    if err != nil {
        return "", fmt.Errorf("failed to create client from auth info: %w", err)
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    user, err := client.AuthenticationV1().SelfSubjectReviews().Create(ctx, &authv1.SelfSubjectReview{}, metav1.CreateOptions{})
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return "", fmt.Errorf("access verification timeout: %w", err)
        }
        return "", fmt.Errorf("failed to verify access for user: %w", err)
    }
    
    return user.Status.UserInfo.Username, nil
}
```

## 8. Modern String Operations

### Before (Manual string operations)
```go
// Splitting authorization header
func parseAuthHeader(header string) (string, string) {
    parts := strings.Split(header, " ")
    if len(parts) != 2 {
        return "", ""
    }
    return strings.ToLower(parts[0]), parts[1]
}
```

### After (Using strings.Cut from Go 1.18+)
```go
func parseAuthHeader(header string) (scheme, token string, ok bool) {
    scheme, token, ok = strings.Cut(header, " ")
    if !ok {
        return "", "", false
    }
    return strings.ToLower(scheme), token, true
}
```

## 9. Complete Auth Manager Modernization

### Before (Original implementation)
```go
// From auth/manager.go
type authManager struct {
    tokenManager            authApi.TokenManager
    clientManager           clientapi.ClientManager
    authenticationModes     authApi.AuthenticationModes
    authenticationSkippable bool
}

func (self authManager) Login(spec *authApi.LoginSpec) (*authApi.AuthResponse, error) {
    authenticator, err := self.getAuthenticator(spec)
    if err != nil {
        return nil, err
    }

    authInfo, err := authenticator.GetAuthInfo()
    if err != nil {
        return nil, err
    }

    username, err := self.healthCheck(authInfo)
    nonCriticalErrors, criticalError := errors.HandleError(err)
    if criticalError != nil || len(nonCriticalErrors) > 0 {
        return &authApi.AuthResponse{Errors: nonCriticalErrors}, criticalError
    }

    token, err := self.tokenManager.Generate(authInfo)
    if err != nil {
        return nil, err
    }

    return &authApi.AuthResponse{JWEToken: token, Errors: nonCriticalErrors, Name: username}, nil
}
```

### After (Modern implementation with logging, metrics, context)
```go
type ModernAuthManager struct {
    tokenManager            authApi.TokenManager
    clientManager           clientapi.ClientManager
    authenticationModes     authApi.AuthenticationModes
    authenticationSkippable bool
    logger                  *slog.Logger
    healthCheckTimeout      time.Duration
    metrics                 *AuthMetrics
}

func (am *ModernAuthManager) Login(spec *authApi.LoginSpec) (*authApi.AuthResponse, error) {
    ctx := context.Background()
    startTime := time.Now()
    am.metrics.LoginAttempts++
    
    am.logger.Info("login attempt started",
        "has_token", len(spec.Token) > 0,
        "has_credentials", len(spec.Username) > 0 && len(spec.Password) > 0,
        "has_kubeconfig", len(spec.KubeConfig) > 0)
    
    authenticator, err := am.getAuthenticator(spec)
    if err != nil {
        am.metrics.FailedLogins++
        am.logger.Error("failed to create authenticator",
            "error", err,
            "duration", time.Since(startTime))
        return nil, fmt.Errorf("authenticator creation failed: %w", err)
    }
    
    authInfo, err := authenticator.GetAuthInfo()
    if err != nil {
        am.metrics.FailedLogins++
        am.logger.Error("failed to get auth info",
            "error", err,
            "authenticator_type", fmt.Sprintf("%T", authenticator),
            "duration", time.Since(startTime))
        return nil, fmt.Errorf("auth info extraction failed: %w", err)
    }
    
    // Health check with timeout
    ctx, cancel := context.WithTimeout(ctx, am.healthCheckTimeout)
    defer cancel()
    
    username, err := am.healthCheckWithContext(ctx, authInfo)
    if err != nil {
        am.metrics.FailedLogins++
        nonCriticalErrors, criticalError := errors.HandleErrors([]error{err})
        if criticalError != nil {
            am.logger.Error("critical error during health check",
                "error", criticalError,
                "duration", time.Since(startTime))
            return &authApi.AuthResponse{Errors: nonCriticalErrors}, criticalError
        }
        
        if len(nonCriticalErrors) > 0 {
            am.logger.Warn("non-critical errors during health check",
                "errors", nonCriticalErrors,
                "duration", time.Since(startTime))
            return &authApi.AuthResponse{Errors: nonCriticalErrors}, nil
        }
    }
    
    token, err := am.tokenManager.Generate(authInfo)
    if err != nil {
        am.metrics.FailedLogins++
        am.logger.Error("failed to generate token",
            "error", err,
            "username", username,
            "duration", time.Since(startTime))
        return nil, fmt.Errorf("token generation failed: %w", err)
    }
    
    am.metrics.SuccessfulLogins++
    am.logger.Info("login successful",
        "username", username,
        "duration", time.Since(startTime),
        "token_length", len(token))
    
    return &authApi.AuthResponse{
        JWEToken: token,
        Name:     username,
        Errors:   nil,
    }, nil
}
```

## Benefits of Modernization

### Type Safety
- **Before**: Runtime panics from type assertions
- **After**: Compile-time error detection with generics

### Performance
- **Before**: Boxing/unboxing overhead with interface{}
- **After**: Zero-cost abstractions with generics

### Observability
- **Before**: Basic print statements
- **After**: Structured logging with context and metrics

### Error Handling
- **Before**: String-based error messages
- **After**: Typed errors with wrapping and context

### Maintainability
- **Before**: Repetitive boilerplate code
- **After**: DRY principles with generics and modern patterns

## Migration Strategy

1. **Phase 1**: Update receiver names and basic patterns
2. **Phase 2**: Introduce structured logging
3. **Phase 3**: Modernize error handling
4. **Phase 4**: Implement generics for type safety
5. **Phase 5**: Add context awareness and metrics

Each phase can be implemented incrementally without breaking existing functionality.