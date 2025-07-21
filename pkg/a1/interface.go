package a1

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/hctsai1006/near-rt-ric/internal/config"
	"github.com/hctsai1006/near-rt-ric/pkg/common/database"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// A1Interface implements the O-RAN A1 interface for policy management
type A1Interface struct {
	config       *config.A1Config
	repository   PolicyRepository
	validator    PolicyValidator
	notifier     NotificationService
	authenticator *JWTAuthenticator
	router       *mux.Router
	server       *http.Server
	rateLimiter  *rate.Limiter
	logger       *logrus.Logger
	
	// Metrics
	metrics      *A1Metrics
	
	// Lifecycle management
	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
	mutex        sync.RWMutex
	startTime    time.Time
}

// A1Status represents the current status of the A1 interface
type A1Status struct {
	Status           string            `json:"status"`
	Version          string            `json:"version"`
	Uptime           time.Duration     `json:"uptime"`
	PolicyTypes      int               `json:"policy_types"`
	ActivePolicies   int               `json:"active_policies"`
	TotalPolicies    int               `json:"total_policies"`
	LastRequestTime  time.Time         `json:"last_request_time"`
	RequestsHandled  int64             `json:"requests_handled"`
	ErrorRate        float64           `json:"error_rate"`
	AvgResponseTime  time.Duration     `json:"avg_response_time"`
}

// JWTAuthenticator handles JWT token validation and creation
type JWTAuthenticator struct {
	publicKey       *rsa.PublicKey
	privateKey      *rsa.PrivateKey
	secretKey       []byte
	algorithm       string
	issuer          string
	audience        []string
	tokenExpiry     time.Duration
	refreshExpiry   time.Duration
	logger          *logrus.Logger
}

// Claims represents JWT claims structure
type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// NewA1Interface creates a new A1Interface with full O-RAN compliance
func NewA1Interface(cfg *config.A1Config, db database.Interface, logger *logrus.Logger) (*A1Interface, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create component logger
	compLogger := logger.WithField("component", "a1-interface")
	
	// Initialize repository
	repository, err := NewPolicyRepository(db, compLogger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create policy repository: %w", err)
	}
	
	// Initialize validator
	validator := NewPolicyValidator(compLogger)
	
	// Initialize notification service
	var notifier NotificationService
	if cfg.PolicyStorage.Backend != "memory" {
		notifier, err = NewNotificationService(cfg, compLogger)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create notification service: %w", err)
		}
	}
	
	// Initialize JWT authenticator
	authenticator, err := NewJWTAuthenticator(&cfg.Auth.JWT, compLogger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create JWT authenticator: %w", err)
	}
	
	// Initialize rate limiter
	rateLimiter := rate.NewLimiter(
		rate.Limit(cfg.RateLimit.RequestsPerSecond),
		cfg.RateLimit.BurstSize,
	)
	
	// Initialize metrics
	metrics := NewA1Metrics()
	
	// Create A1 interface
	a1Interface := &A1Interface{
		config:        cfg,
		repository:    repository,
		validator:     validator,
		notifier:      notifier,
		authenticator: authenticator,
		rateLimiter:   rateLimiter,
		logger:        compLogger,
		metrics:       metrics,
		ctx:           ctx,
		cancel:        cancel,
		running:       false,
		startTime:     time.Now(),
	}
	
	// Setup HTTP routes
	if err := a1Interface.setupRoutes(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to setup routes: %w", err)
	}
	
	return a1Interface, nil
}

// NewJWTAuthenticator creates a new JWT authenticator
func NewJWTAuthenticator(cfg *config.JWTConfig, logger *logrus.Logger) (*JWTAuthenticator, error) {
	auth := &JWTAuthenticator{
		algorithm:     cfg.Algorithm,
		issuer:        cfg.Issuer,
		audience:      cfg.Audience,
		tokenExpiry:   cfg.TokenExpiry,
		refreshExpiry: cfg.RefreshExpiry,
		logger:        logger,
	}
	
	// Load keys based on algorithm
	switch cfg.Algorithm {
	case "HS256", "HS384", "HS512":
		if cfg.SecretKey == "" {
			return nil, fmt.Errorf("secret key is required for HMAC algorithms")
		}
		auth.secretKey = []byte(cfg.SecretKey)
		
	case "RS256", "RS384", "RS512":
		if cfg.PublicKeyFile == "" || cfg.PrivateKeyFile == "" {
			return nil, fmt.Errorf("public and private key files are required for RSA algorithms")
		}
		
		// Load RSA keys (implementation would read from files)
		// For now, using placeholder
		return nil, fmt.Errorf("RSA key loading not implemented in this example")
		
	default:
		return nil, fmt.Errorf("unsupported JWT algorithm: %s", cfg.Algorithm)
	}
	
	return auth, nil
}

