package a1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/hctsai1006/near-rt-ric/pkg/common/monitoring"
	"github.com/sirupsen/logrus"
)

// APIHandlers contains all A1 REST API handlers
type APIHandlers struct {
	policyManager     *PolicyManager
	mlModelManager    *MLModelManager
	enrichmentManager *EnrichmentManager
	authService       *AuthService
	logger            *logrus.Logger
	metrics           *monitoring.MetricsCollector
	startTime         time.Time
}

// NewAPIHandlers creates new API handlers
func NewAPIHandlers(policyManager *PolicyManager, mlModelManager *MLModelManager, enrichmentManager *EnrichmentManager, authService *AuthService, logger *logrus.Logger, metrics *monitoring.MetricsCollector) *APIHandlers {
	return &APIHandlers{
		policyManager:     policyManager,
		mlModelManager:    mlModelManager,
		enrichmentManager: enrichmentManager,
		authService:       authService,
		logger:            logger,
		metrics:           metrics,
		startTime:         time.Now(),
	}
}

// SetupRoutes sets up all A1 REST API routes
func (h *APIHandlers) SetupRoutes(r *mux.Router) {
	// Health and status endpoints (no auth required)
	r.HandleFunc("/a1-p/healthcheck", h.HealthCheck).Methods("GET")
	r.HandleFunc("/a1-p/status", h.GetStatus).Methods("GET")

	// Authentication endpoints
	auth := r.PathPrefix("/a1-p/auth").Subrouter()
	auth.HandleFunc("/token", h.GenerateToken).Methods("POST")
	auth.HandleFunc("/refresh", h.RefreshToken).Methods("POST")
	auth.HandleFunc("/revoke", h.RevokeToken).Methods("POST")

	// Policy Type endpoints (require authentication)
	policyTypes := r.PathPrefix("/a1-p/policytypes").Subrouter()
	policyTypes.Use(h.authService.AuthMiddleware())
	
	policyTypes.HandleFunc("", h.GetPolicyTypes).Methods("GET")
	policyTypes.HandleFunc("/{policy_type_id}", h.GetPolicyType).Methods("GET")
	policyTypes.Handle("/{policy_type_id}", h.authService.RequirePermission("policytype:write")(http.HandlerFunc(h.CreatePolicyType))).Methods("PUT")
	policyTypes.Handle("/{policy_type_id}", h.authService.RequirePermission("policytype:delete")(http.HandlerFunc(h.DeletePolicyType))).Methods("DELETE")

	// Policy Instance endpoints (require authentication)
	policies := r.PathPrefix("/a1-p/policytypes/{policy_type_id}/policies").Subrouter()
	policies.Use(h.authService.AuthMiddleware())
	
	policies.HandleFunc("", h.GetPolicyInstances).Methods("GET")
	policies.HandleFunc("/{policy_id}", h.GetPolicyInstance).Methods("GET")
	policies.Handle("/{policy_id}", h.authService.RequirePermission("policy:write")(http.HandlerFunc(h.CreatePolicyInstance))).Methods("PUT")
	policies.Handle("/{policy_id}", h.authService.RequirePermission("policy:delete")(http.HandlerFunc(h.DeletePolicyInstance))).Methods("DELETE")
	policies.HandleFunc("/{policy_id}/status", h.GetPolicyStatus).Methods("GET")

	// ML Model Management endpoints
	mlModels := r.PathPrefix("/a1-ml/v1/models").Subrouter()
	mlModels.Use(h.authService.AuthMiddleware())

	mlModels.HandleFunc("", h.GetMLModels).Methods("GET")
	mlModels.HandleFunc("/{model_id}", h.GetMLModel).Methods("GET")
	mlModels.Handle("/{model_id}", h.authService.RequirePermission("model:write")(http.HandlerFunc(h.DeployMLModel))).Methods("PUT")
	mlModels.Handle("/{model_id}", h.authService.RequirePermission("model:delete")(http.HandlerFunc(h.DeleteMLModel))).Methods("DELETE")

	// Enrichment Information endpoints
	enrichment := r.PathPrefix("/a1-ei/v1/info").Subrouter()
	enrichment.Use(h.authService.AuthMiddleware())

	enrichment.HandleFunc("", h.GetEnrichmentJobs).Methods("GET")
	enrichment.HandleFunc("/{job_id}", h.GetEnrichmentJob).Methods("GET")
	enrichment.Handle("/{job_id}", h.authService.RequirePermission("enrichment:write")(http.HandlerFunc(h.CreateEnrichmentJob))).Methods("PUT")
	enrichment.Handle("/{job_id}", h.authService.RequirePermission("enrichment:delete")(http.HandlerFunc(h.DeleteEnrichmentJob))).Methods("DELETE")
}

