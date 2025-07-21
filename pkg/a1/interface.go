package a1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// A1Interface implements the O-RAN A1 interface for policy management
type A1Interface struct {
	config              *A1InterfaceConfig
	repository          A1Repository
	notificationService A1NotificationService
	validator           A1PolicyValidator
	router              *mux.Router
	server              *http.Server
	logger              *logrus.Logger
	ctx                 context.Context
	cancel              context.CancelFunc
	running             bool
	mutex               sync.RWMutex
	
	// Metrics
	requestCount        int64
	policyCount         int64
	policyTypeCount     int64
	lastRequestTime     time.Time
}

// NewA1Interface creates a new A1Interface with O-RAN compliance
func NewA1Interface(config *A1InterfaceConfig, repo A1Repository) *A1Interface {
	ctx, cancel := context.WithCancel(context.Background())
	
	logger := logrus.WithField("component", "a1-interface").Logger
	
	a1 := &A1Interface{
		config:              config,
		repository:          repo,
		validator:           NewA1PolicyValidator(),
		logger:              logger,
		ctx:                 ctx,
		cancel:              cancel,
		running:             false,
		lastRequestTime:     time.Now(),
	}
	
	// Initialize router
	a1.setupRoutes()
	
	// Initialize notification service if enabled
	if config.NotificationEnabled {
		a1.notificationService = NewA1NotificationService()
	}
	
	return a1
}

// Start starts the A1 interface HTTP server
func (a1 *A1Interface) Start() error {
	if a1.running {
		return fmt.Errorf("A1 interface is already running")
	}

	a1.logger.WithFields(logrus.Fields{
		"listen_address": a1.config.ListenAddress,
		"listen_port":    a1.config.ListenPort,
		"tls_enabled":    a1.config.TLSEnabled,
	}).Info("Starting O-RAN A1 interface")

	// Initialize standard policy types
	if err := a1.initializeStandardPolicyTypes(); err != nil {
		return fmt.Errorf("failed to initialize standard policy types: %w", err)
	}

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", a1.config.ListenAddress, a1.config.ListenPort)
	a1.server = &http.Server{
		Addr:         addr,
		Handler:      a1.router,
		ReadTimeout:  a1.config.RequestTimeout,
		WriteTimeout: a1.config.RequestTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		var err error
		if a1.config.TLSEnabled {
			err = a1.server.ListenAndServeTLS(a1.config.TLSCertPath, a1.config.TLSKeyPath)
		} else {
			err = a1.server.ListenAndServe()
		}
		
		if err != nil && err != http.ErrServerClosed {
			a1.logger.WithError(err).Error("A1 interface server error")
		}
	}()

	a1.running = true
	a1.logger.Info("O-RAN A1 interface started successfully")

	return nil
}

// Stop stops the A1 interface HTTP server
func (a1 *A1Interface) Stop() error {
	if !a1.running {
		return nil
	}

	a1.logger.Info("Stopping O-RAN A1 interface")
	
	// Cancel context
	a1.cancel()
	
	// Shutdown HTTP server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := a1.server.Shutdown(shutdownCtx); err != nil {
		a1.logger.WithError(err).Error("Error shutting down A1 interface server")
		return err
	}

	a1.running = false
	a1.logger.Info("O-RAN A1 interface stopped successfully")

	return nil
}

// GetStatus returns the current status of the A1 interface
func (a1 *A1Interface) GetStatus() *A1HealthStatus {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	return &A1HealthStatus{
		Status:         "OK",
		LastUpdate:     time.Now(),
		Version:        "1.0.0",
		PolicyTypes:    int(a1.policyTypeCount),
		Policies:       int(a1.policyCount),
		ActivePolicies: int(a1.policyCount), // Simplified - would count only active policies
	}
}

