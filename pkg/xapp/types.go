package xapp

import (
	"time"
)

// xApp Framework Types according to O-RAN specifications

// XAppStatus represents the lifecycle status of an xApp
type XAppStatus string

const (
	XAppStatusDeployed    XAppStatus = "DEPLOYED"
	XAppStatusRunning     XAppStatus = "RUNNING"
	XAppStatusStopped     XAppStatus = "STOPPED"
	XAppStatusError       XAppStatus = "ERROR"
	XAppStatusTerminating XAppStatus = "TERMINATING"
	XAppStatusUnknown     XAppStatus = "UNKNOWN"
)

// XAppDescriptor describes an xApp and its requirements
type XAppDescriptor struct {
	Name         string                 `json:"name" yaml:"name"`
	Version      string                 `json:"version" yaml:"version"`
	Description  string                 `json:"description" yaml:"description"`
	Namespace    string                 `json:"namespace" yaml:"namespace"`
	Image        string                 `json:"image" yaml:"image"`
	Resources    XAppResources          `json:"resources" yaml:"resources"`
	Config       map[string]interface{} `json:"config" yaml:"config"`
	Interfaces   XAppInterfaces         `json:"interfaces" yaml:"interfaces"`
	Permissions  XAppPermissions        `json:"permissions" yaml:"permissions"`
	Dependencies []XAppDependency       `json:"dependencies" yaml:"dependencies"`
	Metadata     XAppMetadata           `json:"metadata" yaml:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// XAppResources defines resource requirements for an xApp
type XAppResources struct {
	CPU              string            `json:"cpu" yaml:"cpu"`
	Memory           string            `json:"memory" yaml:"memory"`
	Storage          string            `json:"storage" yaml:"storage"`
	GPUCount         int               `json:"gpu_count" yaml:"gpu_count"`
	NetworkBandwidth string            `json:"network_bandwidth" yaml:"network_bandwidth"`
	Labels           map[string]string `json:"labels" yaml:"labels"`
	Limits           ResourceLimits    `json:"limits" yaml:"limits"`
	Requests         ResourceRequests  `json:"requests" yaml:"requests"`
}

// ResourceLimits defines maximum resource usage
type ResourceLimits struct {
	CPU              string `json:"cpu" yaml:"cpu"`
	Memory           string `json:"memory" yaml:"memory"`
	EphemeralStorage string `json:"ephemeral_storage" yaml:"ephemeral_storage"`
}

// ResourceRequests defines minimum resource requirements
type ResourceRequests struct {
	CPU              string `json:"cpu" yaml:"cpu"`
	Memory           string `json:"memory" yaml:"memory"`
	EphemeralStorage string `json:"ephemeral_storage" yaml:"ephemeral_storage"`
}

// XAppInterfaces defines the interfaces used by an xApp
type XAppInterfaces struct {
	E2  E2Interface  `json:"e2" yaml:"e2"`
	A1  A1Interface  `json:"a1" yaml:"a1"`
	O1  O1Interface  `json:"o1" yaml:"o1"`
	RMR RMRInterface `json:"rmr" yaml:"rmr"`
}

// E2Interface defines E2 interface configuration for xApp
type E2Interface struct {
	Enabled           bool     `json:"enabled" yaml:"enabled"`
	RequiredFunctions []string `json:"required_functions" yaml:"required_functions"`
	SubscriptionTypes []string `json:"subscription_types" yaml:"subscription_types"`
	ControlCapability bool     `json:"control_capability" yaml:"control_capability"`
}

// A1Interface defines A1 interface configuration for xApp
type A1Interface struct {
	Enabled     bool     `json:"enabled" yaml:"enabled"`
	PolicyTypes []string `json:"policy_types" yaml:"policy_types"`
}

// O1Interface defines O1 interface configuration for xApp
type O1Interface struct {
	Enabled        bool     `json:"enabled" yaml:"enabled"`
	ManagementPort int      `json:"management_port" yaml:"management_port"`
	YangModels     []string `json:"yang_models" yaml:"yang_models"`
}

// RMRInterface defines RMR messaging configuration
type RMRInterface struct {
	Enabled      bool             `json:"enabled" yaml:"enabled"`
	Port         int              `json:"port" yaml:"port"`
	Routes       []RMRRoute       `json:"routes" yaml:"routes"`
	MessageTypes []RMRMessageType `json:"message_types" yaml:"message_types"`
}

// RMRRoute defines an RMR routing rule
type RMRRoute struct {
	MessageType int    `json:"message_type" yaml:"message_type"`
	Endpoint    string `json:"endpoint" yaml:"endpoint"`
}

// RMRMessageType defines supported RMR message types
type RMRMessageType struct {
	Type        int    `json:"type" yaml:"type"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
}

