package a1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestA1Interface_HealthCheck(t *testing.T) {
	// Create test configuration
	config := &A1InterfaceConfig{
		ListenAddress:         "0.0.0.0",
		ListenPort:            10020,
		TLSEnabled:            false,
		AuthenticationEnabled: false,
		RequestTimeout:        30 * time.Second,
		MaxRequestSize:        1024 * 1024,
		RateLimitEnabled:      false,
		NotificationEnabled:   false,
		LogLevel:              "info",
	}

	// Create repository
	repo := NewInMemoryA1Repository()

	// Create A1 interface
	a1Interface := NewA1Interface(config, repo)

	// Create test request
	req := httptest.NewRequest("GET", "/a1-p/healthcheck", nil)
	w := httptest.NewRecorder()

	// Handle request
	a1Interface.handleHealthCheck(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response A1HealthStatus
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "OK", response.Status)
}

func TestA1Interface_PolicyTypeManagement(t *testing.T) {
	// Setup
	config := &A1InterfaceConfig{
		ListenAddress:         "0.0.0.0",
		ListenPort:            10020,
		TLSEnabled:            false,
		AuthenticationEnabled: false,
		RequestTimeout:        30 * time.Second,
		MaxRequestSize:        1024 * 1024,
		RateLimitEnabled:      false,
		NotificationEnabled:   false,
		LogLevel:              "info",
	}

	repo := NewInMemoryA1Repository()
	a1Interface := NewA1Interface(config, repo)

	// Test creating policy type
	policyTypeSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"priority": map[string]interface{}{
				"type":    "integer",
				"minimum": 1,
				"maximum": 10,
			},
		},
		"required": []string{"priority"},
	}

	body, _ := json.Marshal(policyTypeSchema)
	req := httptest.NewRequest("PUT", "/a1-p/policytypes/test-policy-type", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	a1Interface.handleCreatePolicyType(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Test getting policy type
	req = httptest.NewRequest("GET", "/a1-p/policytypes/test-policy-type", nil)
	w = httptest.NewRecorder()

	// Mock mux vars
	req = req.WithContext(
		contextWithVars(req.Context(), map[string]string{
			"policy_type_id": "test-policy-type",
		}),
	)

	a1Interface.handleGetPolicyType(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var retrievedSchema map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&retrievedSchema)
	require.NoError(t, err)
	assert.Equal(t, "object", retrievedSchema["type"])
}

func TestA1Interface_PolicyManagement(t *testing.T) {
	// Setup
	config := &A1InterfaceConfig{
		ListenAddress:         "0.0.0.0",
		ListenPort:            10020,
		TLSEnabled:            false,
		AuthenticationEnabled: false,
		RequestTimeout:        30 * time.Second,
		MaxRequestSize:        1024 * 1024,
		RateLimitEnabled:      false,
		NotificationEnabled:   false,
		LogLevel:              "info",
	}

	repo := NewInMemoryA1Repository()
	a1Interface := NewA1Interface(config, repo)

	// First create a policy type
	policyType := &A1PolicyType{
		PolicyTypeID: "qos-policy-v1",
		Name:         "QoS Policy",
		Description:  "Test QoS policy type",
		PolicySchema: QoSPolicyTypeSchema,
		CreatedAt:    time.Now(),
		LastModified: time.Now(),
	}
	err := repo.CreatePolicyType(policyType)
	require.NoError(t, err)

	// Test creating policy
	policyData := map[string]interface{}{
		"qci":                    5,
		"priority_level":         3,
		"packet_delay_budget":    100,
		"packet_error_loss_rate": 0.001,
	}

	body, _ := json.Marshal(policyData)
	req := httptest.NewRequest("PUT", "/a1-p/policytypes/qos-policy-v1/policies/test-policy", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-RIC-Requestor-ID", "test-requestor")
	w := httptest.NewRecorder()

	// Mock mux vars
	req = req.WithContext(
		contextWithVars(req.Context(), map[string]string{
			"policy_type_id": "qos-policy-v1",
			"policy_id":      "test-policy",
		}),
	)

	a1Interface.handleCreateOrUpdatePolicy(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Test getting policy
	req = httptest.NewRequest("GET", "/a1-p/policytypes/qos-policy-v1/policies/test-policy", nil)
	w = httptest.NewRecorder()

	req = req.WithContext(
		contextWithVars(req.Context(), map[string]string{
			"policy_type_id": "qos-policy-v1",
			"policy_id":      "test-policy",
		}),
	)

	a1Interface.handleGetPolicy(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var retrievedPolicy map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&retrievedPolicy)
	require.NoError(t, err)
	assert.Equal(t, float64(5), retrievedPolicy["qci"])
}

func TestA1Interface_PolicyStatus(t *testing.T) {
	// Setup
	config := &A1InterfaceConfig{
		ListenAddress:         "0.0.0.0",
		ListenPort:            10020,
		TLSEnabled:            false,
		AuthenticationEnabled: false,
		RequestTimeout:        30 * time.Second,
		MaxRequestSize:        1024 * 1024,
		RateLimitEnabled:      false,
		NotificationEnabled:   false,
		LogLevel:              "info",
	}

	repo := NewInMemoryA1Repository()
	a1Interface := NewA1Interface(config, repo)

	// Create policy type and policy
	policyType := &A1PolicyType{
		PolicyTypeID: "test-policy-type",
		Name:         "Test Policy",
		PolicySchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"value": map[string]interface{}{"type": "string"},
			},
		},
		CreatedAt:    time.Now(),
		LastModified: time.Now(),
	}
	repo.CreatePolicyType(policyType)

	policy := &A1Policy{
		PolicyID:     "test-policy",
		PolicyTypeID: "test-policy-type",
		PolicyData:   map[string]interface{}{"value": "test"},
		Status:       A1PolicyStatusEnforced,
		CreatedAt:    time.Now(),
		LastModified: time.Now(),
	}
	repo.CreatePolicy(policy)

	// Test getting policy status
	req := httptest.NewRequest("GET", "/a1-p/policytypes/test-policy-type/policies/test-policy/status", nil)
	w := httptest.NewRecorder()

	req = req.WithContext(
		contextWithVars(req.Context(), map[string]string{
			"policy_type_id": "test-policy-type",
			"policy_id":      "test-policy",
		}),
	)

	a1Interface.handleGetPolicyStatus(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var status map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&status)
	require.NoError(t, err)
	assert.Equal(t, "test-policy", status["policy_id"])
	assert.Equal(t, "test-policy-type", status["policy_type_id"])
	assert.Equal(t, string(A1PolicyStatusEnforced), status["status"])
}

func TestA1Interface_ErrorHandling(t *testing.T) {
	config := &A1InterfaceConfig{
		ListenAddress:         "0.0.0.0",
		ListenPort:            10020,
		TLSEnabled:            false,
		AuthenticationEnabled: false,
		RequestTimeout:        30 * time.Second,
		MaxRequestSize:        1024 * 1024,
		RateLimitEnabled:      false,
		NotificationEnabled:   false,
		LogLevel:              "info",
	}

	repo := NewInMemoryA1Repository()
	a1Interface := NewA1Interface(config, repo)

	// Test getting non-existent policy type
	req := httptest.NewRequest("GET", "/a1-p/policytypes/non-existent", nil)
	w := httptest.NewRecorder()

	req = req.WithContext(
		contextWithVars(req.Context(), map[string]string{
			"policy_type_id": "non-existent",
		}),
	)

	a1Interface.handleGetPolicyType(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResp A1ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Equal(t, A1ErrorNotFound, errorResp.ErrorCode)
}

func TestA1Interface_ValidationErrors(t *testing.T) {
	config := &A1InterfaceConfig{
		ListenAddress:         "0.0.0.0",
		ListenPort:            10020,
		TLSEnabled:            false,
		AuthenticationEnabled: false,
		RequestTimeout:        30 * time.Second,
		MaxRequestSize:        1024 * 1024,
		RateLimitEnabled:      false,
		NotificationEnabled:   false,
		LogLevel:              "info",
	}

	repo := NewInMemoryA1Repository()
	a1Interface := NewA1Interface(config, repo)

	// Test creating policy with invalid JSON
	req := httptest.NewRequest("PUT", "/a1-p/policytypes/test-type", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	req = req.WithContext(
		contextWithVars(req.Context(), map[string]string{
			"policy_type_id": "test-type",
		}),
	)

	a1Interface.handleCreatePolicyType(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp A1ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&errorResp)
	require.NoError(t, err)
	assert.Equal(t, A1ErrorBadRequest, errorResp.ErrorCode)
}

// Helper function to simulate mux variables in context
func contextWithVars(ctx context.Context, vars map[string]string) context.Context {
	// In a real test, we would use gorilla/mux's testing utilities
	// For now, we'll mock this functionality by returning the same context
	return ctx
}