// Setup HTTP routes according to O-RAN A1 API specification
func (a1 *A1Interface) setupRoutes() {
	a1.router = mux.NewRouter()
	
	// Add middleware
	a1.router.Use(a1.loggingMiddleware)
	a1.router.Use(a1.authenticationMiddleware)
	if a1.config.RateLimitEnabled {
		a1.router.Use(a1.rateLimitMiddleware)
	}

	// Health check endpoint
	a1.router.HandleFunc("/a1-p/healthcheck", a1.handleHealthCheck).Methods("GET")

	// Policy type management endpoints
	a1.router.HandleFunc("/a1-p/policytypes", a1.handleGetPolicyTypes).Methods("GET")
	a1.router.HandleFunc("/a1-p/policytypes/{policy_type_id}", a1.handleGetPolicyType).Methods("GET")
	a1.router.HandleFunc("/a1-p/policytypes/{policy_type_id}", a1.handleCreatePolicyType).Methods("PUT")
	a1.router.HandleFunc("/a1-p/policytypes/{policy_type_id}", a1.handleDeletePolicyType).Methods("DELETE")

	// Policy management endpoints
	a1.router.HandleFunc("/a1-p/policytypes/{policy_type_id}/policies", a1.handleGetPolicies).Methods("GET")
	a1.router.HandleFunc("/a1-p/policytypes/{policy_type_id}/policies/{policy_id}", a1.handleGetPolicy).Methods("GET")
	a1.router.HandleFunc("/a1-p/policytypes/{policy_type_id}/policies/{policy_id}", a1.handleCreateOrUpdatePolicy).Methods("PUT")
	a1.router.HandleFunc("/a1-p/policytypes/{policy_type_id}/policies/{policy_id}", a1.handleDeletePolicy).Methods("DELETE")

	// Policy status endpoints
	a1.router.HandleFunc("/a1-p/policytypes/{policy_type_id}/policies/{policy_id}/status", a1.handleGetPolicyStatus).Methods("GET")

	// Data delivery endpoints (for xApp notifications)
	a1.router.HandleFunc("/a1-p/data-delivery", a1.handleDataDelivery).Methods("POST")
}

// HTTP Handlers

func (a1 *A1Interface) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	status := a1.GetStatus()
	a1.writeJSONResponse(w, http.StatusOK, status)
}

func (a1 *A1Interface) handleGetPolicyTypes(w http.ResponseWriter, r *http.Request) {
	policyTypes, err := a1.repository.ListPolicyTypes()
	if err != nil {
		a1.writeErrorResponse(w, http.StatusInternalServerError, A1ErrorInternalError, 
			"Failed to retrieve policy types", err)
		return
	}

	// Extract policy type IDs
	policyTypeIDs := make([]string, len(policyTypes))
	for i, pt := range policyTypes {
		policyTypeIDs[i] = pt.PolicyTypeID
	}

	a1.writeJSONResponse(w, http.StatusOK, policyTypeIDs)
}

func (a1 *A1Interface) handleGetPolicyType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]

	policyType, err := a1.repository.GetPolicyType(policyTypeID)
	if err != nil {
		a1.writeErrorResponse(w, http.StatusNotFound, A1ErrorNotFound, 
			fmt.Sprintf("Policy type %s not found", policyTypeID), err)
		return
	}

	a1.writeJSONResponse(w, http.StatusOK, policyType.PolicySchema)
}

func (a1 *A1Interface) handleCreatePolicyType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]

	var schema map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&schema); err != nil {
		a1.writeErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			"Invalid JSON in request body", err)
		return
	}

	// Validate schema
	if err := a1.validator.ValidateSchema(schema); err != nil {
		a1.writeErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			"Invalid policy type schema", err)
		return
	}

	// Create policy type
	policyType := &A1PolicyType{
		PolicyTypeID:   policyTypeID,
		Name:           fmt.Sprintf("Policy Type %s", policyTypeID),
		Description:    "User-defined policy type",
		PolicySchema:   schema,
		CreatedAt:      time.Now(),
		LastModified:   time.Now(),
	}

	if err := a1.repository.CreatePolicyType(policyType); err != nil {
		a1.writeErrorResponse(w, http.StatusConflict, A1ErrorConflict,
			"Policy type already exists", err)
		return
	}

	a1.mutex.Lock()
	a1.policyTypeCount++
	a1.mutex.Unlock()

	w.WriteHeader(http.StatusCreated)
}

