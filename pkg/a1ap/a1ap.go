package a1ap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/hctsai1006/near-rt-ric/pkg/auth"
	"github.com/sirupsen/logrus"
)

// A1 Interface Types according to O-RAN A1AP specification

// A1PolicyType represents a policy type definition
type A1PolicyType struct {
	PolicyTypeID string                 `json:"policy_type_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	PolicySchema map[string]interface{} `json:"policy_schema"`
	CreatedAt    time.Time              `json:"created_at"`
	LastModified time.Time              `json:"last_modified"`
}

// A1Policy represents an A1 policy instance
type A1Policy struct {
	PolicyID        string                 `json:"policy_id"`
	PolicyTypeID    string                 `json:"policy_type_id"`
	RICRequestorID  string                 `json:"ric_requestor_id"`
	PolicyData      map[string]interface{} `json:"policy_data"`
	Status          A1PolicyStatus         `json:"status"`
	CreatedAt       time.Time              `json:"created_at"`
	LastModified    time.Time              `json:"last_modified"`
	NotificationURI string                 `json:"notification_uri,omitempty"`
}

// A1PolicyStatus represents the status of an A1 policy
type A1PolicyStatus string

const (
	A1PolicyStatusNotEnforced A1PolicyStatus = "NOT_ENFORCED"
	A1PolicyStatusEnforced    A1PolicyStatus = "ENFORCED"
	A1PolicyStatusDeleted     A1PolicyStatus = "DELETED"
	A1PolicyStatusError       A1PolicyStatus = "ERROR"
)

// A1PolicyTypeStatus represents the status of a policy type
type A1PolicyTypeStatus struct {
	PolicyTypeID      string   `json:"policy_type_id"`
	NumberOfPolicies  int      `json:"num_instances"`
	PolicyInstanceIDs []string `json:"policy_instance_list"`
}

// A1HealthStatus represents the health status of the A1 interface
type A1HealthStatus struct {
	Status         string    `json:"status"`
	LastUpdate     time.Time `json:"last_update"`
	Version        string    `json:"version"`
	PolicyTypes    int       `json:"policy_types_count"`
	Policies       int       `json:"policies_count"`
	ActivePolicies int       `json:"active_policies_count"`
}

// A1Notification represents a notification message
type A1Notification struct {
	NotificationID   string                 `json:"notification_id"`
	PolicyID         string                 `json:"policy_id"`
	PolicyTypeID     string                 `json:"policy_type_id"`
	NotificationType A1NotificationType     `json:"notification_type"`
	Timestamp        time.Time              `json:"timestamp"`
	Data             map[string]interface{} `json:"data,omitempty"`
}

// A1NotificationType represents the type of notification
type A1NotificationType string

const (
	A1NotificationPolicyCreated  A1NotificationType = "POLICY_CREATED"
	A1NotificationPolicyUpdated  A1NotificationType = "POLICY_UPDATED"
	A1NotificationPolicyDeleted  A1NotificationType = "POLICY_DELETED"
	A1NotificationPolicyEnforced A1NotificationType = "POLICY_ENFORCED"
	A1NotificationPolicyError    A1NotificationType = "POLICY_ERROR"
)

// A1ErrorResponse represents an error response from A1 interface
type A1ErrorResponse struct {
	ErrorCode    string                 `json:"error_code"`
	ErrorMessage string                 `json:"error_message"`
	Timestamp    time.Time              `json:"timestamp"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// Common error codes
const (
	A1ErrorNotFound           = "NOT_FOUND"
	A1ErrorBadRequest         = "BAD_REQUEST"
	A1ErrorInternalError      = "INTERNAL_ERROR"
	A1ErrorConflict           = "CONFLICT"
	A1ErrorUnauthorized       = "UNAUTHORIZED"
	A1ErrorServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// A1Request represents a generic A1 request
type A1Request struct {
	RequestID   string                 `json:"request_id"`
	Method      string                 `json:"method"`
	URI         string                 `json:"uri"`
	Headers     map[string]string      `json:"headers"`
	Body        map[string]interface{} `json:"body,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestorID string                 `json:"requestor_id"`
}

// A1Response represents a generic A1 response
type A1Response struct {
	RequestID      string                 `json:"request_id"`
	StatusCode     int                    `json:"status_code"`
	Headers        map[string]string      `json:"headers"`
	Body           map[string]interface{} `json:"body,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	ProcessingTime time.Duration          `json:"processing_time"`
}

// A1ServiceModel represents service model information
type A1ServiceModel struct {
	ServiceModelOID     string `json:"service_model_oid"`
	ServiceModelName    string `json:"service_model_name"`
	ServiceModelVersion string `json:"service_model_version"`
	Description         string `json:"description"`
}

// A1InterfaceConfig contains configuration for the A1 interface
type A1InterfaceConfig struct {
	ListenAddress         string        `json:"listen_address"`
	ListenPort            int           `json:"listen_port"`
	TLSEnabled            bool          `json:"tls_enabled"`
	TLSCertPath           string        `json:"tls_cert_path"`
	TLSKeyPath            string        `json:"tls_key_path"`
	AuthenticationEnabled bool          `json:"authentication_enabled"`
	RequestTimeout        time.Duration `json:"request_timeout"`
	MaxRequestSize        int64         `json:"max_request_size"`
	RateLimitEnabled      bool          `json:"rate_limit_enabled"`
	RateLimitPerMinute    int           `json:"rate_limit_per_minute"`
	NotificationEnabled   bool          `json:"notification_enabled"`
	DatabaseURL           string        `json:"database_url"`
	LogLevel              string        `json:"log_level"`
}

// PolicyEnforcement represents policy enforcement status and details
type PolicyEnforcement struct {
	PolicyID        string                 `json:"policy_id"`
	EnforcementTime time.Time              `json:"enforcement_time"`
	EnforcedBy      []string               `json:"enforced_by"`
	Status          A1PolicyStatus         `json:"status"`
	Details         map[string]interface{} `json:"details,omitempty"`
	Metrics         *PolicyMetrics         `json:"metrics,omitempty"`
}

// PolicyMetrics represents metrics for policy enforcement
type PolicyMetrics struct {
	EnforcementLatency  time.Duration `json:"enforcement_latency"`
	SuccessRate         float64       `json:"success_rate"`
	ErrorRate           float64       `json:"error_rate"`
	LastEnforcementTime time.Time     `json:"last_enforcement_time"`
	TotalEnforcements   int64         `json:"total_enforcements"`
	FailedEnforcements  int64         `json:"failed_enforcements"`
}

// A1Repository interface defines storage operations for A1 policies
type A1Repository interface {
	// Policy Type operations
	CreatePolicyType(policyType *A1PolicyType) error
	GetPolicyType(policyTypeID string) (*A1PolicyType, error)
	UpdatePolicyType(policyType *A1PolicyType) error
	DeletePolicyType(policyTypeID string) error
	ListPolicyTypes() ([]*A1PolicyType, error)

	// Policy operations
	CreatePolicy(policy *A1Policy) error
	GetPolicy(policyID string) (*A1Policy, error)
	UpdatePolicy(policy *A1Policy) error
	DeletePolicy(policyID string) error
	ListPolicies() ([]*A1Policy, error)
	ListPoliciesByType(policyTypeID string) ([]*A1Policy, error)

	// Policy enforcement tracking
	RecordEnforcement(enforcement *PolicyEnforcement) error
	GetEnforcementStatus(policyID string) (*PolicyEnforcement, error)
	GetPolicyMetrics(policyID string) (*PolicyMetrics, error)
}

// A1NotificationService interface defines notification operations
type A1NotificationService interface {
	SendNotification(notification *A1Notification) error
	RegisterWebhook(policyID string, webhookURL string) error
	UnregisterWebhook(policyID string) error
	ListWebhooks() (map[string]string, error)
}

// A1PolicyValidator interface defines policy validation operations
type A1PolicyValidator interface {
	ValidatePolicyType(policyType *A1PolicyType) error
	ValidatePolicy(policy *A1Policy, policyType *A1PolicyType) error
	ValidateSchema(schema map[string]interface{}) error
}

// PolicyTypeRegistry manages policy type definitions and validation
type PolicyTypeRegistry struct {
	policyTypes map[string]*A1PolicyType
	validators  map[string]A1PolicyValidator
}

// Standard O-RAN policy type schemas
var (
	// QoS Policy Type Schema
	QoSPolicyTypeSchema = map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    "object",
		"properties": map[string]interface{}{
			"qci": map[string]interface{}{
				"type":    "integer",
				"minimum": 1,
				"maximum": 9,
			},
			"priority_level": map[string]interface{}{
				"type":    "integer",
				"minimum": 1,
				"maximum": 15,
			},
			"packet_delay_budget": map[string]interface{}{
				"type":    "integer",
				"minimum": 50,
				"maximum": 300,
			},
			"packet_error_loss_rate": map[string]interface{}{
				"type":    "number",
				"minimum": 0.000001,
				"maximum": 0.1,
			},
		},
		"required": []string{"qci", "priority_level"},
	}

	// RRM Policy Type Schema (Radio Resource Management)
	RRMPolicyTypeSchema = map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    "object",
		"properties": map[string]interface{}{
			"resource_type": map[string]interface{}{
				"type": "string",
				"enum": []string{"PRB", "SPECTRUM", "POWER"},
			},
			"allocation_strategy": map[string]interface{}{
				"type": "string",
				"enum": []string{"ROUND_ROBIN", "PROPORTIONAL_FAIR", "MAX_THROUGHPUT"},
			},
			"max_allocation_percentage": map[string]interface{}{
				"type":    "number",
				"minimum": 0,
				"maximum": 100,
			},
			"priority_weights": map[string]interface{}{
				"type": "object",
				"additionalProperties": map[string]interface{}{
					"type":    "number",
					"minimum": 0,
					"maximum": 1,
				},
			},
		},
		"required": []string{"resource_type", "allocation_strategy"},
	}

	// SON Policy Type Schema (Self-Organizing Networks)
	SONPolicyTypeSchema = map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    "object",
		"properties": map[string]interface{}{
			"function_type": map[string]interface{}{
				"type": "string",
				"enum": []string{"MLB", "MRO", "CCO", "ANR"},
			},
			"optimization_target": map[string]interface{}{
				"type": "string",
				"enum": []string{"THROUGHPUT", "LATENCY", "COVERAGE", "CAPACITY"},
			},
			"trigger_conditions": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"threshold_type": map[string]interface{}{
						"type": "string",
						"enum": []string{"ABSOLUTE", "PERCENTAGE"},
					},
					"threshold_value": map[string]interface{}{
						"type": "number",
					},
					"measurement_window": map[string]interface{}{
						"type":    "integer",
						"minimum": 60,
						"maximum": 3600,
					},
				},
			},
		},
		"required": []string{"function_type", "optimization_target"},
	}
)