// Start starts the A1 interface HTTP server
func (a1 *A1Interface) Start(ctx context.Context) error {
	a1.mutex.Lock()
	defer a1.mutex.Unlock()
	
	if a1.running {
		return fmt.Errorf("A1 interface is already running")
	}
	
	a1.logger.WithFields(logrus.Fields{
		"listen_address": a1.config.ListenAddress,
		"port":           a1.config.Port,
		"tls_enabled":    a1.config.TLS.Enabled,
		"auth_enabled":   a1.config.Auth.Enabled,
	}).Info("Starting O-RAN A1 interface")
	
	// Register Prometheus metrics
	if err := a1.metrics.Register(); err != nil {
		a1.logger.WithError(err).Warn("Failed to register A1 metrics")
	}
	
	// Initialize standard policy types
	if err := a1.initializeStandardPolicyTypes(); err != nil {
		return fmt.Errorf("failed to initialize standard policy types: %w", err)
	}
	
	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", a1.config.ListenAddress, a1.config.Port)
	a1.server = &http.Server{
		Addr:           addr,
		Handler:        a1.router,
		ReadTimeout:    a1.config.ReadTimeout,
		WriteTimeout:   a1.config.WriteTimeout,
		IdleTimeout:    a1.config.IdleTimeout,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
	
	// Start server in goroutine
	go func() {
		var err error
		if a1.config.TLS.Enabled {
			a1.logger.Info("Starting A1 interface with TLS")
			err = a1.server.ListenAndServeTLS(a1.config.TLS.CertFile, a1.config.TLS.KeyFile)
		} else {
			a1.logger.Info("Starting A1 interface without TLS")
			err = a1.server.ListenAndServe()
		}
		
		if err != nil && err != http.ErrServerClosed {
			a1.logger.WithError(err).Error("A1 interface server error")
		}
	}()
	
	a1.running = true
	a1.startTime = time.Now()
	a1.logger.Info("O-RAN A1 interface started successfully")
	
	// Record startup metric
	a1.metrics.InterfaceStatus.WithLabelValues("a1").Set(1)
	
	return nil
}

// Stop gracefully stops the A1 interface HTTP server
func (a1 *A1Interface) Stop(ctx context.Context) error {
	a1.mutex.Lock()
	defer a1.mutex.Unlock()
	
	if !a1.running {
		return nil
	}
	
	a1.logger.Info("Stopping O-RAN A1 interface")
	
	// Cancel internal context
	a1.cancel()
	
	// Shutdown HTTP server with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	if err := a1.server.Shutdown(shutdownCtx); err != nil {
		a1.logger.WithError(err).Error("Error shutting down A1 interface server")
		return err
	}
	
	// Cleanup notification service
	if a1.notifier != nil {
		if err := a1.notifier.Close(); err != nil {
			a1.logger.WithError(err).Error("Error closing notification service")
		}
	}
	
	// Unregister metrics
	a1.metrics.Unregister()
	
	a1.running = false
	a1.logger.Info("O-RAN A1 interface stopped successfully")
	
	// Record shutdown metric
	a1.metrics.InterfaceStatus.WithLabelValues("a1").Set(0)
	
	return nil
}

// HealthCheck performs a comprehensive health check
func (a1 *A1Interface) HealthCheck() error {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()
	
	if !a1.running {
		return fmt.Errorf("A1 interface is not running")
	}
	
	// Check repository health
	if err := a1.repository.HealthCheck(); err != nil {
		return fmt.Errorf("repository health check failed: %w", err)
	}
	
	// Check notification service health
	if a1.notifier != nil {
		if err := a1.notifier.HealthCheck(); err != nil {
			return fmt.Errorf("notification service health check failed: %w", err)
		}
	}
	
	return nil
}

