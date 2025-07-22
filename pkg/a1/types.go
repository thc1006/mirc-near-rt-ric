package a1

import (
	"time"
)

// A1 Interface Types according to O-RAN A1 specification
// Based on O-RAN.WG2.A1-v06.00

// PolicyTypeID represents a policy type identifier
type PolicyTypeID string

// PolicyID represents a policy instance identifier
type PolicyID string

// PolicyStatus represents the status of a policy instance
type PolicyStatus string

const (
	PolicyStatusNotEnforced PolicyStatus = "NOT_ENFORCED"
	PolicyStatusEnforced    PolicyStatus = "ENFORCED"
	PolicyStatusDeleted     PolicyStatus = "DELETED"
)

// PolicyType defines a policy type with its schema and metadata
type PolicyType struct {
	PolicyTypeID   PolicyTypeID `json:"policy_type_id"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	PolicySchema   interface{}  `json:"policy_schema"`
	CreateSchema   interface{}  `json:"create_schema,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// PolicyInstance represents an instance of a policy
type PolicyInstance struct {
	PolicyID        PolicyID      `json:"policy_id"`
	PolicyTypeID    PolicyTypeID  `json:"policy_type_id"`
	PolicyData      interface{}   `json:"policy_data"`
	Status          PolicyStatus  `json:"status"`
	StatusReason    string        `json:"status_reason,omitempty"`
	TargetNearRTRIC string        `json:"target_near_rt_ric,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	EnforcedAt      *time.Time    `json:"enforced_at,omitempty"`
}

// PolicyStatusInfo contains detailed status information for a policy
type PolicyStatusInfo struct {
	PolicyID        PolicyID     `json:"policy_id"`
	PolicyTypeID    PolicyTypeID `json:"policy_type_id"`
	Status          PolicyStatus `json:"status"`
	StatusReason    string       `json:"status_reason,omitempty"`
	LastUpdate      time.Time    `json:"last_update"`
	EnforcementInfo interface{}  `json:"enforcement_info,omitempty"`
}

// EnrichmentInfo represents enrichment information for ML models
type EnrichmentInfo struct {
	EIJobID     string      `json:"ei_job_id"`
	EITypeID    string      `json:"ei_type_id"`
	JobOwner    string      `json:"job_owner"`
	JobData     interface{} `json:"job_data"`
	TargetURI   string      `json:"target_uri"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Status      string      `json:"status"`
}

// EIType represents an enrichment information type
type EIType struct {
	EITypeID    string      `json:"ei_type_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	JobSchema   interface{} `json:"job_schema"`
	CreatedAt   time.Time   `json:"created_at"`
}

// MLModel represents an ML model managed through A1
type MLModel struct {
	ModelID       string                 `json:"model_id"`
	ModelName     string                 `json:"model_name"`
	ModelVersion  string                 `json:"model_version"`
	ModelType     string                 `json:"model_type"`
	Description   string                 `json:"description"`
	ModelData     []byte                 `json:"model_data,omitempty"`
	ModelURL      string                 `json:"model_url,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Status        string                 `json:"status"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	DeployedAt    *time.Time             `json:"deployed_at,omitempty"`
}

// A1ErrorResponse represents an error response from A1 interface
type A1ErrorResponse struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance,omitempty"`
}

// PolicyTypeCreateRequest represents a request to create a new policy type
type PolicyTypeCreateRequest struct {
	PolicyTypeID PolicyTypeID `json:"policy_type_id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	PolicySchema interface{}  `json:"policy_schema"`
	CreateSchema interface{}  `json:"create_schema,omitempty"`
}

// PolicyInstanceCreateRequest represents a request to create a policy instance
type PolicyInstanceCreateRequest struct {
	PolicyData      interface{} `json:"policy_data"`
	TargetNearRTRIC string      `json:"target_near_rt_ric,omitempty"`
}

// PolicyInstanceUpdateRequest represents a request to update a policy instance
type PolicyInstanceUpdateRequest struct {
	PolicyData      interface{} `json:"policy_data"`
	TargetNearRTRIC string      `json:"target_near_rt_ric,omitempty"`
}

// EIJobCreateRequest represents a request to create an enrichment information job
type EIJobCreateRequest struct {
	EITypeID    string      `json:"ei_type_id"`
	JobOwner    string      `json:"job_owner"`
	JobData     interface{} `json:"job_data"`
	TargetURI   string      `json:"target_uri"`
}

// MLModelDeployRequest represents a request to deploy an ML model
type MLModelDeployRequest struct {
	ModelName     string                 `json:"model_name"`
	ModelVersion  string                 `json:"model_version"`
	ModelType     string                 `json:"model_type"`
	Description   string                 `json:"description"`
	ModelData     []byte                 `json:"model_data,omitempty"`
	ModelURL      string                 `json:"model_url,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	DeployTarget  string                 `json:"deploy_target,omitempty"`
}

// A1HealthCheck represents the health status of the A1 interface
type A1HealthCheck struct {
	Status      string            `json:"status"`
	Timestamp   time.Time         `json:"timestamp"`
	Version     string            `json:"version"`
	Components  map[string]string `json:"components"`
	Uptime      time.Duration     `json:"uptime"`
}

// A1Statistics contains statistics for the A1 interface
type A1Statistics struct {
	TotalPolicyTypes     int           `json:"total_policy_types"`
	TotalPolicyInstances int           `json:"total_policy_instances"`
	EnforcedPolicies     int           `json:"enforced_policies"`
	FailedPolicies       int           `json:"failed_policies"`
	TotalRequests        int64         `json:"total_requests"`
	SuccessfulRequests   int64         `json:"successful_requests"`
	FailedRequests       int64         `json:"failed_requests"`
	AverageResponseTime  time.Duration `json:"average_response_time"`
	Uptime              time.Duration `json:"uptime"`
	LastUpdate          time.Time     `json:"last_update"`
}

// PolicyEvent represents events related to policy lifecycle
type PolicyEvent struct {
	EventID      string                 `json:"event_id"`
	EventType    string                 `json:"event_type"`
	PolicyID     PolicyID               `json:"policy_id"`
	PolicyTypeID PolicyTypeID           `json:"policy_type_id"`
	Timestamp    time.Time              `json:"timestamp"`
	UserID       string                 `json:"user_id,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	OldStatus    PolicyStatus           `json:"old_status,omitempty"`
	NewStatus    PolicyStatus           `json:"new_status,omitempty"`
}

// Policy event types
const (
	PolicyEventCreated  = "POLICY_CREATED"
	PolicyEventUpdated  = "POLICY_UPDATED"
	PolicyEventDeleted  = "POLICY_DELETED"
	PolicyEventEnforced = "POLICY_ENFORCED"
	PolicyEventFailed   = "POLICY_FAILED"
)

// String methods for type safety and debugging

func (pt PolicyTypeID) String() string {
	return string(pt)
}

func (pi PolicyID) String() string {
	return string(pi)
}

func (ps PolicyStatus) String() string {
	return string(ps)
}

// IsValid checks if a policy status is valid
func (ps PolicyStatus) IsValid() bool {
	switch ps {
	case PolicyStatusNotEnforced, PolicyStatusEnforced, PolicyStatusDeleted:
		return true
	default:
		return false
	}
}

// PolicyTypeExists checks if a policy type ID is valid format
func (pt PolicyTypeID) IsValid() bool {
	return len(string(pt)) > 0 && len(string(pt)) <= 100
}

// PolicyIDExists checks if a policy ID is valid format
func (pi PolicyID) IsValid() bool {
	return len(string(pi)) > 0 && len(string(pi)) <= 100
}