// Predefined policy types for O-RAN
func GetStandardPolicyTypes() []*A1PolicyType {
	now := time.Now()

	return []*A1PolicyType{
		{
			PolicyTypeID: "o-ran-qos-policy-v1.0",
			Name:         "QoS Management Policy",
			Description:  "Policy for managing Quality of Service parameters",
			PolicySchema: QoSPolicyTypeSchema,
			CreatedAt:    now,
			LastModified: now,
		},
		{
			PolicyTypeID: "o-ran-rrm-policy-v1.0",
			Name:         "Radio Resource Management Policy",
			Description:  "Policy for managing radio resource allocation",
			PolicySchema: RRMPolicyTypeSchema,
			CreatedAt:    now,
			LastModified: now,
		},
		{
			PolicyTypeID: "o-ran-son-policy-v1.0",
			Name:         "Self-Organizing Networks Policy",
			Description:  "Policy for SON function optimization",
			PolicySchema: SONPolicyTypeSchema,
			CreatedAt:    now,
			LastModified: now,
		},
	}
}

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
	requestCount    int64
	policyCount     int64
	policyTypeCount int64
	lastRequestTime time.Time
}

// NewA1Interface creates a new A1Interface with O-RAN compliance
func NewA1Interface(config *A1InterfaceConfig, repo A1Repository) *A1Interface {
	ctx, cancel := context.WithCancel(context.Background())

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	a1 := &A1Interface{
		config:          config,
		repository:      repo,
		validator:       NewA1PolicyValidator(),
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
		running:         false,
		lastRequestTime: time.Now(),
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
	a1.router.Use(auth.AuthMiddleware)
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
	WriteJSONResponse(w, http.StatusOK, status)
}

func (a1 *A1Interface) handleGetPolicyTypes(w http.ResponseWriter, r *http.Request) {
	policyTypes, err := a1.repository.ListPolicyTypes()
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, A1ErrorInternalError,
			"Failed to retrieve policy types", err)
		return
	}

	// Extract policy type IDs
	policyTypeIDs := make([]string, len(policyTypes))
	for i, pt := range policyTypes {
		policyTypeIDs[i] = pt.PolicyTypeID
	}

	WriteJSONResponse(w, http.StatusOK, policyTypeIDs)
}

func (a1 *A1Interface) handleGetPolicyType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]

	policyType, err := a1.repository.GetPolicyType(policyTypeID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, A1ErrorNotFound,
			fmt.Sprintf("Policy type %s not found", policyTypeID), err)
		return
	}

	WriteJSONResponse(w, http.StatusOK, policyType.PolicySchema)
}

func (a1 *A1Interface) handleCreatePolicyType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]

	var schema map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&schema); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			"Invalid JSON in request body", err)
		return
	}

	// Validate schema
	if err := a1.validator.ValidateSchema(schema); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			"Invalid policy type schema", err)
		return
	}

	// Create policy type
	policyType := &A1PolicyType{
		PolicyTypeID: policyTypeID,
		Name:         fmt.Sprintf("Policy Type %s", policyTypeID),
		Description:  "User-defined policy type",
		PolicySchema: schema,
		CreatedAt:    time.Now(),
		LastModified: time.Now(),
	}

	if err := a1.repository.CreatePolicyType(policyType); err != nil {
		WriteErrorResponse(w, http.StatusConflict, A1ErrorConflict,
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
		WriteErrorResponse(w, http.StatusInternalServerError, A1ErrorInternalError,
			"Failed to check existing policies", err)
		return
	}

	if len(policies) > 0 {
		WriteErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			"Cannot delete policy type with existing policies", nil)
		return
	}

	if err := a1.repository.DeletePolicyType(policyTypeID); err != nil {
		WriteErrorResponse(w, http.StatusNotFound, A1ErrorNotFound,
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
		WriteErrorResponse(w, http.StatusInternalServerError, A1ErrorInternalError,
			"Failed to retrieve policies", err)
		return
	}

	// Extract policy IDs
	policyIDs := make([]string, len(policies))
	for i, p := range policies {
		policyIDs[i] = p.PolicyID
	}

	WriteJSONResponse(w, http.StatusOK, policyIDs)
}