// GetStatus returns the current status of the A1 interface
func (a1 *A1Interface) GetStatus() *A1Status {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()
	
	stats := a1.repository.GetStatistics()
	
	return &A1Status{
		Status:          "OK",
		Version:         "1.0.0",
		Uptime:          time.Since(a1.startTime),
		PolicyTypes:     stats.PolicyTypes,
		ActivePolicies:  stats.ActivePolicies,
		TotalPolicies:   stats.TotalPolicies,
		LastRequestTime: stats.LastRequestTime,
		RequestsHandled: stats.RequestsHandled,
		ErrorRate:       stats.ErrorRate,
		AvgResponseTime: stats.AvgResponseTime,
	}
}

// setupRoutes configures HTTP routes according to O-RAN A1 API specification
func (a1 *A1Interface) setupRoutes() error {
	a1.router = mux.NewRouter()
	
	// Add middleware in order
	a1.router.Use(a1.loggingMiddleware)
	a1.router.Use(a1.corsMiddleware)
	a1.router.Use(a1.rateLimitMiddleware)
	a1.router.Use(a1.securityHeadersMiddleware)
	
	// Health check endpoint (no auth required)
	a1.router.HandleFunc("/a1-p/healthcheck", a1.handleHealthCheck).Methods("GET")
	
	// Create authenticated subrouter
	authRouter := a1.router.PathPrefix("/a1-p").Subrouter()
	if a1.config.Auth.Enabled {
		authRouter.Use(a1.authenticationMiddleware)
	}
	
	// Policy type management endpoints
	authRouter.HandleFunc("/policytypes", a1.handleGetPolicyTypes).Methods("GET")
	authRouter.HandleFunc("/policytypes/{policy_type_id}", a1.handleGetPolicyType).Methods("GET")
	authRouter.HandleFunc("/policytypes/{policy_type_id}", a1.handleCreatePolicyType).Methods("PUT")
	authRouter.HandleFunc("/policytypes/{policy_type_id}", a1.handleDeletePolicyType).Methods("DELETE")
	
	// Policy instance management endpoints
	authRouter.HandleFunc("/policytypes/{policy_type_id}/policies", a1.handleGetPolicies).Methods("GET")
	authRouter.HandleFunc("/policytypes/{policy_type_id}/policies/{policy_id}", a1.handleGetPolicy).Methods("GET")
	authRouter.HandleFunc("/policytypes/{policy_type_id}/policies/{policy_id}", a1.handleCreateOrUpdatePolicy).Methods("PUT")
	authRouter.HandleFunc("/policytypes/{policy_type_id}/policies/{policy_id}", a1.handleDeletePolicy).Methods("DELETE")
	
	// Policy status endpoints
	authRouter.HandleFunc("/policytypes/{policy_type_id}/policies/{policy_id}/status", a1.handleGetPolicyStatus).Methods("GET")
	
	// Data delivery endpoints (for xApp notifications)
	authRouter.HandleFunc("/data-delivery", a1.handleDataDelivery).Methods("POST")
	
	// Authentication endpoints
	if a1.config.Auth.Enabled {
		a1.router.HandleFunc("/a1-p/auth/login", a1.handleLogin).Methods("POST")
		a1.router.HandleFunc("/a1-p/auth/refresh", a1.handleRefreshToken).Methods("POST")
		a1.router.HandleFunc("/a1-p/auth/logout", a1.handleLogout).Methods("POST")
	}
	
	// Metrics endpoint
	a1.router.HandleFunc("/metrics", a1.handleMetrics).Methods("GET")
	
	return nil
}

// HTTP Handlers

func (a1 *A1Interface) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		a1.metrics.RequestDuration.WithLabelValues("GET", "/healthcheck").Observe(time.Since(start).Seconds())
		a1.metrics.RequestsTotal.WithLabelValues("GET", "/healthcheck", "200").Inc()
	}()
	
	status := a1.GetStatus()
	a1.writeJSONResponse(w, http.StatusOK, status)
}

func (a1 *A1Interface) handleGetPolicyTypes(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		a1.metrics.RequestDuration.WithLabelValues("GET", "/policytypes").Observe(time.Since(start).Seconds())
	}()
	
	policyTypes, err := a1.repository.ListPolicyTypes()
	if err != nil {
		a1.writeErrorResponse(w, http.StatusInternalServerError, "internal_error",
			"Failed to retrieve policy types", err)
		a1.metrics.RequestsTotal.WithLabelValues("GET", "/policytypes", "500").Inc()
		return
	}
	
	// Extract policy type IDs for O-RAN compliance
	policyTypeIDs := make([]string, len(policyTypes))
	for i, pt := range policyTypes {
		policyTypeIDs[i] = pt.ID
	}
	
	a1.writeJSONResponse(w, http.StatusOK, policyTypeIDs)
	a1.metrics.RequestsTotal.WithLabelValues("GET", "/policytypes", "200").Inc()
}