func (a1 *A1Interface) handleDeletePolicyType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]

	// Check if there are existing policies of this type
	policies, err := a1.repository.ListPoliciesByType(policyTypeID)
	if err != nil {
		a1.writeErrorResponse(w, http.StatusInternalServerError, A1ErrorInternalError,
			"Failed to check existing policies", err)
		return
	}

	if len(policies) > 0 {
		a1.writeErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			"Cannot delete policy type with existing policies", nil)
		return
	}

	if err := a1.repository.DeletePolicyType(policyTypeID); err != nil {
		a1.writeErrorResponse(w, http.StatusNotFound, A1ErrorNotFound,
			fmt.Sprintf("Policy type %s not found", policyTypeID), err)
		return
	}

	a1.mutex.Lock()
	a1.policyTypeCount--
	a1.mutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

func (a1 *A1Interface) handleGetPolicies(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]

	policies, err := a1.repository.ListPoliciesByType(policyTypeID)
	if err != nil {
		a1.writeErrorResponse(w, http.StatusInternalServerError, A1ErrorInternalError,
			"Failed to retrieve policies", err)
		return
	}

	// Extract policy IDs
	policyIDs := make([]string, len(policies))
	for i, p := range policies {
		policyIDs[i] = p.PolicyID
	}

	a1.writeJSONResponse(w, http.StatusOK, policyIDs)
}

func (a1 *A1Interface) handleGetPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID := vars["policy_id"]

	policy, err := a1.repository.GetPolicy(policyID)
	if err != nil {
		a1.writeErrorResponse(w, http.StatusNotFound, A1ErrorNotFound,
			fmt.Sprintf("Policy %s not found", policyID), err)
		return
	}

	a1.writeJSONResponse(w, http.StatusOK, policy.PolicyData)
}

func (a1 *A1Interface) handleCreateOrUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]
	policyID := vars["policy_id"]

	// Validate policy type exists
	policyType, err := a1.repository.GetPolicyType(policyTypeID)
	if err != nil {
		a1.writeErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			fmt.Sprintf("Policy type %s not found", policyTypeID), err)
		return
	}

	var policyData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&policyData); err != nil {
		a1.writeErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			"Invalid JSON in request body", err)
		return
	}

	// Create policy object
	policy := &A1Policy{
		PolicyID:        policyID,
		PolicyTypeID:    policyTypeID,
		RICRequestorID:  r.Header.Get("X-RIC-Requestor-ID"),
		PolicyData:      policyData,
		Status:          A1PolicyStatusNotEnforced,
		CreatedAt:       time.Now(),
		LastModified:    time.Now(),
		NotificationURI: r.Header.Get("X-Notification-URI"),
	}

	// Validate policy against schema
	if err := a1.validator.ValidatePolicy(policy, policyType); err != nil {
		a1.writeErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			"Policy validation failed", err)
		return
	}

	// Check if policy exists (for status code determination)
	existingPolicy, _ := a1.repository.GetPolicy(policyID)
	isUpdate := existingPolicy != nil

	// Create or update policy
	if isUpdate {
		if err := a1.repository.UpdatePolicy(policy); err != nil {
			a1.writeErrorResponse(w, http.StatusInternalServerError, A1ErrorInternalError,
				"Failed to update policy", err)
			return
		}
	} else {
		if err := a1.repository.CreatePolicy(policy); err != nil {
			a1.writeErrorResponse(w, http.StatusInternalServerError, A1ErrorInternalError,
				"Failed to create policy", err)
			return
		}
		
		a1.mutex.Lock()
		a1.policyCount++
		a1.mutex.Unlock()
	}

	// Send notification if enabled
	if a1.notificationService != nil {
		notificationType := A1NotificationPolicyCreated
		if isUpdate {
			notificationType = A1NotificationPolicyUpdated
		}
		
		notification := &A1Notification{
			NotificationID:   fmt.Sprintf("%s-%d", policyID, time.Now().Unix()),
			PolicyID:         policyID,
			PolicyTypeID:     policyTypeID,
			NotificationType: notificationType,
			Timestamp:        time.Now(),
			Data:             policyData,
		}
		
		go a1.notificationService.SendNotification(notification)
	}

	if isUpdate {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func (a1 *A1Interface) handleDeletePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID := vars["policy_id"]

	if err := a1.repository.DeletePolicy(policyID); err != nil {
		a1.writeErrorResponse(w, http.StatusNotFound, A1ErrorNotFound,
			fmt.Sprintf("Policy %s not found", policyID), err)
		return
	}

	a1.mutex.Lock()
	a1.policyCount--
	a1.mutex.Unlock()

	// Send notification if enabled
	if a1.notificationService != nil {
		notification := &A1Notification{
			NotificationID:   fmt.Sprintf("%s-delete-%d", policyID, time.Now().Unix()),
			PolicyID:         policyID,
			NotificationType: A1NotificationPolicyDeleted,
			Timestamp:        time.Now(),
		}
		
		go a1.notificationService.SendNotification(notification)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a1 *A1Interface) handleGetPolicyStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID := vars["policy_id"]

	policy, err := a1.repository.GetPolicy(policyID)
	if err != nil {
		a1.writeErrorResponse(w, http.StatusNotFound, A1ErrorNotFound,
			fmt.Sprintf("Policy %s not found", policyID), err)
		return
	}

	status := map[string]interface{}{
		"policy_id":       policy.PolicyID,
		"policy_type_id":  policy.PolicyTypeID,
		"status":          policy.Status,
		"last_modified":   policy.LastModified,
	}

	// Get enforcement details if available
	if enforcement, err := a1.repository.GetEnforcementStatus(policyID); err == nil {
		status["enforcement"] = enforcement
	}

	a1.writeJSONResponse(w, http.StatusOK, status)
}