func (a1 *A1Interface) handleGetPolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyID := vars["policy_id"]

	policy, err := a1.repository.GetPolicy(policyID)
	if err != nil {
		WriteErrorResponse(w, http.StatusNotFound, A1ErrorNotFound,
			fmt.Sprintf("Policy %s not found", policyID), err)
		return
	}

	WriteJSONResponse(w, http.StatusOK, policy.PolicyData)
}

func (a1 *A1Interface) handleCreateOrUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	policyTypeID := vars["policy_type_id"]
	policyID := vars["policy_id"]

	// Validate policy type exists
	policyType, err := a1.repository.GetPolicyType(policyTypeID)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			fmt.Sprintf("Policy type %s not found", policyTypeID), err)
		return
	}

	var policyData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&policyData); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
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
		WriteErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			"Policy validation failed", err)
		return
	}

	// Check if policy exists (for status code determination)
	existingPolicy, _ := a1.repository.GetPolicy(policyID)
	isUpdate := existingPolicy != nil

	// Create or update policy
	if isUpdate {
		if err := a1.repository.UpdatePolicy(policy); err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, A1ErrorInternalError,
				"Failed to update policy", err)
			return
		}
	} else {
		if err := a1.repository.CreatePolicy(policy); err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, A1ErrorInternalError,
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
		WriteErrorResponse(w, http.StatusNotFound, A1ErrorNotFound,
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
		WriteErrorResponse(w, http.StatusNotFound, A1ErrorNotFound,
			fmt.Sprintf("Policy %s not found", policyID), err)
		return
	}

	status := map[string]interface{}{
		"policy_id":      policy.PolicyID,
		"policy_type_id": policy.PolicyTypeID,
		"status":         policy.Status,
		"last_modified":  policy.LastModified,
	}

	// Get enforcement details if available
	if enforcement, err := a1.repository.GetEnforcementStatus(policyID); err == nil {
		status["enforcement"] = enforcement
	}

	WriteJSONResponse(w, http.StatusOK, status)
}

func (a1 *A1Interface) handleDataDelivery(w http.ResponseWriter, r *http.Request) {
	// Handle data delivery from xApps (policy enforcement status updates)
	var deliveryData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&deliveryData); err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, A1ErrorBadRequest,
			"Invalid JSON in request body", err)
		return
	}

	a1.logger.WithField("data", deliveryData).Info("Received data delivery from xApp")

	// Process the delivery data (would implement specific logic here)
	w.WriteHeader(http.StatusOK)
}