func (a1 *A1Interface) handleGetPolicyType(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]
	
	defer func() {
		a1.metrics.RequestDuration.WithLabelValues("GET", "/policytypes/*").Observe(time.Since(start).Seconds())
	}()
	
	policyType, err := a1.repository.GetPolicyType(policyTypeID)
	if err != nil {
		statusCode := http.StatusNotFound
		if err != ErrNotFound {
			statusCode = http.StatusInternalServerError
		}
		
		a1.writeErrorResponse(w, statusCode, "not_found",
			fmt.Sprintf("Policy type %s not found", policyTypeID), err)
		a1.metrics.RequestsTotal.WithLabelValues("GET", "/policytypes/*", fmt.Sprintf("%d", statusCode)).Inc()
		return
	}
	
	a1.writeJSONResponse(w, http.StatusOK, policyType.Schema)
	a1.metrics.RequestsTotal.WithLabelValues("GET", "/policytypes/*", "200").Inc()
}

func (a1 *A1Interface) handleCreatePolicyType(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]
	
	defer func() {
		a1.metrics.RequestDuration.WithLabelValues("PUT", "/policytypes/*").Observe(time.Since(start).Seconds())
	}()
	
	// Parse request body
	var schema map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&schema); err != nil {
		a1.writeErrorResponse(w, http.StatusBadRequest, "bad_request",
			"Invalid JSON in request body", err)
		a1.metrics.RequestsTotal.WithLabelValues("PUT", "/policytypes/*", "400").Inc()
		return
	}
	
	// Validate schema
	if err := a1.validator.ValidateSchema(schema); err != nil {
		a1.writeErrorResponse(w, http.StatusBadRequest, "invalid_schema",
			"Invalid policy type schema", err)
		a1.metrics.RequestsTotal.WithLabelValues("PUT", "/policytypes/*", "400").Inc()
		return
	}
	
	// Create policy type
	policyType := &PolicyType{
		ID:          policyTypeID,
		Name:        fmt.Sprintf("Policy Type %s", policyTypeID),
		Description: "User-defined policy type",
		Schema:      schema,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	if err := a1.repository.CreatePolicyType(policyType); err != nil {
		statusCode := http.StatusConflict
		if err != ErrAlreadyExists {
			statusCode = http.StatusInternalServerError
		}
		
		a1.writeErrorResponse(w, statusCode, "creation_failed",
			"Failed to create policy type", err)
		a1.metrics.RequestsTotal.WithLabelValues("PUT", "/policytypes/*", fmt.Sprintf("%d", statusCode)).Inc()
		return
	}
	
	w.WriteHeader(http.StatusCreated)
	a1.metrics.RequestsTotal.WithLabelValues("PUT", "/policytypes/*", "201").Inc()
	a1.metrics.PolicyTypesTotal.WithLabelValues("created").Inc()
}

func (a1 *A1Interface) handleDeletePolicyType(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]
	
	defer func() {
		a1.metrics.RequestDuration.WithLabelValues("DELETE", "/policytypes/*").Observe(time.Since(start).Seconds())
	}()
	
	// Check if there are existing policies of this type
	policies, err := a1.repository.ListPoliciesByType(policyTypeID)
	if err != nil && err != ErrNotFound {
		a1.writeErrorResponse(w, http.StatusInternalServerError, "internal_error",
			"Failed to check existing policies", err)
		a1.metrics.RequestsTotal.WithLabelValues("DELETE", "/policytypes/*", "500").Inc()
		return
	}
	
	if len(policies) > 0 {
		a1.writeErrorResponse(w, http.StatusConflict, "conflict",
			"Cannot delete policy type with existing policies", nil)
		a1.metrics.RequestsTotal.WithLabelValues("DELETE", "/policytypes/*", "409").Inc()
		return
	}
	
	if err := a1.repository.DeletePolicyType(policyTypeID); err != nil {
		statusCode := http.StatusNotFound
		if err != ErrNotFound {
			statusCode = http.StatusInternalServerError
		}
		
		a1.writeErrorResponse(w, statusCode, "deletion_failed",
			fmt.Sprintf("Failed to delete policy type %s", policyTypeID), err)
		a1.metrics.RequestsTotal.WithLabelValues("DELETE", "/policytypes/*", fmt.Sprintf("%d", statusCode)).Inc()
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
	a1.metrics.RequestsTotal.WithLabelValues("DELETE", "/policytypes/*", "204").Inc()
	a1.metrics.PolicyTypesTotal.WithLabelValues("deleted").Inc()
}