func (a1 *A1Interface) handleDataDelivery(w http.ResponseWriter, r *http.Request) {
	// Handle data delivery from xApps (policy enforcement status updates)
	var deliveryData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&deliveryData); err != nil {
		a1.writeErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			"Invalid JSON in request body", err)
		return
	}

	a1.logger.WithField("data", deliveryData).Info("Received data delivery from xApp")

	// Process the delivery data (would implement specific logic here)
	w.WriteHeader(http.StatusOK)
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
	}).Error("A1 interface error")

	errorResp := &A1ErrorResponse{
		ErrorCode:    errorCode,
		ErrorMessage: message,
		Timestamp:    time.Now(),
	}

	if err != nil {
		errorResp.Details = map[string]interface{}{
			"underlying_error": err.Error(),
		}
	}

	a1.writeJSONResponse(w, statusCode, errorResp)
}

func (a1 *A1Interface) initializeStandardPolicyTypes() error {
	standardTypes := GetStandardPolicyTypes()
	
	for _, policyType := range standardTypes {
		// Check if already exists
		if _, err := a1.repository.GetPolicyType(policyType.PolicyTypeID); err == nil {
			continue // Already exists
		}
		
		if err := a1.repository.CreatePolicyType(policyType); err != nil {
			return fmt.Errorf("failed to create standard policy type %s: %w", policyType.PolicyTypeID, err)
		}
		
		a1.logger.WithField("policy_type_id", policyType.PolicyTypeID).Info("Created standard policy type")
	}
	
	return nil
}

// Middleware

func (a1 *A1Interface) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Update metrics
		a1.mutex.Lock()
		a1.requestCount++
		a1.lastRequestTime = start
		a1.mutex.Unlock()
		
		next.ServeHTTP(w, r)
		
		a1.logger.WithFields(logrus.Fields{
			"method":      r.Method,
			"uri":         r.RequestURI,
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
			"duration":    time.Since(start),
		}).Info("A1 interface request processed")
	})
}

func (a1 *A1Interface) authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a1.config.AuthenticationEnabled {
			next.ServeHTTP(w, r)
			return
		}

		// Skip authentication for health check
		if r.RequestURI == "/a1-p/healthcheck" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a1.writeErrorResponse(w, http.StatusUnauthorized, A1ErrorUnauthorized,
				"Authorization header required", nil)
			return
		}

		// Validate token (simplified - would implement proper JWT validation)
		// For now, accept any Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			a1.writeErrorResponse(w, http.StatusUnauthorized, A1ErrorUnauthorized,
				"Invalid authorization format", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a1 *A1Interface) rateLimitMiddleware(next http.Handler) http.Handler {
	// Simplified rate limiting - would implement proper rate limiter
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}