// Helper methods

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string, err error) {
	logrus.WithFields(logrus.Fields{
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

	WriteJSONResponse(w, statusCode, errorResp)
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

func (a1 *A1Interface) rateLimitMiddleware(next http.Handler) http.Handler {
	// Simplified rate limiting - would implement proper rate limiter
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// A1NotificationServiceImpl implements the A1 notification service
type A1NotificationServiceImpl struct {
	webhooks      map[string]string // policyID -> webhook URL
	httpClient    *http.Client
	logger        *logrus.Logger
	mutex         sync.RWMutex
	retryAttempts int
	retryDelay    time.Duration
	timeout       time.Duration
}

// NewA1NotificationService creates a new A1 notification service
func NewA1NotificationService() A1NotificationService {
	return &A1NotificationServiceImpl{
		webhooks: make(map[string]string),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:        logrus.WithField("component", "a1-notification").Logger,
		retryAttempts: 3,
		retryDelay:    2 * time.Second,
		timeout:       30 * time.Second,
	}
}

// SendNotification sends a notification to registered webhooks
func (ns *A1NotificationServiceImpl) SendNotification(notification *A1Notification) error {
	ns.mutex.RLock()
	webhookURL, exists := ns.webhooks[notification.PolicyID]
	ns.mutex.RUnlock()

	if !exists {
		// No webhook registered for this policy - log and continue
		ns.logger.WithField("policy_id", notification.PolicyID).Debug("No webhook registered for policy")
		return nil
	}

	// Send notification with retry logic
	return ns.sendNotificationWithRetry(notification, webhookURL)
}

// RegisterWebhook registers a webhook URL for a specific policy
func (ns *A1NotificationServiceImpl) RegisterWebhook(policyID string, webhookURL string) error {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	// Validate webhook URL format
	if webhookURL == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	// Basic URL validation
	if len(webhookURL) < 7 || (webhookURL[:7] != "http://" && webhookURL[:8] != "https://") {
		return fmt.Errorf("webhook URL must start with http:// or https://")
	}

	ns.webhooks[policyID] = webhookURL

	ns.logger.WithFields(logrus.Fields{
		"policy_id":   policyID,
		"webhook_url": webhookURL,
	}).Info("Webhook registered")

	return nil
}

// UnregisterWebhook removes a webhook registration for a specific policy
func (ns *A1NotificationServiceImpl) UnregisterWebhook(policyID string) error {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	if _, exists := ns.webhooks[policyID]; !exists {
		return fmt.Errorf("no webhook registered for policy %s", policyID)
	}

	delete(ns.webhooks, policyID)

	ns.logger.WithField("policy_id", policyID).Info("Webhook unregistered")

	return nil
}

// ListWebhooks returns all registered webhooks
func (ns *A1NotificationServiceImpl) ListWebhooks() (map[string]string, error) {
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()

	// Return a copy to prevent external modifications
	webhooksCopy := make(map[string]string)
	for policyID, webhookURL := range ns.webhooks {
		webhooksCopy[policyID] = webhookURL
	}

	return webhooksCopy, nil
}

// sendNotificationWithRetry sends a notification with retry logic
func (ns *A1NotificationServiceImpl) sendNotificationWithRetry(notification *A1Notification, webhookURL string) error {
	var lastErr error

	for attempt := 0; attempt < ns.retryAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retrying
			time.Sleep(ns.retryDelay * time.Duration(attempt))
		}

		err := ns.sendNotificationHTTP(notification, webhookURL)
		if err == nil {
			ns.logger.WithFields(logrus.Fields{
				"notification_id": notification.NotificationID,
				"policy_id":       notification.PolicyID,
				"webhook_url":     webhookURL,
				"attempt":         attempt + 1,
			}).Info("Notification sent successfully")
			return nil
		}

		lastErr = err
		ns.logger.WithFields(logrus.Fields{
			"notification_id": notification.NotificationID,
			"policy_id":       notification.PolicyID,
			"webhook_url":     webhookURL,
			"attempt":         attempt + 1,
			"error":           err,
		}).Warn("Notification sending failed, will retry")
	}

	ns.logger.WithFields(logrus.Fields{
		"notification_id": notification.NotificationID,
		"policy_id":       notification.PolicyID,
		"webhook_url":     webhookURL,
		"attempts":        ns.retryAttempts,
		"error":           lastErr,
	}).Error("Failed to send notification after all retry attempts")

	return fmt.Errorf("failed to send notification after %d attempts: %w", ns.retryAttempts, lastErr)
}

// sendNotificationHTTP sends a single HTTP notification
func (ns *A1NotificationServiceImpl) sendNotificationHTTP(notification *A1Notification, webhookURL string) error {
	// Create HTTP request body
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Create HTTP request with context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), ns.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "O-RAN-Near-RT-RIC-A1/1.0")
	req.Header.Set("X-A1-Notification-ID", notification.NotificationID)
	req.Header.Set("X-A1-Policy-ID", notification.PolicyID)
	req.Header.Set("X-A1-Notification-Type", string(notification.NotificationType))

	// Send HTTP request
	resp, err := ns.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// GetNotificationStatistics returns statistics about notification sending
func (ns *A1NotificationServiceImpl) GetNotificationStatistics() map[string]interface{} {
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()

	return map[string]interface{}{
		"registered_webhooks": len(ns.webhooks),
		"retry_attempts":      ns.retryAttempts,
		"retry_delay_ms":      ns.retryDelay.Milliseconds(),
		"timeout_ms":          ns.timeout.Milliseconds(),
	}
}

// UpdateConfiguration updates the notification service configuration
func (ns *A1NotificationServiceImpl) UpdateConfiguration(retryAttempts int, retryDelay, timeout time.Duration) {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	ns.retryAttempts = retryAttempts
	ns.retryDelay = retryDelay
	ns.timeout = timeout

	// Update HTTP client timeout
	ns.httpClient.Timeout = timeout

	ns.logger.WithFields(logrus.Fields{
		"retry_attempts": retryAttempts,
		"retry_delay":    retryDelay,
		"timeout":        timeout,
	}).Info("Notification service configuration updated")
}

// TestWebhook tests a webhook URL by sending a test notification
func (ns *A1NotificationServiceImpl) TestWebhook(webhookURL string) error {
	testNotification := &A1Notification{
		NotificationID:   fmt.Sprintf("test-%d", time.Now().Unix()),
		PolicyID:         "test-policy",
		PolicyTypeID:     "test-policy-type",
		NotificationType: A1NotificationPolicyCreated,
		Timestamp:        time.Now(),
		Data: map[string]interface{}{
			"test":    true,
			"message": "This is a test notification from O-RAN Near-RT RIC A1 interface",
		},
	}

	return ns.sendNotificationHTTP(testNotification, webhookURL)
}

// BatchNotificationSender handles sending multiple notifications efficiently
type BatchNotificationSender struct {
	service       *A1NotificationServiceImpl
	batchSize     int
	batchTimeout  time.Duration
	notifications chan *A1Notification
	done          chan struct{}
	logger        *logrus.Logger
}

// NewBatchNotificationSender creates a new batch notification sender
func NewBatchNotificationSender(service *A1NotificationServiceImpl, batchSize int, batchTimeout time.Duration) *BatchNotificationSender {
	return &BatchNotificationSender{
		service:       service,
		batchSize:     batchSize,
		batchTimeout:  batchTimeout,
		notifications: make(chan *A1Notification, batchSize*2),
		done:          make(chan struct{}),
		logger:        logrus.WithField("component", "batch-notification-sender").Logger,
	}
}

// Start starts the batch notification sender
func (bns *BatchNotificationSender) Start() {
	go bns.processNotifications()
}

// Stop stops the batch notification sender
func (bns *BatchNotificationSender) Stop() {
	close(bns.done)
}

// QueueNotification queues a notification for batch sending
func (bns *BatchNotificationSender) QueueNotification(notification *A1Notification) {
	select {
	case bns.notifications <- notification:
	default:
		bns.logger.Warn("Notification queue is full, dropping notification")
	}
}

// processNotifications processes notifications in batches
func (bns *BatchNotificationSender) processNotifications() {
	ticker := time.NewTicker(bns.batchTimeout)
	defer ticker.Stop()

	var batch []*A1Notification

	for {
		select {
		case notification := <-bns.notifications:
			batch = append(batch, notification)

			if len(batch) >= bns.batchSize {
				bns.sendBatch(batch)
				batch = nil
			}

		case <-ticker.C:
			if len(batch) > 0 {
				bns.sendBatch(batch)
				batch = nil
			}

		case <-bns.done:
			// Send remaining notifications before stopping
			if len(batch) > 0 {
				bns.sendBatch(batch)
			}
			return
		}
	}
}

// sendBatch sends a batch of notifications
func (bns *BatchNotificationSender) sendBatch(batch []*A1Notification) {
	bns.logger.WithField("batch_size", len(batch)).Debug("Sending notification batch")

	// Send notifications concurrently
	var wg sync.WaitGroup
	for _, notification := range batch {
		wg.Add(1)
		go func(n *A1Notification) {
			defer wg.Done()
			if err := bns.service.SendNotification(n); err != nil {
				bns.logger.WithFields(logrus.Fields{
					"notification_id": n.NotificationID,
					"error":           err,
				}).Error("Failed to send notification in batch")
			}
		}(notification)
	}

	wg.Wait()
	bns.logger.WithField("batch_size", len(batch)).Debug("Notification batch sent")
}

// NotificationMetrics tracks metrics for notification sending
type NotificationMetrics struct {
	TotalSent            int64     `json:"total_sent"`
	TotalFailed          int64     `json:"total_failed"`
	AverageLatency       float64   `json:"average_latency_ms"`
	LastNotificationTime time.Time `json:"last_notification_time"`
	mutex                sync.RWMutex
}

// NewNotificationMetrics creates a new notification metrics tracker
func NewNotificationMetrics() *NotificationMetrics {
	return &NotificationMetrics{}
}

// RecordSuccess records a successful notification
func (nm *NotificationMetrics) RecordSuccess(latency time.Duration) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	nm.TotalSent++
	nm.LastNotificationTime = time.Now()

	// Update average latency (simple moving average)
	if nm.TotalSent == 1 {
		nm.AverageLatency = float64(latency.Milliseconds())
	} else {
		nm.AverageLatency = (nm.AverageLatency*float64(nm.TotalSent-1) + float64(latency.Milliseconds())) / float64(nm.TotalSent)
	}
}

// RecordFailure records a failed notification
func (nm *NotificationMetrics) RecordFailure() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	nm.TotalFailed++
}

// GetMetrics returns current metrics
func (nm *NotificationMetrics) GetMetrics() map[string]interface{} {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()

	successRate := float64(0)
	if nm.TotalSent+nm.TotalFailed > 0 {
		successRate = float64(nm.TotalSent) / float64(nm.TotalSent+nm.TotalFailed) * 100
	}

	return map[string]interface{}{
		"total_sent":             nm.TotalSent,
		"total_failed":           nm.TotalFailed,
		"success_rate_percent":   successRate,
		"average_latency_ms":     nm.AverageLatency,
		"last_notification_time": nm.LastNotificationTime,
	}
}

// InMemoryA1Repository provides an in-memory implementation of A1Repository
// In production, this would be replaced with a database-backed implementation
type InMemoryA1Repository struct {
	policyTypes  map[string]*A1PolicyType
	policies     map[string]*A1Policy
	enforcements map[string]*PolicyEnforcement
	metrics      map[string]*PolicyMetrics
	mutex        sync.RWMutex
	logger       *logrus.Logger
}

// NewInMemoryA1Repository creates a new in-memory A1 repository
func NewInMemoryA1Repository() A1Repository {
	return &InMemoryA1Repository{
		policyTypes:  make(map[string]*A1PolicyType),
		policies:     make(map[string]*A1Policy),
		enforcements: make(map[string]*PolicyEnforcement),
		metrics:      make(map[string]*PolicyMetrics),
		logger:       logrus.WithField("component", "a1-repository").Logger,
	}
}

// Policy Type operations

func (repo *InMemoryA1Repository) CreatePolicyType(policyType *A1PolicyType) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.policyTypes[policyType.PolicyTypeID]; exists {
		return fmt.Errorf("policy type %s already exists", policyType.PolicyTypeID)
	}

	// Deep copy to prevent external modifications
	policyTypeCopy := *policyType
	policyTypeCopy.CreatedAt = time.Now()
	policyTypeCopy.LastModified = time.Now()

	repo.policyTypes[policyType.PolicyTypeID] = &policyTypeCopy

	repo.logger.WithField("policy_type_id", policyType.PolicyTypeID).Info("Policy type created")
	return nil
}