// HealthCheck returns the health status of the A1 interface
func (h *APIHandlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer h.recordMetrics(r, start, http.StatusOK)

	health := &A1HealthCheck{
		Status:    "UP",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Components: map[string]string{
			"policy_manager": "UP",
			"auth_service":   "UP",
		},
		Uptime: time.Since(h.startTime),
	}

	h.writeJSONResponse(w, http.StatusOK, health)
}

// GetStatus returns detailed status information
func (h *APIHandlers) GetStatus(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer h.recordMetrics(r, start, http.StatusOK)

	stats, err := h.policyManager.GetStatistics()
	if err != nil {
		h.recordMetrics(r, start, http.StatusInternalServerError)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve statistics", err.Error())
		return
	}
	stats.Uptime = time.Since(h.startTime)

	h.writeJSONResponse(w, http.StatusOK, stats)
}

// GenerateToken generates a new JWT token
func (h *APIHandlers) GenerateToken(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	var req TokenRequest
	if err := h.decodeJSONRequest(r, &req); err != nil {
		h.recordMetrics(r, start, http.StatusBadRequest)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Authenticate user
	user, err := h.authService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		h.recordMetrics(r, start, http.StatusUnauthorized)
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials", "")
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(user.UserID, user.Username, user.Email, user.Roles)
	if err != nil {
		h.recordMetrics(r, start, http.StatusInternalServerError)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusOK)
	h.writeJSONResponse(w, http.StatusOK, token)
}

// RefreshToken refreshes an existing token
func (h *APIHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	var req RefreshTokenRequest
	if err := h.decodeJSONRequest(r, &req); err != nil {
		h.recordMetrics(r, start, http.StatusBadRequest)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// For demo purposes, return error - refresh tokens not implemented
	h.recordMetrics(r, start, http.StatusNotImplemented)
	h.writeErrorResponse(w, http.StatusNotImplemented, "Refresh tokens not implemented", "")
}

// RevokeToken revokes a token
func (h *APIHandlers) RevokeToken(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Get token from Authorization header
	user, ok := GetUserFromContext(r.Context())
	if !ok {
		h.recordMetrics(r, start, http.StatusUnauthorized)
		h.writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated", "")
		return
	}

	// Revoke token
	if err := h.authService.RevokeToken(user.Token); err != nil {
		h.recordMetrics(r, start, http.StatusInternalServerError)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to revoke token", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusOK)
	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Token revoked successfully"})
}

// GetPolicyTypes returns all policy types
func (h *APIHandlers) GetPolicyTypes(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer h.recordMetrics(r, start, http.StatusOK)

	policyTypes, err := h.policyManager.GetAllPolicyTypes()
	if err != nil {
		h.recordMetrics(r, start, http.StatusInternalServerError)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve policy types", err.Error())
		return
	}
	
	// Extract just the IDs for the list response
	typeIDs := make([]PolicyTypeID, len(policyTypes))
	for i, pt := range policyTypes {
		typeIDs[i] = pt.PolicyTypeID
	}

	h.writeJSONResponse(w, http.StatusOK, typeIDs)
}

// GetPolicyType returns a specific policy type
func (h *APIHandlers) GetPolicyType(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	vars := mux.Vars(r)
	policyTypeID := PolicyTypeID(vars["policy_type_id"])

	policyType, err := h.policyManager.GetPolicyType(policyTypeID)
	if err != nil {
		h.recordMetrics(r, start, http.StatusNotFound)
		h.writeErrorResponse(w, http.StatusNotFound, "Policy type not found", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusOK)
	h.writeJSONResponse(w, http.StatusOK, policyType)
}

// CreatePolicyType creates a new policy type
func (h *APIHandlers) CreatePolicyType(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	vars := mux.Vars(r)
	policyTypeID := PolicyTypeID(vars["policy_type_id"])

	var req PolicyTypeCreateRequest
	if err := h.decodeJSONRequest(r, &req); err != nil {
		h.recordMetrics(r, start, http.StatusBadRequest)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Set the policy type ID from URL
	req.PolicyTypeID = policyTypeID

	policyType, err := h.policyManager.CreatePolicyType(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "policy type already exists" || err.Error() == "maximum number of policy types reached" {
			statusCode = http.StatusConflict
		}
		h.recordMetrics(r, start, statusCode)
		h.writeErrorResponse(w, statusCode, "Failed to create policy type", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusCreated)
	h.writeJSONResponse(w, http.StatusCreated, policyType)
}

// DeletePolicyType deletes a policy type
func (h *APIHandlers) DeletePolicyType(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	vars := mux.Vars(r)
	policyTypeID := PolicyTypeID(vars["policy_type_id"])

	if err := h.policyManager.DeletePolicyType(policyTypeID); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "policy type not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "active policy instances exist" {
			statusCode = http.StatusConflict
		}
		h.recordMetrics(r, start, statusCode)
		h.writeErrorResponse(w, statusCode, "Failed to delete policy type", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusNoContent)
	w.WriteHeader(http.StatusNoContent)
}

// GetPolicyInstances returns all policy instances for a policy type
func (h *APIHandlers) GetPolicyInstances(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	vars := mux.Vars(r)
	policyTypeID := PolicyTypeID(vars["policy_type_id"])

	policies, err := h.policyManager.GetPolicyInstancesByType(policyTypeID)
	if err != nil {
		h.recordMetrics(r, start, http.StatusInternalServerError)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve policy instances", err.Error())
		return
	}
	
	// Extract just the IDs for the list response
	policyIDs := make([]PolicyID, len(policies))
	for i, policy := range policies {
		policyIDs[i] = policy.PolicyID
	}

	h.recordMetrics(r, start, http.StatusOK)
	h.writeJSONResponse(w, http.StatusOK, policyIDs)
}

// GetPolicyInstance returns a specific policy instance
func (h *APIHandlers) GetPolicyInstance(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	vars := mux.Vars(r)
	policyID := PolicyID(vars["policy_id"])

	policy, err := h.policyManager.GetPolicyInstance(policyID)
	if err != nil {
		h.recordMetrics(r, start, http.StatusNotFound)
		h.writeErrorResponse(w, http.StatusNotFound, "Policy instance not found", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusOK)
	h.writeJSONResponse(w, http.StatusOK, policy)
}

// CreatePolicyInstance creates a new policy instance
func (h *APIHandlers) CreatePolicyInstance(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	vars := mux.Vars(r)
	policyTypeID := PolicyTypeID(vars["policy_type_id"])
	policyID := PolicyID(vars["policy_id"])

	var req PolicyInstanceCreateRequest
	if err := h.decodeJSONRequest(r, &req); err != nil {
		h.recordMetrics(r, start, http.StatusBadRequest)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Get user from context
	user, _ := GetUserFromContext(r.Context())
	userID := ""
	if user != nil {
		userID = user.UserID
	}

	policy, err := h.policyManager.CreatePolicyInstance(policyTypeID, policyID, &req, userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "policy type not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "policy instance already exists" || err.Error() == "maximum number of policy instances reached" {
			statusCode = http.StatusConflict
		} else if err.Error() == "policy data validation failed" {
			statusCode = http.StatusBadRequest
		}
		h.recordMetrics(r, start, statusCode)
		h.writeErrorResponse(w, statusCode, "Failed to create policy instance", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusCreated)
	h.writeJSONResponse(w, http.StatusCreated, policy)
}

// DeletePolicyInstance deletes a policy instance
func (h *APIHandlers) DeletePolicyInstance(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	vars := mux.Vars(r)
	policyID := PolicyID(vars["policy_id"])

	// Get user from context
	user, _ := GetUserFromContext(r.Context())
	userID := ""
	if user != nil {
		userID = user.UserID
	}

	if err := h.policyManager.DeletePolicyInstance(policyID, userID); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "policy instance not found" || err.Error() == "policy instance is already deleted" {
			statusCode = http.StatusNotFound
		}
		h.recordMetrics(r, start, statusCode)
		h.writeErrorResponse(w, statusCode, "Failed to delete policy instance", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusNoContent)
	w.WriteHeader(http.StatusNoContent)
}

// GetPolicyStatus returns the status of a policy instance
func (h *APIHandlers) GetPolicyStatus(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	vars := mux.Vars(r)
	policyID := PolicyID(vars["policy_id"])

	status, err := h.policyManager.GetPolicyStatus(policyID)
	if err != nil {
		h.recordMetrics(r, start, http.StatusNotFound)
		h.writeErrorResponse(w, http.StatusNotFound, "Policy instance not found", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusOK)
	h.writeJSONResponse(w, http.StatusOK, status)
}

// ML Model Management Handlers

func (h *APIHandlers) GetMLModels(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer h.recordMetrics(r, start, http.StatusOK)

	models, err := h.mlModelManager.GetMLModels()
	if err != nil {
		h.recordMetrics(r, start, http.StatusInternalServerError)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve ML models", err.Error())
		return
	}
	h.writeJSONResponse(w, http.StatusOK, models)
}

func (h *APIHandlers) GetMLModel(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	vars := mux.Vars(r)
	modelID := vars["model_id"]

	model, err := h.mlModelManager.GetMLModel(modelID)
	if err != nil {
		h.recordMetrics(r, start, http.StatusNotFound)
		h.writeErrorResponse(w, http.StatusNotFound, "ML model not found", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusOK)
	h.writeJSONResponse(w, http.StatusOK, model)
}

func (h *APIHandlers) DeployMLModel(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req MLModelDeployRequest
	if err := h.decodeJSONRequest(r, &req); err != nil {
		h.recordMetrics(r, start, http.StatusBadRequest)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	model, err := h.mlModelManager.DeployMLModel(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "ML model already exists" || err.Error() == "maximum number of ML models reached" {
			statusCode = http.StatusConflict
		}
		h.recordMetrics(r, start, statusCode)
		h.writeErrorResponse(w, statusCode, "Failed to deploy ML model", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusCreated)
	h.writeJSONResponse(w, http.StatusCreated, model)
}

func (h *APIHandlers) DeleteMLModel(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	vars := mux.Vars(r)
	modelID := vars["model_id"]

	if err := h.mlModelManager.DeleteMLModel(modelID); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "ML model not found" {
			statusCode = http.StatusNotFound
		}
		h.recordMetrics(r, start, statusCode)
		h.writeErrorResponse(w, statusCode, "Failed to delete ML model", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusNoContent)
	w.WriteHeader(http.StatusNoContent)
}

// Enrichment Information Handlers

func (h *APIHandlers) GetEnrichmentJobs(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer h.recordMetrics(r, start, http.StatusOK)

	jobs, err := h.enrichmentManager.GetEnrichmentJobs()
	if err != nil {
		h.recordMetrics(r, start, http.StatusInternalServerError)
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve enrichment jobs", err.Error())
		return
	}
	h.writeJSONResponse(w, http.StatusOK, jobs)
}

func (h *APIHandlers) GetEnrichmentJob(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	vars := mux.Vars(r)
	jobID := vars["job_id"]

	job, err := h.enrichmentManager.GetEnrichmentJob(jobID)
	if err != nil {
		h.recordMetrics(r, start, http.StatusNotFound)
		h.writeErrorResponse(w, http.StatusNotFound, "Enrichment job not found", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusOK)
	h.writeJSONResponse(w, http.StatusOK, job)
}

func (h *APIHandlers) CreateEnrichmentJob(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req EIJobCreateRequest
	if err := h.decodeJSONRequest(r, &req); err != nil {
		h.recordMetrics(r, start, http.StatusBadRequest)
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	job, err := h.enrichmentManager.CreateEnrichmentJob(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "enrichment job already exists" {
			statusCode = http.StatusConflict
		}
		h.recordMetrics(r, start, statusCode)
		h.writeErrorResponse(w, statusCode, "Failed to create enrichment job", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusCreated)
	h.writeJSONResponse(w, http.StatusCreated, job)
}

func (h *APIHandlers) DeleteEnrichmentJob(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	vars := mux.Vars(r)
	jobID := vars["job_id"]

	if err := h.enrichmentManager.DeleteEnrichmentJob(jobID); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "enrichment job not found" {
			statusCode = http.StatusNotFound
		}
		h.recordMetrics(r, start, statusCode)
		h.writeErrorResponse(w, statusCode, "Failed to delete enrichment job", err.Error())
		return
	}

	h.recordMetrics(r, start, http.StatusNoContent)
	w.WriteHeader(http.StatusNoContent)
}



// Helper methods

func (h *APIHandlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (h *APIHandlers) writeErrorResponse(w http.ResponseWriter, statusCode int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(statusCode)
	
	errorResp := A1ErrorResponse{
		Type:   "https://tools.ietf.org/html/rfc7807",
		Title:  title,
		Status: statusCode,
		Detail: detail,
	}
	
	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		h.logger.WithError(err).Error("Failed to encode error response")
	}
}

func (h *APIHandlers) decodeJSONRequest(r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("content-type must be application/json")
	}
	
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	
	return decoder.Decode(dst)
}

func (h *APIHandlers) recordMetrics(r *http.Request, start time.Time, statusCode int) {
	duration := time.Since(start)
	method := r.Method
	path := r.URL.Path
	
	// Record metrics
	h.metrics.RecordA1Request(method, path, statusCode, duration, 0, 0)
	
	// Log request
	h.logger.WithFields(logrus.Fields{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration":    duration,
		"remote_addr": r.RemoteAddr,
	}).Info("A1 API request processed")
}