// XAppPermissions defines security permissions for an xApp
type XAppPermissions struct {
	RBAC      RBACPermissions     `json:"rbac" yaml:"rbac"`
	Network   NetworkPermissions  `json:"network" yaml:"network"`
	Resources ResourcePermissions `json:"resources" yaml:"resources"`
}

// RBACPermissions defines role-based access control permissions
type RBACPermissions struct {
	ClusterRoles []string   `json:"cluster_roles" yaml:"cluster_roles"`
	Roles        []string   `json:"roles" yaml:"roles"`
	Rules        []RBACRule `json:"rules" yaml:"rules"`
}

// RBACRule defines a specific RBAC rule
type RBACRule struct {
	APIGroups []string `json:"api_groups" yaml:"api_groups"`
	Resources []string `json:"resources" yaml:"resources"`
	Verbs     []string `json:"verbs" yaml:"verbs"`
}

// NetworkPermissions defines network access permissions
type NetworkPermissions struct {
	AllowedPorts    []int    `json:"allowed_ports" yaml:"allowed_ports"`
	AllowedHosts    []string `json:"allowed_hosts" yaml:"allowed_hosts"`
	NetworkPolicies []string `json:"network_policies" yaml:"network_policies"`
}

// ResourcePermissions defines resource access permissions
type ResourcePermissions struct {
	PersistentVolumes []string `json:"persistent_volumes" yaml:"persistent_volumes"`
	ConfigMaps        []string `json:"config_maps" yaml:"config_maps"`
	Secrets           []string `json:"secrets" yaml:"secrets"`
}

// XAppDependency defines a dependency on another xApp or service
type XAppDependency struct {
	Name       string                 `json:"name" yaml:"name"`
	Version    string                 `json:"version" yaml:"version"`
	Type       string                 `json:"type" yaml:"type"` // xapp, service, database
	Required   bool                   `json:"required" yaml:"required"`
	Parameters map[string]interface{} `json:"parameters" yaml:"parameters"`
}

// XAppMetadata contains additional metadata about the xApp
type XAppMetadata struct {
	Category    string            `json:"category" yaml:"category"`
	Tags        []string          `json:"tags" yaml:"tags"`
	Author      string            `json:"author" yaml:"author"`
	License     string            `json:"license" yaml:"license"`
	Homepage    string            `json:"homepage" yaml:"homepage"`
	Repository  string            `json:"repository" yaml:"repository"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
}

// XAppInstance represents a running instance of an xApp
type XAppInstance struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Version   string                 `json:"version"`
	Namespace string                 `json:"namespace"`
	Status    XAppStatus             `json:"status"`
	Config    map[string]interface{} `json:"config"`
	Metrics   XAppMetrics            `json:"metrics"`
	Health    XAppHealth             `json:"health"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	LastPing  time.Time              `json:"last_ping"`
}

// XAppMetrics contains runtime metrics for an xApp
type XAppMetrics struct {
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  int64     `json:"memory_usage"`
	NetworkIn    int64     `json:"network_in"`
	NetworkOut   int64     `json:"network_out"`
	RequestCount int64     `json:"request_count"`
	ErrorCount   int64     `json:"error_count"`
	ResponseTime float64   `json:"response_time_ms"`
	LastUpdated  time.Time `json:"last_updated"`
}

// XAppHealth represents health status of an xApp
type XAppHealth struct {
	Status    string                 `json:"status"`
	Checks    map[string]HealthCheck `json:"checks"`
	LastCheck time.Time              `json:"last_check"`
	Uptime    time.Duration          `json:"uptime"`
}

// HealthCheck represents a specific health check
type HealthCheck struct {
	Status    string        `json:"status"`
	Message   string        `json:"message"`
	CheckedAt time.Time     `json:"checked_at"`
	Duration  time.Duration `json:"duration"`
}