func (repo *InMemoryA1Repository) GetPolicyType(policyTypeID string) (*A1PolicyType, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	policyType, exists := repo.policyTypes[policyTypeID]
	if !exists {
		return nil, fmt.Errorf("policy type %s not found", policyTypeID)
	}

	// Return a copy to prevent external modifications
	policyTypeCopy := *policyType
	return &policyTypeCopy, nil
}

func (repo *InMemoryA1Repository) UpdatePolicyType(policyType *A1PolicyType) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.policyTypes[policyType.PolicyTypeID]; !exists {
		return fmt.Errorf("policy type %s not found", policyType.PolicyTypeID)
	}

	policyTypeCopy := *policyType
	policyTypeCopy.LastModified = time.Now()

	repo.policyTypes[policyType.PolicyTypeID] = &policyTypeCopy

	repo.logger.WithField("policy_type_id", policyType.PolicyTypeID).Info("Policy type updated")
	return nil
}

func (repo *InMemoryA1Repository) DeletePolicyType(policyTypeID string) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.policyTypes[policyTypeID]; !exists {
		return fmt.Errorf("policy type %s not found", policyTypeID)
	}

	delete(repo.policyTypes, policyTypeID)

	repo.logger.WithField("policy_type_id", policyTypeID).Info("Policy type deleted")
	return nil
}

func (repo *InMemoryA1Repository) ListPolicyTypes() ([]*A1PolicyType, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	policyTypes := make([]*A1PolicyType, 0, len(repo.policyTypes))
	for _, policyType := range repo.policyTypes {
		policyTypeCopy := *policyType
		policyTypes = append(policyTypes, &policyTypeCopy)
	}

	return policyTypes, nil
}

// Policy operations

func (repo *InMemoryA1Repository) CreatePolicy(policy *A1Policy) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.policies[policy.PolicyID]; exists {
		return fmt.Errorf("policy %s already exists", policy.PolicyID)
	}

	// Verify policy type exists
	if _, exists := repo.policyTypes[policy.PolicyTypeID]; !exists {
		return fmt.Errorf("policy type %s not found", policy.PolicyTypeID)
	}

	// Deep copy to prevent external modifications
	policyCopy := *policy
	policyCopy.CreatedAt = time.Now()
	policyCopy.LastModified = time.Now()
	policyCopy.Status = A1PolicyStatusNotEnforced

	repo.policies[policy.PolicyID] = &policyCopy

	// Initialize metrics
	repo.metrics[policy.PolicyID] = &PolicyMetrics{
		EnforcementLatency:  0,
		SuccessRate:         0,
		ErrorRate:           0,
		LastEnforcementTime: time.Time{},
		TotalEnforcements:   0,
		FailedEnforcements:  0,
	}

	repo.logger.WithFields(logrus.Fields{
		"policy_id":      policy.PolicyID,
		"policy_type_id": policy.PolicyTypeID,
	}).Info("Policy created")

	return nil
}

func (repo *InMemoryA1Repository) GetPolicy(policyID string) (*A1Policy, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	policy, exists := repo.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("policy %s not found", policyID)
	}

	// Return a copy to prevent external modifications
	policyCopy := *policy
	return &policyCopy, nil
}

func (repo *InMemoryA1Repository) UpdatePolicy(policy *A1Policy) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	existingPolicy, exists := repo.policies[policy.PolicyID]
	if !exists {
		return fmt.Errorf("policy %s not found", policy.PolicyID)
	}

	// Preserve creation time and update modification time
	policyCopy := *policy
	policyCopy.CreatedAt = existingPolicy.CreatedAt
	policyCopy.LastModified = time.Now()

	repo.policies[policy.PolicyID] = &policyCopy

	repo.logger.WithField("policy_id", policy.PolicyID).Info("Policy updated")
	return nil
}

func (repo *InMemoryA1Repository) DeletePolicy(policyID string) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.policies[policyID]; !exists {
		return fmt.Errorf("policy %s not found", policyID)
	}

	delete(repo.policies, policyID)
	delete(repo.enforcements, policyID)
	delete(repo.metrics, policyID)

	repo.logger.WithField("policy_id", policyID).Info("Policy deleted")
	return nil
}

func (repo *InMemoryA1Repository) ListPolicies() ([]*A1Policy, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	policies := make([]*A1Policy, 0, len(repo.policies))
	for _, policy := range repo.policies {
		policyCopy := *policy
		policies = append(policies, &policyCopy)
	}

	return policies, nil
}

func (repo *InMemoryA1Repository) ListPoliciesByType(policyTypeID string) ([]*A1Policy, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	var policies []*A1Policy
	for _, policy := range repo.policies {
		if policy.PolicyTypeID == policyTypeID {
			policyCopy := *policy
			policies = append(policies, &policyCopy)
		}
	}

	return policies, nil
}

// Policy enforcement tracking

func (repo *InMemoryA1Repository) RecordEnforcement(enforcement *PolicyEnforcement) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	// Verify policy exists
	policy, exists := repo.policies[enforcement.PolicyID]
	if !exists {
		return fmt.Errorf("policy %s not found", enforcement.PolicyID)
	}

	// Update policy status
	policy.Status = enforcement.Status
	policy.LastModified = time.Now()

	// Store enforcement record
	enforcementCopy := *enforcement
	enforcementCopy.EnforcementTime = time.Now()
	repo.enforcements[enforcement.PolicyID] = &enforcementCopy

	// Update metrics
	metrics, exists := repo.metrics[enforcement.PolicyID]
	if !exists {
		metrics = &PolicyMetrics{}
		repo.metrics[enforcement.PolicyID] = metrics
	}

	metrics.TotalEnforcements++
	metrics.LastEnforcementTime = time.Now()

	if enforcement.Status == A1PolicyStatusEnforced {
		metrics.SuccessRate = float64(metrics.TotalEnforcements-metrics.FailedEnforcements) / float64(metrics.TotalEnforcements)
	} else if enforcement.Status == A1PolicyStatusError {
		metrics.FailedEnforcements++
		metrics.SuccessRate = float64(metrics.TotalEnforcements-metrics.FailedEnforcements) / float64(metrics.TotalEnforcements)
	}

	metrics.ErrorRate = float64(metrics.FailedEnforcements) / float64(metrics.TotalEnforcements)

	if enforcement.Metrics != nil {
		metrics.EnforcementLatency = enforcement.Metrics.EnforcementLatency
	}

	repo.logger.WithFields(logrus.Fields{
		"policy_id": enforcement.PolicyID,
		"status":    enforcement.Status,
	}).Info("Policy enforcement recorded")

	return nil
}

func (repo *InMemoryA1Repository) GetEnforcementStatus(policyID string) (*PolicyEnforcement, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	enforcement, exists := repo.enforcements[policyID]
	if !exists {
		return nil, fmt.Errorf("no enforcement record for policy %s", policyID)
	}

	// Return a copy
	enforcementCopy := *enforcement
	return &enforcementCopy, nil
}

func (repo *InMemoryA1Repository) GetPolicyMetrics(policyID string) (*PolicyMetrics, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	metrics, exists := repo.metrics[policyID]
	if !exists {
		return nil, fmt.Errorf("no metrics for policy %s", policyID)
	}

	// Return a copy
	metricsCopy := *metrics
	return &metricsCopy, nil
}

// Additional utility methods

