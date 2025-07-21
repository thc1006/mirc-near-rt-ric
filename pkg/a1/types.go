package a1

import (
	"time"
)

// A1 Interface Types according to O-RAN A1AP specification

// A1PolicyType represents a policy type definition
type A1PolicyType struct {
	PolicyTypeID   string                 `json:"policy_type_id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	PolicySchema   map[string]interface{} `json:"policy_schema"`
	CreatedAt      time.Time              `json:"created_at"`
	LastModified   time.Time              `json:"last_modified"`
}

// A1Policy represents an A1 policy instance
type A1Policy struct {
	PolicyID       string                 `json:"policy_id"`
	PolicyTypeID   string                 `json:"policy_type_id"`
	RICRequestorID string                 `json:"ric_requestor_id"`
	PolicyData     map[string]interface{} `json:"policy_data"`
	Status         A1PolicyStatus         `json:"status"`
	CreatedAt      time.Time              `json:"created_at"`
	LastModified   time.Time              `json:"last_modified"`
	NotificationURI string                `json:"notification_uri,omitempty"`
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
	PolicyTypeID    string   `json:"policy_type_id"`
	NumberOfPolicies int     `json:"num_instances"`
	PolicyInstanceIDs []string `json:"policy_instance_list"`
}

// A1HealthStatus represents the health status of the A1 interface
type A1HealthStatus struct {
	Status          string    `json:"status"`
	LastUpdate      time.Time `json:"last_update"`
	Version         string    `json:"version"`
	PolicyTypes     int       `json:"policy_types_count"`
	Policies        int       `json:"policies_count"`
	ActivePolicies  int       `json:"active_policies_count"`
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
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Timestamp    time.Time `json:"timestamp"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// Common error codes
const (
	A1ErrorNotFound           = "NOT_FOUND"
	A1ErrorBadRequest         = "BAD_REQUEST"
	A1ErrorInternalError      = "INTERNAL_ERROR"
	A1ErrorConflict          = "CONFLICT"
	A1ErrorUnauthorized      = "UNAUTHORIZED"
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
	RequestID    string                 `json:"request_id"`
	StatusCode   int                    `json:"status_code"`
	Headers      map[string]string      `json:"headers"`
	Body         map[string]interface{} `json:"body,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	ProcessingTime time.Duration        `json:"processing_time"`
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
	ListenAddress        string        `json:"listen_address"`
	ListenPort           int           `json:"listen_port"`
	TLSEnabled           bool          `json:"tls_enabled"`
	TLSCertPath          string        `json:"tls_cert_path"`
	TLSKeyPath           string        `json:"tls_key_path"`
	AuthenticationEnabled bool         `json:"authentication_enabled"`
	RequestTimeout       time.Duration `json:"request_timeout"`
	MaxRequestSize       int64         `json:"max_request_size"`
	RateLimitEnabled     bool          `json:"rate_limit_enabled"`
	RateLimitPerMinute   int           `json:"rate_limit_per_minute"`
	NotificationEnabled  bool          `json:"notification_enabled"`
	DatabaseURL          string        `json:"database_url"`
	LogLevel             string        `json:"log_level"`
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
	EnforcementLatency   time.Duration `json:"enforcement_latency"`
	SuccessRate          float64       `json:"success_rate"`
	ErrorRate            float64       `json:"error_rate"`
	LastEnforcementTime  time.Time     `json:"last_enforcement_time"`
	TotalEnforcements    int64         `json:"total_enforcements"`
	FailedEnforcements   int64         `json:"failed_enforcements"`
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
			PolicyTypeID:   "o-ran-qos-policy-v1.0",
			Name:           "QoS Management Policy",
			Description:    "Policy for managing Quality of Service parameters",
			PolicySchema:   QoSPolicyTypeSchema,
			CreatedAt:      now,
			LastModified:   now,
		},
		{
			PolicyTypeID:   "o-ran-rrm-policy-v1.0",
			Name:           "Radio Resource Management Policy",
			Description:    "Policy for managing radio resource allocation",
			PolicySchema:   RRMPolicyTypeSchema,
			CreatedAt:      now,
			LastModified:   now,
		},
		{
			PolicyTypeID:   "o-ran-son-policy-v1.0",
			Name:           "Self-Organizing Networks Policy",
			Description:    "Policy for SON function optimization",
			PolicySchema:   SONPolicyTypeSchema,
			CreatedAt:      now,
			LastModified:   now,
		},
	}
}