// XAppEvent represents events in the xApp lifecycle
type XAppEvent struct {
	ID        string                 `json:"id"`
	XAppID    string                 `json:"xapp_id"`
	EventType XAppEventType          `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Severity  EventSeverity          `json:"severity"`
}

// XAppEventType defines types of xApp events
type XAppEventType string

const (
	XAppEventDeployment     XAppEventType = "DEPLOYMENT"
	XAppEventStartup        XAppEventType = "STARTUP"
	XAppEventShutdown       XAppEventType = "SHUTDOWN"
	XAppEventError          XAppEventType = "ERROR"
	XAppEventHealthCheck    XAppEventType = "HEALTH_CHECK"
	XAppEventConfigChange   XAppEventType = "CONFIG_CHANGE"
	XAppEventResourceChange XAppEventType = "RESOURCE_CHANGE"
)

// EventSeverity defines severity levels for events
type EventSeverity string

const (
	EventSeverityInfo     EventSeverity = "INFO"
	EventSeverityWarning  EventSeverity = "WARNING"
	EventSeverityError    EventSeverity = "ERROR"
	EventSeverityCritical EventSeverity = "CRITICAL"
)

// XAppConflict represents a conflict between xApps
type XAppConflict struct {
	ID           string             `json:"id"`
	XAppID1      string             `json:"xapp_id_1"`
	XAppID2      string             `json:"xapp_id_2"`
	ConflictType ConflictType       `json:"conflict_type"`
	Description  string             `json:"description"`
	Resolution   ConflictResolution `json:"resolution"`
	DetectedAt   time.Time          `json:"detected_at"`
	ResolvedAt   *time.Time         `json:"resolved_at,omitempty"`
	Status       ConflictStatus     `json:"status"`
}

// ConflictType defines types of conflicts between xApps
type ConflictType string

const (
	ConflictTypeResource  ConflictType = "RESOURCE"
	ConflictTypePolicy    ConflictType = "POLICY"
	ConflictTypeFunction  ConflictType = "FUNCTION"
	ConflictTypeInterface ConflictType = "INTERFACE"
)

// ConflictResolution defines how a conflict was resolved
type ConflictResolution string

const (
	ResolutionPriority    ConflictResolution = "PRIORITY"
	ResolutionNegotiation ConflictResolution = "NEGOTIATION"
	ResolutionTermination ConflictResolution = "TERMINATION"
	ResolutionResource    ConflictResolution = "RESOURCE_ALLOCATION"
)

// ConflictStatus represents the status of conflict resolution
type ConflictStatus string

const (
	ConflictStatusDetected  ConflictStatus = "DETECTED"
	ConflictStatusResolving ConflictStatus = "RESOLVING"
	ConflictStatusResolved  ConflictStatus = "RESOLVED"
	ConflictStatusFailed    ConflictStatus = "FAILED"
)

// XAppManager interface defines operations for managing xApps
type XAppManager interface {
	// xApp Lifecycle Management
	Deploy(descriptor *XAppDescriptor, config map[string]interface{}) (*XAppInstance, error)
	Undeploy(xappID string) error
	Start(xappID string) error
	Stop(xappID string) error
	Restart(xappID string) error

	// xApp Information
	List() ([]*XAppInstance, error)
	Get(xappID string) (*XAppInstance, error)
	GetStatus(xappID string) (XAppStatus, error)
	GetHealth(xappID string) (*XAppHealth, error)
	GetMetrics(xappID string) (*XAppMetrics, error)

	// xApp Configuration
	UpdateConfig(xappID string, config map[string]interface{}) error
	GetConfig(xappID string) (map[string]interface{}, error)

	// Conflict Management
	DetectConflicts(xappID string) ([]*XAppConflict, error)
	ResolveConflict(conflictID string, resolution ConflictResolution) error
	GetConflicts() ([]*XAppConflict, error)

	// Event Management
	GetEvents(xappID string) ([]*XAppEvent, error)
	Subscribe(eventTypes []XAppEventType) (<-chan *XAppEvent, error)
}

// XAppRepository interface defines storage operations for xApps
type XAppRepository interface {
	// xApp Descriptor operations
	SaveDescriptor(descriptor *XAppDescriptor) error
	GetDescriptor(name, version string) (*XAppDescriptor, error)
	ListDescriptors() ([]*XAppDescriptor, error)
	DeleteDescriptor(name, version string) error

	// xApp Instance operations
	SaveInstance(instance *XAppInstance) error
	GetInstance(xappID string) (*XAppInstance, error)
	ListInstances() ([]*XAppInstance, error)
	UpdateInstance(instance *XAppInstance) error
	DeleteInstance(xappID string) error

	// Event operations
	SaveEvent(event *XAppEvent) error
	GetEvents(xappID string, limit int) ([]*XAppEvent, error)
	DeleteOldEvents(before time.Time) error

	// Conflict operations
	SaveConflict(conflict *XAppConflict) error
	GetConflict(conflictID string) (*XAppConflict, error)
	ListConflicts() ([]*XAppConflict, error)
	UpdateConflict(conflict *XAppConflict) error
	DeleteConflict(conflictID string) error
}

// XAppOrchestrator interface defines orchestration operations
type XAppOrchestrator interface {
	// Deployment orchestration
	DeployXApp(descriptor *XAppDescriptor, config map[string]interface{}) error
	UndeployXApp(xappID string) error
	ScaleXApp(xappID string, replicas int) error

	// Resource management
	AllocateResources(xappID string, resources *XAppResources) error
	DeallocateResources(xappID string) error
	GetResourceUsage(xappID string) (*XAppResources, error)

	// Health monitoring
	CheckHealth(xappID string) (*XAppHealth, error)
	StartHealthMonitoring(xappID string) error
	StopHealthMonitoring(xappID string) error
}

// XAppSDK interface defines the SDK for xApp development
type XAppSDK interface {
	// RIC Interface access
	GetE2Client() E2Client
	GetA1Client() A1Client
	GetO1Client() O1Client
	GetRMRClient() RMRClient

	// Configuration management
	GetConfig() map[string]interface{}
	WatchConfig(callback func(map[string]interface{})) error

	// Metrics and health
	ReportMetrics(metrics *XAppMetrics) error
	SetHealthStatus(status string) error
	RegisterHealthCheck(name string, check func() HealthCheck) error

	// Event publishing
	PublishEvent(eventType XAppEventType, message string, data map[string]interface{}) error

	// Lifecycle hooks
	RegisterStartupHook(hook func() error) error
	RegisterShutdownHook(hook func() error) error
}

// Interface clients for xApp SDK

type E2Client interface {
	Subscribe(functionID int, eventTrigger []byte, actions []int) error
	Unsubscribe(subscriptionID string) error
	SendControl(functionID int, controlHeader, controlMessage []byte) error
}

type A1Client interface {
	GetPolicyTypes() ([]string, error)
	GetPolicy(policyID string) (map[string]interface{}, error)
	ReportPolicyStatus(policyID string, status string) error
}

type O1Client interface {
	GetConfig() (map[string]interface{}, error)
	UpdateConfig(config map[string]interface{}) error
	SendNotification(notification map[string]interface{}) error
}

type RMRClient interface {
	Send(messageType int, payload []byte, target string) error
	Receive() (<-chan RMRMessage, error)
	RegisterHandler(messageType int, handler func(RMRMessage)) error
}

type RMRMessage struct {
	Type      int               `json:"type"`
	Payload   []byte            `json:"payload"`
	Source    string            `json:"source"`
	Headers   map[string]string `json:"headers"`
	Timestamp time.Time         `json:"timestamp"`
}

// XAppConfigurationManager manages xApp configuration
type XAppConfigurationManager interface {
	ValidateConfig(descriptor *XAppDescriptor, config map[string]interface{}) error
	MergeConfigs(base, override map[string]interface{}) map[string]interface{}
	EncryptSensitiveData(config map[string]interface{}) (map[string]interface{}, error)
	DecryptSensitiveData(config map[string]interface{}) (map[string]interface{}, error)
}

// XAppRegistry manages xApp discovery and registration
type XAppRegistry interface {
	Register(instance *XAppInstance) error
	Unregister(xappID string) error
	Discover(criteria map[string]interface{}) ([]*XAppInstance, error)
	GetEndpoints(xappID string) (map[string]string, error)
	UpdateHeartbeat(xappID string) error
}

// Configuration for xApp framework
type XAppFrameworkConfig struct {
	KubernetesConfig    string        `json:"kubernetes_config"`
	Namespace           string        `json:"namespace"`
	RegistryURL         string        `json:"registry_url"`
	ImagePullSecrets    []string      `json:"image_pull_secrets"`
	DefaultResources    XAppResources `json:"default_resources"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	MetricsInterval     time.Duration `json:"metrics_interval"`
	EventRetention      time.Duration `json:"event_retention"`
	ConflictDetection   bool          `json:"conflict_detection"`
	AutoScale           bool          `json:"auto_scale"`
	SecurityEnabled     bool          `json:"security_enabled"`
}