func (a1 *A1Interface) handleGetPolicies(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]
	
	defer func() {
		a1.metrics.RequestDuration.WithLabelValues("GET", "/policies").Observe(time.Since(start).Seconds())
	}()
	
	policies, err := a1.repository.ListPoliciesByType(policyTypeID)
	if err != nil && err != ErrNotFound {
		a1.writeErrorResponse(w, http.StatusInternalServerError, "internal_error",
			"Failed to retrieve policies", err)
		a1.metrics.RequestsTotal.WithLabelValues("GET", "/policies", "500").Inc()
		return
	}
	
	// Extract policy IDs for O-RAN compliance
	policyIDs := make([]string, len(policies))
	for i, p := range policies {
		policyIDs[i] = p.ID
	}
	
	a1.writeJSONResponse(w, http.StatusOK, policyIDs)
	a1.metrics.RequestsTotal.WithLabelValues("GET", "/policies", "200").Inc()
}

func (a1 *A1Interface) handleCreateOrUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]
	policyID := vars["policy_id"]
	
	defer func() {
		a1.metrics.RequestDuration.WithLabelValues("PUT", "/policies/*").Observe(time.Since(start).Seconds())
	}()
	
	// Get user context from JWT claims
	userContext := r.Context().Value("user").(*Claims)
	
	// Validate policy type exists
	policyType, err := a1.repository.GetPolicyType(policyTypeID)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err == ErrNotFound {
			statusCode = http.StatusBadRequest
		}
		
		a1.writeErrorResponse(w, statusCode, "invalid_policy_type",
			fmt.Sprintf("Policy type %s not found", policyTypeID), err)
		a1.metrics.RequestsTotal.WithLabelValues("PUT", "/policies/*", fmt.Sprintf("%d", statusCode)).Inc()
		return
	}
	
	// Parse policy data
	var policyData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&policyData); err != nil {
		a1.writeErrorResponse(w, http.StatusBadRequest, "bad_request",
			"Invalid JSON in request body", err)
		a1.metrics.RequestsTotal.WithLabelValues("PUT", "/policies/*", "400").Inc()
		return
	}
	
	// Create policy object
	policy := &Policy{
		ID:              policyID,
		TypeID:          policyTypeID,
		Data:            policyData,
		Status:          PolicyStatusNotEnforced,
		CreatedBy:       userContext.UserID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		NotificationURI: r.Header.Get("X-Notification-URI"),
		RICRequestorID:  r.Header.Get("X-RIC-Requestor-ID"),
	}
	
	// Validate policy against schema
	if err := a1.validator.ValidatePolicy(policy, policyType); err != nil {
		a1.writeErrorResponse(w, http.StatusBadRequest, "validation_failed",
			"Policy validation failed", err)
		a1.metrics.RequestsTotal.WithLabelValues("PUT", "/policies/*", "400").Inc()
		return
	}
	
	// Check if policy exists (for status code determination)
	existingPolicy, _ := a1.repository.GetPolicy(policyID)
	isUpdate := existingPolicy != nil
	
	// Create or update policy
	if isUpdate {
		policy.CreatedAt = existingPolicy.CreatedAt
		policy.CreatedBy = existingPolicy.CreatedBy
		if err := a1.repository.UpdatePolicy(policy); err != nil {
			a1.writeErrorResponse(w, http.StatusInternalServerError, "update_failed",
				"Failed to update policy", err)
			a1.metrics.RequestsTotal.WithLabelValues("PUT", "/policies/*", "500").Inc()
			return
		}
	} else {
		if err := a1.repository.CreatePolicy(policy); err != nil {
			statusCode := http.StatusInternalServerError
			if err == ErrAlreadyExists {
				statusCode = http.StatusConflict
			}
			
			a1.writeErrorResponse(w, statusCode, "creation_failed",
				"Failed to create policy", err)
			a1.metrics.RequestsTotal.WithLabelValues("PUT", "/policies/*", fmt.Sprintf("%d", statusCode)).Inc()
			return
		}
	}
	
	// Send notification if enabled
	if a1.notifier != nil {
		notificationType := NotificationPolicyCreated
		if isUpdate {
			notificationType = NotificationPolicyUpdated
		}
		
		notification := &PolicyNotification{
			ID:           fmt.Sprintf("%s-%d", policyID, time.Now().Unix()),
			PolicyID:     policyID,
			PolicyTypeID: policyTypeID,
			Type:         notificationType,
			Timestamp:    time.Now(),
			Data:         policyData,
			UserID:       userContext.UserID,
		}
		
		go a1.notifier.SendNotification(notification)
	}
	
	statusCode := http.StatusCreated
	if isUpdate {
		statusCode = http.StatusOK
	}
	
	w.WriteHeader(statusCode)
	a1.metrics.RequestsTotal.WithLabelValues("PUT", "/policies/*", fmt.Sprintf("%d", statusCode)).Inc()
	
	if isUpdate {
		a1.metrics.PoliciesTotal.WithLabelValues("updated").Inc()
	} else {
		a1.metrics.PoliciesTotal.WithLabelValues("created").Inc()
	}
}