func (repo *InMemoryA1Repository) GetPolicyTypeStatus(policyTypeID string) (*A1PolicyTypeStatus, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	// Verify policy type exists
	if _, exists := repo.policyTypes[policyTypeID]; !exists {
		return nil, fmt.Errorf("policy type %s not found", policyTypeID)
	}

	// Count policies of this type
	var policyInstanceIDs []string
	for _, policy := range repo.policies {
		if policy.PolicyTypeID == policyTypeID {
			policyInstanceIDs = append(policyInstanceIDs, policy.PolicyID)
		}
	}

	return &A1PolicyTypeStatus{
		PolicyTypeID:      policyTypeID,
		NumberOfPolicies:  len(policyInstanceIDs),
		PolicyInstanceIDs: policyInstanceIDs,
	}, nil
}

func (repo *InMemoryA1Repository) GetStatistics() map[string]interface{} {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	// Count policies by status
	statusCounts := make(map[A1PolicyStatus]int)
	for _, policy := range repo.policies {
		statusCounts[policy.Status]++
	}

	// Count policies by type
	typeCounts := make(map[string]int)
	for _, policy := range repo.policies {
		typeCounts[policy.PolicyTypeID]++
	}

	// Calculate average metrics
	var totalLatency time.Duration
	var totalSuccessRate, totalErrorRate float64
	metricsCount := 0

	for _, metrics := range repo.metrics {
		if metrics.TotalEnforcements > 0 {
			totalLatency += metrics.EnforcementLatency
			totalSuccessRate += metrics.SuccessRate
			totalErrorRate += metrics.ErrorRate
			metricsCount++
		}
	}

	avgLatency := time.Duration(0)
	avgSuccessRate := float64(0)
	avgErrorRate := float64(0)

	if metricsCount > 0 {
		avgLatency = totalLatency / time.Duration(metricsCount)
		avgSuccessRate = totalSuccessRate / float64(metricsCount)
		avgErrorRate = totalErrorRate / float64(metricsCount)
	}

	return map[string]interface{}{
		"total_policy_types":   len(repo.policyTypes),
		"total_policies":       len(repo.policies),
		"policies_by_status":   statusCounts,
		"policies_by_type":     typeCounts,
		"average_latency_ms":   avgLatency.Milliseconds(),
		"average_success_rate": avgSuccessRate,
		"average_error_rate":   avgErrorRate,
		"total_enforcements": func() int64 {
			var total int64
			for _, metrics := range repo.metrics {
				total += metrics.TotalEnforcements
			}
			return total
		}(),
	}
}

// Export/Import functions for backup and restore

func (repo *InMemoryA1Repository) Export() ([]byte, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	exportData := struct {
		PolicyTypes  map[string]*A1PolicyType      `json:"policy_types"`
		Policies     map[string]*A1Policy          `json:"policies"`
		Enforcements map[string]*PolicyEnforcement `json:"enforcements"`
		Metrics      map[string]*PolicyMetrics     `json:"metrics"`
		ExportTime   time.Time                     `json:"export_time"`
	}{
		PolicyTypes:  repo.policyTypes,
		Policies:     repo.policies,
		Enforcements: repo.enforcements,
		Metrics:      repo.metrics,
		ExportTime:   time.Now(),
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal export data: %w", err)
	}

	return data, nil
}

func (repo *InMemoryA1Repository) Import(data []byte) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	var importData struct {
		PolicyTypes  map[string]*A1PolicyType      `json:"policy_types"`
		Policies     map[string]*A1Policy          `json:"policies"`
		Enforcements map[string]*PolicyEnforcement `json:"enforcements"`
		Metrics      map[string]*PolicyMetrics     `json:"metrics"`
		ExportTime   time.Time                     `json:"export_time"`
	}

	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	// Clear existing data
	repo.policyTypes = make(map[string]*A1PolicyType)
	repo.policies = make(map[string]*A1Policy)
	repo.enforcements = make(map[string]*PolicyEnforcement)
	repo.metrics = make(map[string]*PolicyMetrics)

	// Import data
	repo.policyTypes = importData.PolicyTypes
	repo.policies = importData.Policies
	repo.enforcements = importData.Enforcements
	repo.metrics = importData.Metrics

	repo.logger.WithFields(logrus.Fields{
		"policy_types": len(repo.policyTypes),
		"policies":     len(repo.policies),
		"export_time":  importData.ExportTime,
	}).Info("Data imported successfully")

	return nil
}

// Cleanup removes old enforcement records and metrics
func (repo *InMemoryA1Repository) Cleanup(maxAge time.Duration) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	cleaned := 0

	// Clean old enforcement records
	for policyID, enforcement := range repo.enforcements {
		if enforcement.EnforcementTime.Before(cutoff) {
			// Only clean if the policy no longer exists
			if _, exists := repo.policies[policyID]; !exists {
				delete(repo.enforcements, policyID)
				delete(repo.metrics, policyID)
				cleaned++
			}
		}
	}

	if cleaned > 0 {
		repo.logger.WithField("cleaned_records", cleaned).Info("Cleanup completed")
	}

	return nil
}

// A1PolicyValidatorImpl implements policy validation according to JSON Schema
type A1PolicyValidatorImpl struct {
	logger *logrus.Logger
}

// NewA1PolicyValidator creates a new policy validator
func NewA1PolicyValidator() A1PolicyValidator {
	return &A1PolicyValidatorImpl{
		logger: logrus.WithField("component", "a1-validator").Logger,
	}
}

// ValidatePolicyType validates a policy type definition
func (v *A1PolicyValidatorImpl) ValidatePolicyType(policyType *A1PolicyType) error {
	if policyType.PolicyTypeID == "" {
		return fmt.Errorf("policy type ID cannot be empty")
	}

	if policyType.Name == "" {
		return fmt.Errorf("policy type name cannot be empty")
	}

	if policyType.PolicySchema == nil {
		return fmt.Errorf("policy schema cannot be nil")
	}

	// Validate the schema structure
	if err := v.ValidateSchema(policyType.PolicySchema); err != nil {
		return fmt.Errorf("invalid policy schema: %w", err)
	}

	return nil
}

// ValidatePolicy validates a policy against its policy type schema
func (v *A1PolicyValidatorImpl) ValidatePolicy(policy *A1Policy, policyType *A1PolicyType) error {
	if policy.PolicyID == "" {
		return fmt.Errorf("policy ID cannot be empty")
	}

	if policy.PolicyTypeID != policyType.PolicyTypeID {
		return fmt.Errorf("policy type ID mismatch: expected %s, got %s",
			policyType.PolicyTypeID, policy.PolicyTypeID)
	}

	if policy.PolicyData == nil {
		return fmt.Errorf("policy data cannot be nil")
	}

	// Validate policy data against schema
	if err := v.validateAgainstSchema(policy.PolicyData, policyType.PolicySchema); err != nil {
		return fmt.Errorf("policy data validation failed: %w", err)
	}

	return nil
}

// ValidateSchema validates a JSON schema structure
func (v *A1PolicyValidatorImpl) ValidateSchema(schema map[string]interface{}) error {
	// Check for required JSON Schema fields
	schemaType, exists := schema["type"]
	if !exists {
		return fmt.Errorf("schema must have a 'type' field")
	}

	typeStr, ok := schemaType.(string)
	if !ok {
		return fmt.Errorf("schema 'type' must be a string")
	}

	// Validate based on type
	switch typeStr {
	case "object":
		return v.validateObjectSchema(schema)
	case "array":
		return v.validateArraySchema(schema)
	case "string":
		return v.validateStringSchema(schema)
	case "number", "integer":
		return v.validateNumberSchema(schema)
	case "boolean":
		return nil // Boolean schemas are simple
	default:
		return fmt.Errorf("unsupported schema type: %s", typeStr)
	}
}

// validateAgainstSchema validates data against a JSON schema
func (v *A1PolicyValidatorImpl) validateAgainstSchema(data interface{}, schema map[string]interface{}) error {
	schemaType, exists := schema["type"]
	if !exists {
		return fmt.Errorf("schema missing 'type' field")
	}

	typeStr, ok := schemaType.(string)
	if !ok {
		return fmt.Errorf("schema 'type' must be a string")
	}

	switch typeStr {
	case "object":
		return v.validateObjectData(data, schema)
	case "array":
		return v.validateArrayData(data, schema)
	case "string":
		return v.validateStringData(data, schema)
	case "number":
		return v.validateNumberData(data, schema, false)
	case "integer":
		return v.validateNumberData(data, schema, true)
	case "boolean":
		return v.validateBooleanData(data, schema)
	default:
		return fmt.Errorf("unsupported schema type: %s", typeStr)
	}
}

// Schema validation helpers

func (v *A1PolicyValidatorImpl) validateObjectSchema(schema map[string]interface{}) error {
	// Check properties if they exist
	if properties, exists := schema["properties"]; exists {
		propsMap, ok := properties.(map[string]interface{})
		if !ok {
			return fmt.Errorf("'properties' must be an object")
		}

		// Validate each property schema
		for propName, propSchema := range propsMap {
			propSchemaMap, ok := propSchema.(map[string]interface{})
			if !ok {
				return fmt.Errorf("property '%s' schema must be an object", propName)
			}

			if err := v.ValidateSchema(propSchemaMap); err != nil {
				return fmt.Errorf("invalid schema for property '%s': %w", propName, err)
			}
		}
	}

	// Validate required array if it exists
	if required, exists := schema["required"]; exists {
		reqArray, ok := required.([]interface{})
		if !ok {
			// Try []string
			if reqStringArray, ok := required.([]string); ok {
				// Convert to []interface{}
				reqArray = make([]interface{}, len(reqStringArray))
				for i, s := range reqStringArray {
					reqArray[i] = s
				}
			} else {
				return fmt.Errorf("'required' must be an array")
			}
		}

		if reqArray != nil {
			// Validate that all required fields are strings
			for i, req := range reqArray {
				if _, ok := req.(string); !ok {
					return fmt.Errorf("required field %d must be a string", i)
				}
			}
		}
	}

	return nil
}

func (v *A1PolicyValidatorImpl) validateArraySchema(schema map[string]interface{}) error {
	// Check items if they exist
	if items, exists := schema["items"]; exists {
		itemsMap, ok := items.(map[string]interface{})
		if !ok {
			return fmt.Errorf("'items' must be an object")
		}

		if err := v.ValidateSchema(itemsMap); err != nil {
			return fmt.Errorf("invalid items schema: %w", err)
		}
	}

	return nil
}

func (v *A1PolicyValidatorImpl) validateStringSchema(schema map[string]interface{}) error {
	// Validate enum if it exists
	if enum, exists := schema["enum"]; exists {
		enumArray, ok := enum.([]interface{})
		if !ok {
			return fmt.Errorf("enum must be an array")
		}

		// All enum values should be strings for string type
		for i, val := range enumArray {
			if _, ok := val.(string); !ok {
				return fmt.Errorf("enum value %d must be a string", i)
			}
		}
	}

	return nil
}

func (v *A1PolicyValidatorImpl) validateNumberSchema(schema map[string]interface{}) error {
	// Validate minimum if it exists
	if minimum, exists := schema["minimum"]; exists {
		if _, ok := minimum.(float64); !ok {
			if _, ok := minimum.(int); !ok {
				return fmt.Errorf("'minimum' must be a number")
			}
		}
	}

	// Validate maximum if it exists
	if maximum, exists := schema["maximum"]; exists {
		if _, ok := maximum.(float64); !ok {
			if _, ok := maximum.(int); !ok {
				return fmt.Errorf("'maximum' must be a number")
			}
		}
	}

	return nil
}

// Data validation helpers

func (v *A1PolicyValidatorImpl) validateObjectData(data interface{}, schema map[string]interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object, got %T", data)
	}

	// Check required fields
	if required, exists := schema["required"]; exists {
		reqArray, ok := required.([]interface{})
		if !ok {
			// Try []string
			if reqStringArray, ok := required.([]string); ok {
				reqArray = make([]interface{}, len(reqStringArray))
				for i, s := range reqStringArray {
					reqArray[i] = s
				}
			}
		}

		if reqArray != nil {
			for _, req := range reqArray {
				reqStr, ok := req.(string)
				if !ok {
					continue
				}

				if _, exists := dataMap[reqStr]; !exists {
					return fmt.Errorf("missing required field: %s", reqStr)
				}
			}
		}
	}

	// Validate properties
	if properties, exists := schema["properties"]; exists {
		propsMap, ok := properties.(map[string]interface{})
		if !ok {
			return fmt.Errorf("properties schema must be an object")
		}

		for fieldName, fieldValue := range dataMap {
			if propSchema, exists := propsMap[fieldName]; exists {
				propSchemaMap, ok := propSchema.(map[string]interface{})
				if !ok {
					return fmt.Errorf("property schema for '%s' must be an object", fieldName)
				}

				if err := v.validateAgainstSchema(fieldValue, propSchemaMap); err != nil {
					return fmt.Errorf("validation failed for field '%s': %w", fieldName, err)
				}
			} else {
				// Check if additional properties are allowed
				if additionalProps, exists := schema["additionalProperties"]; exists {
					if allowed, ok := additionalProps.(bool); ok && !allowed {
						return fmt.Errorf("additional property '%s' not allowed", fieldName)
					}
					// If additionalProperties is a schema, validate against it
					if addPropsSchema, ok := additionalProps.(map[string]interface{}); ok {
						if err := v.validateAgainstSchema(fieldValue, addPropsSchema); err != nil {
							return fmt.Errorf("validation failed for additional property '%s': %w", fieldName, err)
						}
					}
				}
			}
		}
	}

	return nil
}

func (v *A1PolicyValidatorImpl) validateArrayData(data interface{}, schema map[string]interface{}) error {
	dataArray, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("expected array, got %T", data)
	}

	// Validate items if schema specifies
	if items, exists := schema["items"]; exists {
		itemsMap, ok := items.(map[string]interface{})
		if !ok {
			return fmt.Errorf("items schema must be an object")
		}

		for i, item := range dataArray {
			if err := v.validateAgainstSchema(item, itemsMap); err != nil {
				return fmt.Errorf("validation failed for item %d: %w", i, err)
			}
		}
	}

	return nil
}

func (v *A1PolicyValidatorImpl) validateStringData(data interface{}, schema map[string]interface{}) error {
	dataStr, ok := data.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", data)
	}

	// Check enum constraint
	if enum, exists := schema["enum"]; exists {
		enumArray, ok := enum.([]interface{})
		if !ok {
			return fmt.Errorf("enum must be an array")
		}

		found := false
		for _, val := range enumArray {
			if valStr, ok := val.(string); ok && valStr == dataStr {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("value '%s' not allowed by enum constraint", dataStr)
		}
	}

	// Check minLength
	if minLength, exists := schema["minLength"]; exists {
		if minLen, ok := minLength.(float64); ok {
			if len(dataStr) < int(minLen) {
				return fmt.Errorf("string length %d is less than minimum %d", len(dataStr), int(minLen))
			}
		}
	}

	// Check maxLength
	if maxLength, exists := schema["maxLength"]; exists {
		if maxLen, ok := maxLength.(float64); ok {
			if len(dataStr) > int(maxLen) {
				return fmt.Errorf("string length %d exceeds maximum %d", len(dataStr), int(maxLen))
			}
		}
	}

	// Check pattern (simplified - would use regex in production)
	if pattern, exists := schema["pattern"]; exists {
		if patternStr, ok := pattern.(string); ok {
			// Simplified pattern matching
			if strings.Contains(patternStr, "^") && !strings.HasPrefix(dataStr, strings.TrimPrefix(patternStr, "^")) {
				return fmt.Errorf("string does not match pattern: %s", patternStr)
			}
		}
	}

	return nil
}