// Authentication handlers

func (a1 *A1Interface) handleLogin(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		a1.metrics.RequestDuration.WithLabelValues("POST", "/auth/login").Observe(time.Since(start).Seconds())
	}()
	
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		a1.writeErrorResponse(w, http.StatusBadRequest, "bad_request",
			"Invalid JSON in request body", err)
		a1.metrics.RequestsTotal.WithLabelValues("POST", "/auth/login", "400").Inc()
		return
	}
	
	// Validate credentials (this would typically check against a database)
	user, err := a1.validateCredentials(loginReq.Username, loginReq.Password)
	if err != nil {
		a1.writeErrorResponse(w, http.StatusUnauthorized, "invalid_credentials",
			"Invalid username or password", err)
		a1.metrics.RequestsTotal.WithLabelValues("POST", "/auth/login", "401").Inc()
		return
	}
	
	// Generate JWT tokens
	accessToken, err := a1.authenticator.GenerateToken(user)
	if err != nil {
		a1.writeErrorResponse(w, http.StatusInternalServerError, "token_generation_failed",
			"Failed to generate access token", err)
		a1.metrics.RequestsTotal.WithLabelValues("POST", "/auth/login", "500").Inc()
		return
	}
	
	refreshToken, err := a1.authenticator.GenerateRefreshToken(user)
	if err != nil {
		a1.writeErrorResponse(w, http.StatusInternalServerError, "token_generation_failed",
			"Failed to generate refresh token", err)
		a1.metrics.RequestsTotal.WithLabelValues("POST", "/auth/login", "500").Inc()
		return
	}
	
	response := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    int(a1.authenticator.tokenExpiry.Seconds()),
		"user": map[string]interface{}{
			"user_id":  user.UserID,
			"username": user.Username,
			"roles":    user.Roles,
		},
	}
	
	a1.writeJSONResponse(w, http.StatusOK, response)
	a1.metrics.RequestsTotal.WithLabelValues("POST", "/auth/login", "200").Inc()
	a1.metrics.AuthenticationTotal.WithLabelValues("login", "success").Inc()
}

// Middleware functions

func (a1 *A1Interface) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)
		
		duration := time.Since(start)
		
		a1.logger.WithFields(logrus.Fields{
			"method":      r.Method,
			"uri":         r.RequestURI,
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
			"status_code": wrapped.statusCode,
			"duration":    duration,
		}).Info("A1 interface request processed")
		
		// Update metrics
		a1.metrics.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
	})
}

func (a1 *A1Interface) authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a1.writeErrorResponse(w, http.StatusUnauthorized, "missing_authorization",
				"Authorization header is required", nil)
			a1.metrics.AuthenticationTotal.WithLabelValues("middleware", "missing_header").Inc()
			return
		}
		
		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			a1.writeErrorResponse(w, http.StatusUnauthorized, "invalid_authorization_format",
				"Authorization header must be 'Bearer <token>'", nil)
			a1.metrics.AuthenticationTotal.WithLabelValues("middleware", "invalid_format").Inc()
			return
		}
		
		token := parts[1]
		
		// Validate JWT token
		claims, err := a1.authenticator.ValidateToken(token)
		if err != nil {
			a1.writeErrorResponse(w, http.StatusUnauthorized, "invalid_token",
				"Invalid or expired token", err)
			a1.metrics.AuthenticationTotal.WithLabelValues("middleware", "invalid_token").Inc()
			return
		}
		
		// Add user context to request
		ctx := context.WithValue(r.Context(), "user", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a1 *A1Interface) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a1.config.RateLimit.Enabled {
			next.ServeHTTP(w, r)
			return
		}
		
		if !a1.rateLimiter.Allow() {
			a1.writeErrorResponse(w, http.StatusTooManyRequests, "rate_limit_exceeded",
				"Rate limit exceeded", nil)
			a1.metrics.RateLimitTotal.WithLabelValues("exceeded").Inc()
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (a1 *A1Interface) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a1.config.Security.CORS.Enabled {
			next.ServeHTTP(w, r)
			return
		}
		
		// Set CORS headers
		origin := r.Header.Get("Origin")
		if a1.isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(a1.config.Security.CORS.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(a1.config.Security.CORS.AllowedHeaders, ", "))
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(a1.config.Security.CORS.ExposedHeaders, ", "))
		
		if a1.config.Security.CORS.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", a1.config.Security.CORS.MaxAge))
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

func (a1 *A1Interface) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		
		next.ServeHTTP(w, r)
	})
}

// Helper methods

func (a1 *A1Interface) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (a1 *A1Interface) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string, err error) {
	a1.logger.WithFields(logrus.Fields{
		"error_code": errorCode,
		"message":    message,
		"error":      err,
		"status":     statusCode,
	}).Error("A1 interface error")
	
	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    errorCode,
			"message": message,
		},
		"timestamp": time.Now().ISO8601(),
	}
	
	if err != nil {
		errorResp["error"].(map[string]interface{})["details"] = err.Error()
	}
	
	a1.writeJSONResponse(w, statusCode, errorResp)
}

// GenerateToken generates a new JWT access token
func (auth *JWTAuthenticator) GenerateToken(user *User) (string, error) {
	now := time.Now()
	
	claims := &Claims{
		UserID:      user.UserID,
		Username:    user.Username,
		Roles:       user.Roles,
		Permissions: user.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    auth.issuer,
			Subject:   user.UserID,
			Audience:  auth.audience,
			ExpiresAt: jwt.NewNumericDate(now.Add(auth.tokenExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(auth.secretKey)
}

// ValidateToken validates a JWT token and returns the claims
func (auth *JWTAuthenticator) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return auth.secretKey, nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, fmt.Errorf("invalid token")
}

// Additional supporting types and methods would be implemented...

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Initialize standard policy types
func (a1 *A1Interface) initializeStandardPolicyTypes() error {
	// Implementation would load O-RAN standard policy types
	return nil
}

// Placeholder methods for missing dependencies
func (a1 *A1Interface) validateCredentials(username, password string) (*User, error) {
	// This would typically validate against a database
	return &User{
		UserID:   "user123",
		Username: username,
		Roles:    []string{"admin"},
		Permissions: []string{"policy:read", "policy:write"},
	}, nil
}

func (a1 *A1Interface) isAllowedOrigin(origin string) bool {
	if len(a1.config.Security.CORS.AllowedOrigins) == 0 {
		return true
	}
	for _, allowed := range a1.config.Security.CORS.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func (a1 *A1Interface) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Prometheus metrics endpoint would be handled here
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("# Metrics endpoint"))
}

// Additional handlers would be implemented for remaining endpoints...
func (a1 *A1Interface) handleGetPolicy(w http.ResponseWriter, r *http.Request) {}
func (a1 *A1Interface) handleDeletePolicy(w http.ResponseWriter, r *http.Request) {}
func (a1 *A1Interface) handleGetPolicyStatus(w http.ResponseWriter, r *http.Request) {}
func (a1 *A1Interface) handleDataDelivery(w http.ResponseWriter, r *http.Request) {}
func (a1 *A1Interface) handleRefreshToken(w http.ResponseWriter, r *http.Request) {}
func (a1 *A1Interface) handleLogout(w http.ResponseWriter, r *http.Request) {}