func (v *A1PolicyValidatorImpl) validateNumberData(data interface{}, schema map[string]interface{}, mustBeInteger bool) error {
	var numVal float64
	var ok bool

	if mustBeInteger {
		// For integer type, accept both int and float64 but ensure it's a whole number
		switch val := data.(type) {
		case int:
			numVal = float64(val)
			ok = true
		case float64:
			numVal = val
			ok = true
			// Check if it's actually a whole number
			if numVal != float64(int(numVal)) {
				return fmt.Errorf("expected integer, got float with decimal part")
			}
		case json.Number:
			if f, err := val.Float64(); err == nil {
				numVal = f
				ok = true
				if mustBeInteger && numVal != float64(int(numVal)) {
					return fmt.Errorf("expected integer, got float with decimal part")
				}
			}
		}
	} else {
		// For number type, accept int, float64, and json.Number
		switch val := data.(type) {
		case int:
			numVal = float64(val)
			ok = true
		case float64:
			numVal = val
			ok = true
		case json.Number:
			if f, err := val.Float64(); err == nil {
				numVal = f
				ok = true
			}
		}
	}

	if !ok {
		expectedType := "number"
		if mustBeInteger {
			expectedType = "integer"
		}
		return fmt.Errorf("expected %s, got %T", expectedType, data)
	}

	// Check minimum constraint
	if minimum, exists := schema["minimum"]; exists {
		var minVal float64
		switch min := minimum.(type) {
		case int:
			minVal = float64(min)
		case float64:
			minVal = min
		case json.Number:
			if f, err := min.Float64(); err == nil {
				minVal = f
			}
		}

		if numVal < minVal {
			return fmt.Errorf("value %v is less than minimum %v", numVal, minVal)
		}
	}

	// Check maximum constraint
	if maximum, exists := schema["maximum"]; exists {
		var maxVal float64
		switch max := maximum.(type) {
		case int:
			maxVal = float64(max)
		case float64:
			maxVal = max
		case json.Number:
			if f, err := max.Float64(); err == nil {
				maxVal = f
			}
		}

		if numVal > maxVal {
			return fmt.Errorf("value %v exceeds maximum %v", numVal, maxVal)
		}
	}

	return nil
}

func (v *A1PolicyValidatorImpl) validateBooleanData(data interface{}, schema map[string]interface{}) error {
	if _, ok := data.(bool); !ok {
		return fmt.Errorf("expected boolean, got %T", data)
	}

	return nil
}

// ValidateSpecificPolicyTypes provides validation for O-RAN specific policy types

// ValidateQoSPolicy validates a QoS policy according to O-RAN specifications
func (v *A1PolicyValidatorImpl) ValidateQoSPolicy(policyData map[string]interface{}) error {
	// Validate QCI
	qci, exists := policyData["qci"]
	if !exists {
		return fmt.Errorf("QCI is required")
	}

	qciNum, ok := qci.(float64)
	if !ok {
		if qciInt, ok := qci.(int); ok {
			qciNum = float64(qciInt)
		} else {
			return fmt.Errorf("QCI must be a number")
		}
	}

	if qciNum < 1 || qciNum > 9 {
		return fmt.Errorf("QCI must be between 1 and 9")
	}

	// Validate priority level
	if priority, exists := policyData["priority_level"]; exists {
		priorityNum, ok := priority.(float64)
		if !ok {
			if priorityInt, ok := priority.(int); ok {
				priorityNum = float64(priorityInt)
			} else {
				return fmt.Errorf("priority_level must be a number")
			}
		}

		if priorityNum < 1 || priorityNum > 15 {
			return fmt.Errorf("priority_level must be between 1 and 15")
		}
	}

	// Validate packet delay budget
	if pdb, exists := policyData["packet_delay_budget"]; exists {
		pdbNum, ok := pdb.(float64)
		if !ok {
			if pdbInt, ok := pdb.(int); ok {
				pdbNum = float64(pdbInt)
			} else {
				return fmt.Errorf("packet_delay_budget must be a number")
			}
		}

		if pdbNum < 50 || pdbNum > 300 {
			return fmt.Errorf("packet_delay_budget must be between 50 and 300 ms")
		}
	}

	return nil
}

// ValidateRRMPolicy validates a Radio Resource Management policy
func (v *A1PolicyValidatorImpl) ValidateRRMPolicy(policyData map[string]interface{}) error {
	// Validate resource type
	resourceType, exists := policyData["resource_type"]
	if !exists {
		return fmt.Errorf("resource_type is required")
	}

	resourceTypeStr, ok := resourceType.(string)
	if !ok {
		return fmt.Errorf("resource_type must be a string")
	}

	validResourceTypes := map[string]bool{
		"PRB":      true,
		"SPECTRUM": true,
		"POWER":    true,
	}

	if !validResourceTypes[resourceTypeStr] {
		return fmt.Errorf("invalid resource_type: %s", resourceTypeStr)
	}

	// Validate allocation strategy
	strategy, exists := policyData["allocation_strategy"]
	if !exists {
		return fmt.Errorf("allocation_strategy is required")
	}

	strategyStr, ok := strategy.(string)
	if !ok {
		return fmt.Errorf("allocation_strategy must be a string")
	}

	validStrategies := map[string]bool{
		"ROUND_ROBIN":       true,
		"PROPORTIONAL_FAIR": true,
		"MAX_THROUGHPUT":    true,
	}

	if !validStrategies[strategyStr] {
		return fmt.Errorf("invalid allocation_strategy: %s", strategyStr)
	}

	return nil
}

// ValidateSONPolicy validates a Self-Organizing Networks policy
func (v *A1PolicyValidatorImpl) ValidateSONPolicy(policyData map[string]interface{}) error {
	// Validate function type
	functionType, exists := policyData["function_type"]
	if !exists {
		return fmt.Errorf("function_type is required")
	}

	functionTypeStr, ok := functionType.(string)
	if !ok {
		return fmt.Errorf("function_type must be a string")
	}

	validFunctionTypes := map[string]bool{
		"MLB": true,
		"MRO": true,
		"CCO": true,
		"ANR": true,
	}

	if !validFunctionTypes[functionTypeStr] {
		return fmt.Errorf("invalid function_type: %s", functionTypeStr)
	}

	// Validate optimization target
	target, exists := policyData["optimization_target"]
	if !exists {
		return fmt.Errorf("optimization_target is required")
	}

	targetStr, ok := target.(string)
	if !ok {
		return fmt.Errorf("optimization_target must be a string")
	}

	validTargets := map[string]bool{
		"THROUGHPUT": true,
		"LATENCY":    true,
		"COVERAGE":   true,
		"CAPACITY":   true,
	}

	if !validTargets[targetStr] {
		return fmt.Errorf("invalid optimization_target: %s", targetStr)
	}

	return nil
}
