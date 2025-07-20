// Copyright 2024 The O-RAN Near-RT RIC Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package federatedlearning

import (
	"context"
	"crypto/ed25519"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RRMTaskType defines the type of Radio Resource Management task
type RRMTaskType string

const (
	// Resource allocation and optimization tasks
	RRMTaskResourceAllocation  RRMTaskType = "resource_allocation"
	RRMTaskPowerControl       RRMTaskType = "power_control" 
	RRMTaskSpectrumManagement RRMTaskType = "spectrum_management"
	RRMTaskNetworkSlicing     RRMTaskType = "network_slicing"
	
	// Traffic and QoS management tasks
	RRMTaskTrafficPrediction  RRMTaskType = "traffic_prediction"
	RRMTaskQoSOptimization    RRMTaskType = "qos_optimization"
	RRMTaskLoadBalancing      RRMTaskType = "load_balancing"
	
	// Mobility and handover management tasks
	RRMTaskHandoverOptimization RRMTaskType = "handover_optimization"
	RRMTaskMobilityPrediction   RRMTaskType = "mobility_prediction"
	RRMTaskCellSelection        RRMTaskType = "cell_selection"
	
	// Anomaly detection and security tasks
	RRMTaskAnomalyDetection     RRMTaskType = "anomaly_detection"
	RRMTaskInterferenceManagement RRMTaskType = "interference_management"
)

// SecurityAction defines actions to take for security threats
type SecurityAction string

const (
	SecurityActionAllow      SecurityAction = "allow"
	SecurityActionMonitor    SecurityAction = "monitor"
	SecurityActionMitigate   SecurityAction = "mitigate"
	SecurityActionQuarantine SecurityAction = "quarantine"
	SecurityActionRemove     SecurityAction = "remove"
	SecurityActionBlock      SecurityAction = "block"
)

// RiskLevel defines severity levels for security risks
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// ThreatLevel defines overall threat assessment levels
type ThreatLevel string

const (
	ThreatLevelMinimal   ThreatLevel = "minimal"
	ThreatLevelLow       ThreatLevel = "low"
	ThreatLevelModerate  ThreatLevel = "moderate"
	ThreatLevelHigh      ThreatLevel = "high"
	ThreatLevelSevere    ThreatLevel = "severe"
	ThreatLevelCritical  ThreatLevel = "critical"
)

// SecurityFinding represents a specific security issue or observation
type SecurityFinding struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Severity    RiskLevel              `json:"severity"`
	Category    string                 `json:"category"`
	Evidence    map[string]interface{} `json:"evidence"`
	Timestamp   time.Time              `json:"timestamp"`
	Remediation string                 `json:"remediation"`
}

// MonitoringRequest defines monitoring requirements
type MonitoringRequest struct {
	TargetID    string        `json:"target_id"`
	Metrics     []string      `json:"metrics"`
	Frequency   time.Duration `json:"frequency"`
	Duration    time.Duration `json:"duration"`
	Thresholds  map[string]float64 `json:"thresholds"`
	AlertRules  []string      `json:"alert_rules"`
}

// ClientTrainingHistory stores historical training data for clients
type ClientTrainingHistory struct {
	ClientID           string                    `json:"client_id"`
	RegistrationTime   time.Time                 `json:"registration_time"`
	LastActivityTime   time.Time                 `json:"last_activity_time"`
	TotalRounds        int                       `json:"total_rounds"`
	SuccessfulRounds   int                       `json:"successful_rounds"`
	FailedRounds       int                       `json:"failed_rounds"`
	AverageLatency     time.Duration             `json:"average_latency"`
	DataQuality        float64                   `json:"data_quality"`
	ModelAccuracy      float64                   `json:"model_accuracy"`
	ReputationScore    float64                   `json:"reputation_score"`
	TrustScore         float64                   `json:"trust_score"`
	BehaviorPattern    *BehaviorPattern          `json:"behavior_pattern"`
	SecurityIncidents  []*SecurityIncident       `json:"security_incidents"`
	PerformanceMetrics *ClientPerformanceMetrics `json:"performance_metrics"`
}

// SecurityIncident represents a security-related incident
type SecurityIncident struct {
	ID            string                 `json:"id"`
	ClientID      string                 `json:"client_id"`
	IncidentType  string                 `json:"incident_type"`
	Severity      RiskLevel              `json:"severity"`
	Description   string                 `json:"description"`
	Timestamp     time.Time              `json:"timestamp"`
	Resolved      bool                   `json:"resolved"`
	ResolutionTime *time.Time            `json:"resolution_time,omitempty"`
	Evidence      map[string]interface{} `json:"evidence"`
	Actions       []SecurityAction       `json:"actions"`
}

// ClientPerformanceMetrics tracks client performance over time
type ClientPerformanceMetrics struct {
	AverageLatency       time.Duration `json:"average_latency"`
	P95Latency          time.Duration `json:"p95_latency"`
	P99Latency          time.Duration `json:"p99_latency"`
	Throughput          float64       `json:"throughput"`
	ErrorRate           float64       `json:"error_rate"`
	Availability        float64       `json:"availability"`
	ResourceUtilization float64       `json:"resource_utilization"`
	DataQualityScore    float64       `json:"data_quality_score"`
	ModelConvergence    float64       `json:"model_convergence"`
}

// FLClientStatus represents the current status of a federated learning client
type FLClientStatus string

const (
	FLClientStatusRegistered   FLClientStatus = "registered"
	FLClientStatusTraining     FLClientStatus = "training"
	FLClientStatusIdle         FLClientStatus = "idle"
	FLClientStatusDisconnected FLClientStatus = "disconnected"
	FLClientStatusFailed       FLClientStatus = "failed"
)

// ModelFormat defines the format/framework of the ML model
type ModelFormat string

const (
	ModelFormatTensorFlow   ModelFormat = "tensorflow"
	ModelFormatPyTorch      ModelFormat = "pytorch"
	ModelFormatONNX         ModelFormat = "onnx"
	ModelFormatSKLearn      ModelFormat = "sklearn"
	ModelFormatCustom       ModelFormat = "custom"
)

// AggregationAlgorithm defines the federated aggregation algorithm
type AggregationAlgorithm string

const (
	AggregationFedAvg         AggregationAlgorithm = "fedavg"
	AggregationFedProx        AggregationAlgorithm = "fedprox"
	AggregationFedNova        AggregationAlgorithm = "fednova"
	AggregationSecureAgg      AggregationAlgorithm = "secure_aggregation"
	AggregationByzantineFT    AggregationAlgorithm = "byzantine_ft"
)

// PrivacyMechanism defines the privacy preservation mechanism
type PrivacyMechanism string

const (
	PrivacyDifferential       PrivacyMechanism = "differential_privacy"
	PrivacyHomomorphic        PrivacyMechanism = "homomorphic_encryption"
	PrivacySecureMultiparty   PrivacyMechanism = "secure_multiparty"
	PrivacyLocalDP            PrivacyMechanism = "local_differential_privacy"
)

// FLClient represents a federated learning client (xApp)
type FLClient struct {
	// Identity and registration
	ID           string    `json:"id" bson:"_id"`
	XAppName     string    `json:"xapp_name" bson:"xapp_name"`
	XAppVersion  string    `json:"xapp_version" bson:"xapp_version"`
	Namespace    string    `json:"namespace" bson:"namespace"`
	
	// Network and communication
	Endpoint     string    `json:"endpoint" bson:"endpoint"`
	PublicKey    ed25519.PublicKey `json:"public_key" bson:"public_key"`
	Certificate  []byte    `json:"certificate" bson:"certificate"`
	
	// Capabilities and configuration
	RRMTasks     []RRMTaskType `json:"rrm_tasks" bson:"rrm_tasks"`
	ModelFormats []ModelFormat `json:"model_formats" bson:"model_formats"`
	ComputeResources ComputeCapabilities `json:"compute_resources" bson:"compute_resources"`
	
	// Status and metrics
	Status           FLClientStatus `json:"status" bson:"status"`
	LastHeartbeat    time.Time     `json:"last_heartbeat" bson:"last_heartbeat"`
	LastTrainingTime time.Time     `json:"last_training_time" bson:"last_training_time"`
	TotalRounds      int64         `json:"total_rounds" bson:"total_rounds"`
	SuccessfulRounds int64         `json:"successful_rounds" bson:"successful_rounds"`
	
	// Privacy and security
	PrivacyBudget    float64           `json:"privacy_budget" bson:"privacy_budget"`
	TrustScore       float64           `json:"trust_score" bson:"trust_score"`
	SecurityFeatures []SecurityFeature `json:"security_features" bson:"security_features"`
	
	// Network slice association
	NetworkSlices []NetworkSliceInfo `json:"network_slices" bson:"network_slices"`
	
	// Metadata
	RegisteredAt time.Time         `json:"registered_at" bson:"registered_at"`
	UpdatedAt    time.Time         `json:"updated_at" bson:"updated_at"`
	Metadata     map[string]string `json:"metadata" bson:"metadata"`
}

// ComputeCapabilities represents the computational resources of a client
type ComputeCapabilities struct {
	CPUCores        int     `json:"cpu_cores" bson:"cpu_cores"`
	MemoryGB        float64 `json:"memory_gb" bson:"memory_gb"`
	GPUCount        int     `json:"gpu_count" bson:"gpu_count"`
	GPUMemoryGB     float64 `json:"gpu_memory_gb" bson:"gpu_memory_gb"`
	StorageGB       float64 `json:"storage_gb" bson:"storage_gb"`
	NetworkBandwidth int64  `json:"network_bandwidth_mbps" bson:"network_bandwidth_mbps"`
	
	// Performance metrics
	FLOPSCapacity   int64   `json:"flops_capacity" bson:"flops_capacity"`
	LatencyProfile  LatencyProfile `json:"latency_profile" bson:"latency_profile"`
}

// LatencyProfile represents latency characteristics for RRM requirements
type LatencyProfile struct {
	ComputeLatencyMs    float64 `json:"compute_latency_ms" bson:"compute_latency_ms"`
	CommunicationLatencyMs float64 `json:"communication_latency_ms" bson:"communication_latency_ms"`
	E2LatencyMs         float64 `json:"e2_latency_ms" bson:"e2_latency_ms"`
	TargetLatencyMs     float64 `json:"target_latency_ms" bson:"target_latency_ms"`
}

// SecurityFeature represents security capabilities of a client
type SecurityFeature struct {
	Name        string            `json:"name" bson:"name"`
	Version     string            `json:"version" bson:"version"`
	Enabled     bool              `json:"enabled" bson:"enabled"`
	Parameters  map[string]string `json:"parameters" bson:"parameters"`
}

// NetworkSliceInfo represents network slice association
type NetworkSliceInfo struct {
	SliceID     string  `json:"slice_id" bson:"slice_id"`
	SliceType   string  `json:"slice_type" bson:"slice_type"`
	Priority    int     `json:"priority" bson:"priority"`
	Bandwidth   int64   `json:"bandwidth_mbps" bson:"bandwidth_mbps"`
	Latency     float64 `json:"latency_ms" bson:"latency_ms"`
	Reliability float64 `json:"reliability" bson:"reliability"`
}

// GlobalModel represents the global federated learning model
type GlobalModel struct {
	// Model identification
	ID           string      `json:"id" bson:"_id"`
	Name         string      `json:"name" bson:"name"`
	Version      string      `json:"version" bson:"version"`
	RRMTask      RRMTaskType `json:"rrm_task" bson:"rrm_task"`
	
	// Model specification
	Format       ModelFormat `json:"format" bson:"format"`
	Architecture string      `json:"architecture" bson:"architecture"`
	Parameters   []byte      `json:"parameters" bson:"parameters"`
	ParametersHash string    `json:"parameters_hash" bson:"parameters_hash"`
	
	// Training configuration
	TrainingConfig TrainingConfiguration `json:"training_config" bson:"training_config"`
	
	// Aggregation and privacy
	AggregationAlg  AggregationAlgorithm `json:"aggregation_algorithm" bson:"aggregation_algorithm"`
	PrivacyMech     PrivacyMechanism     `json:"privacy_mechanism" bson:"privacy_mechanism"`
	PrivacyParams   PrivacyParameters    `json:"privacy_params" bson:"privacy_params"`
	
	// Training progress
	CurrentRound    int64     `json:"current_round" bson:"current_round"`
	MaxRounds       int64     `json:"max_rounds" bson:"max_rounds"`
	TargetAccuracy  float64   `json:"target_accuracy" bson:"target_accuracy"`
	CurrentAccuracy float64   `json:"current_accuracy" bson:"current_accuracy"`
	
	// Performance metrics
	ModelMetrics    ModelMetrics  `json:"model_metrics" bson:"model_metrics"`
	TrainingHistory []RoundMetrics `json:"training_history" bson:"training_history"`
	
	// Status and lifecycle
	Status      TrainingStatus `json:"status" bson:"status"`
	CreatedAt   time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" bson:"updated_at"`
	StartedAt   *time.Time    `json:"started_at,omitempty" bson:"started_at,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty" bson:"completed_at,omitempty"`
	
	// Access control
	AllowedClients []string          `json:"allowed_clients" bson:"allowed_clients"`
	NetworkSlices  []string          `json:"network_slices" bson:"network_slices"`
	Metadata       map[string]string `json:"metadata" bson:"metadata"`
}

// TrainingConfiguration defines the federated learning training parameters
type TrainingConfiguration struct {
	BatchSize          int     `json:"batch_size" bson:"batch_size"`
	LocalEpochs        int     `json:"local_epochs" bson:"local_epochs"`
	LearningRate       float64 `json:"learning_rate" bson:"learning_rate"`
	MinParticipants    int     `json:"min_participants" bson:"min_participants"`
	MaxParticipants    int     `json:"max_participants" bson:"max_participants"`
	SelectionStrategy  string  `json:"selection_strategy" bson:"selection_strategy"`
	ConvergenceThreshold float64 `json:"convergence_threshold" bson:"convergence_threshold"`
	TimeoutSeconds     int64   `json:"timeout_seconds" bson:"timeout_seconds"`
	
	// RRM-specific parameters
	E2LatencyRequirement  float64 `json:"e2_latency_requirement_ms" bson:"e2_latency_requirement_ms"`
	ResourceConstraints   ResourceConstraints `json:"resource_constraints" bson:"resource_constraints"`
	QualityRequirements   QualityRequirements `json:"quality_requirements" bson:"quality_requirements"`
}

// ResourceConstraints defines resource limitations for training
type ResourceConstraints struct {
	MaxCPUUsage     float64 `json:"max_cpu_usage" bson:"max_cpu_usage"`
	MaxMemoryUsage  float64 `json:"max_memory_usage_gb" bson:"max_memory_usage_gb"`
	MaxNetworkUsage int64   `json:"max_network_usage_mbps" bson:"max_network_usage_mbps"`
	PowerBudget     float64 `json:"power_budget_watts" bson:"power_budget_watts"`
}

// QualityRequirements defines minimum quality standards
type QualityRequirements struct {
	MinAccuracy     float64 `json:"min_accuracy" bson:"min_accuracy"`
	MaxLoss         float64 `json:"max_loss" bson:"max_loss"`
	StabilityMetric float64 `json:"stability_metric" bson:"stability_metric"`
	RobustnessScore float64 `json:"robustness_score" bson:"robustness_score"`
}

// PrivacyParameters defines privacy preservation parameters
type PrivacyParameters struct {
	// Differential Privacy
	Epsilon         float64 `json:"epsilon,omitempty" bson:"epsilon,omitempty"`
	Delta           float64 `json:"delta,omitempty" bson:"delta,omitempty"`
	NoiseMultiplier float64 `json:"noise_multiplier,omitempty" bson:"noise_multiplier,omitempty"`
	
	// Homomorphic Encryption
	EncryptionScheme string `json:"encryption_scheme,omitempty" bson:"encryption_scheme,omitempty"`
	KeySize          int    `json:"key_size,omitempty" bson:"key_size,omitempty"`
	
	// Secure Aggregation
	ThresholdScheme string `json:"threshold_scheme,omitempty" bson:"threshold_scheme,omitempty"`
	ThresholdK      int    `json:"threshold_k,omitempty" bson:"threshold_k,omitempty"`
	ThresholdN      int    `json:"threshold_n,omitempty" bson:"threshold_n,omitempty"`
}

// TrainingStatus represents the current status of model training
type TrainingStatus string

const (
	TrainingStatusInitializing TrainingStatus = "initializing"
	TrainingStatusWaiting      TrainingStatus = "waiting"
	TrainingStatusRunning      TrainingStatus = "running"
	TrainingStatusAggregating  TrainingStatus = "aggregating"
	TrainingStatusCompleted    TrainingStatus = "completed"
	TrainingStatusFailed       TrainingStatus = "failed"
	TrainingStatusPaused       TrainingStatus = "paused"
)

// ModelMetrics represents comprehensive model performance metrics
type ModelMetrics struct {
	// Primary metrics
	Accuracy    float64 `json:"accuracy" bson:"accuracy"`
	Loss        float64 `json:"loss" bson:"loss"`
	F1Score     float64 `json:"f1_score" bson:"f1_score"`
	Precision   float64 `json:"precision" bson:"precision"`
	Recall      float64 `json:"recall" bson:"recall"`
	
	// RRM-specific metrics
	SpectrumEfficiency    float64 `json:"spectrum_efficiency" bson:"spectrum_efficiency"`
	EnergyEfficiency      float64 `json:"energy_efficiency" bson:"energy_efficiency"`
	ThroughputImprovement float64 `json:"throughput_improvement" bson:"throughput_improvement"`
	LatencyReduction      float64 `json:"latency_reduction" bson:"latency_reduction"`
	HandoverSuccessRate   float64 `json:"handover_success_rate" bson:"handover_success_rate"`
	
	// Robustness metrics
	AdversarialRobustness float64 `json:"adversarial_robustness" bson:"adversarial_robustness"`
	FairnessMetric        float64 `json:"fairness_metric" bson:"fairness_metric"`
	PrivacyLeakage        float64 `json:"privacy_leakage" bson:"privacy_leakage"`
	
	// Convergence metrics
	ConvergenceRate       float64 `json:"convergence_rate" bson:"convergence_rate"`
	StabilityIndex        float64 `json:"stability_index" bson:"stability_index"`
	CommunicationCost     int64   `json:"communication_cost_bytes" bson:"communication_cost_bytes"`
}

// RoundMetrics represents metrics for a specific training round
type RoundMetrics struct {
	Round             int64         `json:"round" bson:"round"`
	Timestamp         time.Time     `json:"timestamp" bson:"timestamp"`
	ParticipatingClients int        `json:"participating_clients" bson:"participating_clients"`
	ModelMetrics      ModelMetrics  `json:"model_metrics" bson:"model_metrics"`
	AggregationTimeMs float64       `json:"aggregation_time_ms" bson:"aggregation_time_ms"`
	CommunicationCost int64         `json:"communication_cost_bytes" bson:"communication_cost_bytes"`
	PrivacyBudgetUsed float64       `json:"privacy_budget_used" bson:"privacy_budget_used"`
	
	// Client-specific metrics
	ClientMetrics     []ClientRoundMetrics `json:"client_metrics" bson:"client_metrics"`
	DroppedClients    []string            `json:"dropped_clients" bson:"dropped_clients"`
	MaliciousActivity []SecurityIncident  `json:"malicious_activity" bson:"malicious_activity"`
}

// ClientRoundMetrics represents per-client metrics for a training round
type ClientRoundMetrics struct {
	ClientID          string    `json:"client_id" bson:"client_id"`
	DataSamplesCount  int64     `json:"data_samples_count" bson:"data_samples_count"`
	TrainingTimeMs    float64   `json:"training_time_ms" bson:"training_time_ms"`
	UploadTimeMs      float64   `json:"upload_time_ms" bson:"upload_time_ms"`
	LocalAccuracy     float64   `json:"local_accuracy" bson:"local_accuracy"`
	LocalLoss         float64   `json:"local_loss" bson:"local_loss"`
	ModelSize         int64     `json:"model_size_bytes" bson:"model_size_bytes"`
	PrivacyNoiseLevel float64   `json:"privacy_noise_level" bson:"privacy_noise_level"`
	ResourceUsage     ResourceUsageMetrics `json:"resource_usage" bson:"resource_usage"`
}

// ResourceUsageMetrics represents resource consumption during training
type ResourceUsageMetrics struct {
	CPUUsagePercent    float64 `json:"cpu_usage_percent" bson:"cpu_usage_percent"`
	MemoryUsageGB      float64 `json:"memory_usage_gb" bson:"memory_usage_gb"`
	GPUUsagePercent    float64 `json:"gpu_usage_percent" bson:"gpu_usage_percent"`
	NetworkUsageMbps   float64 `json:"network_usage_mbps" bson:"network_usage_mbps"`
	PowerConsumptionW  float64 `json:"power_consumption_watts" bson:"power_consumption_watts"`
}

// SecurityIncident definition moved above to avoid duplication

// TrainingJob represents a federated learning training job
type TrainingJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	
	Spec   TrainingJobSpec   `json:"spec,omitempty"`
	Status TrainingJobStatus `json:"status,omitempty"`
}

// TrainingJobSpec defines the desired state of a training job
type TrainingJobSpec struct {
	ModelID         string                `json:"modelId"`
	RRMTask         RRMTaskType          `json:"rrmTask"`
	TrainingConfig  TrainingConfiguration `json:"trainingConfig"`
	ClientSelector  ClientSelector        `json:"clientSelector"`
	PrivacyConfig   PrivacyParameters     `json:"privacyConfig"`
	
	// Scheduling
	Schedule        string    `json:"schedule,omitempty"`        // Cron expression
	Deadline        *time.Time `json:"deadline,omitempty"`       // Maximum completion time
	Priority        int32     `json:"priority,omitempty"`        // Job priority
	
	// Resource management
	Resources       ResourceRequirements  `json:"resources,omitempty"`
	NetworkSlices   []string             `json:"networkSlices,omitempty"`
}

// ClientSelector defines criteria for selecting participating clients
type ClientSelector struct {
	MatchLabels    map[string]string `json:"matchLabels,omitempty"`
	MatchRRMTasks  []RRMTaskType     `json:"matchRRMTasks,omitempty"`
	MinTrustScore  float64           `json:"minTrustScore,omitempty"`
	MaxClients     int               `json:"maxClients,omitempty"`
	GeographicZone string            `json:"geographicZone,omitempty"`
	
	// Resource requirements
	MinComputeCapacity ComputeCapabilities `json:"minComputeCapacity,omitempty"`
	MaxLatency         float64             `json:"maxLatencyMs,omitempty"`
}

// ResourceRequirements defines resource requirements for the training job
type ResourceRequirements struct {
	CPU              string `json:"cpu,omitempty"`
	Memory           string `json:"memory,omitempty"`
	GPU              string `json:"gpu,omitempty"`
	Storage          string `json:"storage,omitempty"`
	NetworkBandwidth string `json:"networkBandwidth,omitempty"`
}

// TrainingJobStatus defines the observed state of a training job
type TrainingJobStatus struct {
	Phase            TrainingStatus        `json:"phase,omitempty"`
	CurrentRound     int64                `json:"currentRound,omitempty"`
	ParticipatingClients []string         `json:"participatingClients,omitempty"`
	ModelMetrics     ModelMetrics         `json:"modelMetrics,omitempty"`
	LastUpdate       metav1.Time          `json:"lastUpdate,omitempty"`
	Conditions       []metav1.Condition   `json:"conditions,omitempty"`
	Message          string               `json:"message,omitempty"`
}

// FederatedLearningManager defines the interface for managing federated learning
type FederatedLearningManager interface {
	// Client management
	RegisterClient(ctx context.Context, client *FLClient) error
	UnregisterClient(ctx context.Context, clientID string) error
	GetClient(ctx context.Context, clientID string) (*FLClient, error)
	ListClients(ctx context.Context, selector ClientSelector) ([]*FLClient, error)
	UpdateClientStatus(ctx context.Context, clientID string, status FLClientStatus) error
	
	// Model management
	CreateGlobalModel(ctx context.Context, model *GlobalModel) error
	GetGlobalModel(ctx context.Context, modelID string) (*GlobalModel, error)
	UpdateGlobalModel(ctx context.Context, model *GlobalModel) error
	ListGlobalModels(ctx context.Context, rrmTask RRMTaskType) ([]*GlobalModel, error)
	
	// Training coordination
	StartTraining(ctx context.Context, jobSpec TrainingJobSpec) (*TrainingJob, error)
	StopTraining(ctx context.Context, jobID string) error
	GetTrainingStatus(ctx context.Context, jobID string) (*TrainingJobStatus, error)
	
	// Model aggregation
	AggregateModels(ctx context.Context, modelUpdates []ModelUpdate) (*GlobalModel, error)
	ValidateModelUpdate(ctx context.Context, update ModelUpdate) error
	
	// Privacy and security
	ApplyPrivacyMechanism(ctx context.Context, model *GlobalModel, mechanism PrivacyMechanism) error
	DetectMaliciousClients(ctx context.Context, roundMetrics []ClientRoundMetrics) ([]string, error)
	
	// Monitoring and metrics
	GetTrainingMetrics(ctx context.Context, modelID string) (*ModelMetrics, error)
	GetClientMetrics(ctx context.Context, clientID string) (*ClientRoundMetrics, error)
}

// ModelUpdate represents a model update from a client
type ModelUpdate struct {
	ClientID        string            `json:"client_id"`
	ModelID         string            `json:"model_id"`
	Round           int64             `json:"round"`
	Parameters      []byte            `json:"parameters"`
	ParametersHash  string            `json:"parameters_hash"`
	DataSamplesCount int64            `json:"data_samples_count"`
	LocalMetrics    ModelMetrics      `json:"local_metrics"`
	Signature       []byte            `json:"signature"`
	Timestamp       time.Time         `json:"timestamp"`
	Metadata        map[string]string `json:"metadata"`
}

// ClientInfo represents information about a federated learning client
type ClientInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	LastSeen     time.Time `json:"last_seen"`
	Capabilities []string  `json:"capabilities"`
}

// ClientStore holds the information about registered clients.
type ClientStore struct {
	Clients map[string]*FLClient `json:"clients"`
	mutex   sync.RWMutex
}

// NewClientStore creates a new in-memory client store.
func NewClientStore() *ClientStore {
	return &ClientStore{
		Clients: make(map[string]*FLClient),
	}
}

// Store saves a client to the store.
func (cs *ClientStore) Store(ctx context.Context, client *FLClient) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	cs.Clients[client.ID] = client
	return nil
}

// Get retrieves a client from the store.
func (cs *ClientStore) Get(ctx context.Context, clientID string) (*FLClient, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	
	client, ok := cs.Clients[clientID]
	if !ok {
		return nil, fmt.Errorf("client %s not found", clientID)
	}
	return client, nil
}

// Delete removes a client from the store.
func (cs *ClientStore) Delete(ctx context.Context, clientID string) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	
	delete(cs.Clients, clientID)
	return nil
}

// List returns all clients from the store.
func (cs *ClientStore) List(ctx context.Context) ([]*FLClient, error) {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()
	
	clients := make([]*FLClient, 0, len(cs.Clients))
	for _, client := range cs.Clients {
		clients = append(clients, client)
	}
	return clients, nil
}

// ModelStore manages storage and retrieval of federated learning models
type ModelStore struct {
	Models    map[string]*GlobalModel `json:"models"`
	Versions  map[string][]string     `json:"versions"` // Stores versions for each model ID
	mutex     sync.RWMutex
}

// NewModelStore creates a new in-memory model store.
func NewModelStore() *ModelStore {
	return &ModelStore{
		Models:   make(map[string]*GlobalModel),
		Versions: make(map[string][]string),
	}
}

// CreateGlobalModel saves a new global model to the store.
func (ms *ModelStore) CreateGlobalModel(ctx context.Context, model *GlobalModel) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, exists := ms.Models[model.ID]; exists {
		return fmt.Errorf("model with ID %s already exists", model.ID)
	}

	ms.Models[model.ID] = model
	ms.Versions[model.ID] = append(ms.Versions[model.ID], model.Version)
	return nil
}

// GetGlobalModel retrieves a global model by its ID.
func (ms *ModelStore) GetGlobalModel(ctx context.Context, modelID string) (*GlobalModel, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	model, ok := ms.Models[modelID]
	if !ok {
		return nil, fmt.Errorf("model with ID %s not found", modelID)
	}
	return model, nil
}

// UpdateGlobalModel updates an existing global model in the store.
func (ms *ModelStore) UpdateGlobalModel(ctx context.Context, model *GlobalModel) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, exists := ms.Models[model.ID]; !exists {
		return fmt.Errorf("model with ID %s not found for update", model.ID)
	}

	ms.Models[model.ID] = model
	// Ensure version is added if it's new
	found := false
	for _, v := range ms.Versions[model.ID] {
		if v == model.Version {
			found = true
			break
		}
	}
	if !found {
		ms.Versions[model.ID] = append(ms.Versions[model.ID], model.Version)
	}
	return nil
}

// ListGlobalModels returns all global models, optionally filtered by RRMTask.
func (ms *ModelStore) ListGlobalModels(ctx context.Context, rrmTask RRMTaskType) ([]*GlobalModel, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	models := make([]*GlobalModel, 0, len(ms.Models))
	for _, model := range ms.Models {
		if rrmTask == "" || model.RRMTask == rrmTask {
			models = append(models, model)
		}
	}
	return models, nil
}

// MetricsStore manages storage and retrieval of training metrics
type MetricsStore struct {
	ClientMetrics map[string]map[int64]*ClientRoundMetrics `json:"client_metrics"` // clientID -> round -> metrics
	ModelMetrics  map[string]map[int64]*ModelMetrics       `json:"model_metrics"`  // modelID -> round -> metrics
	mutex         sync.RWMutex
}

// NewMetricsStore creates a new in-memory metrics store.
func NewMetricsStore() *MetricsStore {
	return &MetricsStore{
		ClientMetrics: make(map[string]map[int64]*ClientRoundMetrics),
		ModelMetrics:  make(map[string]map[int64]*ModelMetrics),
	}
}

// RecordClientMetrics records client-specific metrics for a round.
func (ms *MetricsStore) RecordClientMetrics(ctx context.Context, clientID string, round int64, metrics *ClientRoundMetrics) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, ok := ms.ClientMetrics[clientID]; !ok {
		ms.ClientMetrics[clientID] = make(map[int64]*ClientRoundMetrics)
	}
	ms.ClientMetrics[clientID][round] = metrics
	return nil
}

// RecordModelMetrics records global model metrics for a round.
func (ms *MetricsStore) RecordModelMetrics(ctx context.Context, modelID string, round int64, metrics *ModelMetrics) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, ok := ms.ModelMetrics[modelID]; !ok {
		ms.ModelMetrics[modelID] = make(map[int64]*ModelMetrics)
	}
	ms.ModelMetrics[modelID][round] = metrics
	return nil
}

// GetClientMetrics retrieves client-specific metrics.
func (ms *MetricsStore) GetClientMetrics(ctx context.Context, clientID string) (map[int64]*ClientRoundMetrics, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	metrics, ok := ms.ClientMetrics[clientID]
	if !ok {
		return nil, fmt.Errorf("metrics for client %s not found", clientID)
	}
	return metrics, nil
}

// GetModelMetrics retrieves global model metrics.
func (ms *MetricsStore) GetModelMetrics(ctx context.Context, modelID string) (map[int64]*ModelMetrics, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	metrics, ok := ms.ModelMetrics[modelID]
	if !ok {
		return nil, fmt.Errorf("metrics for model %s not found", modelID)
	}
	return metrics, nil
}

// CryptoEngine provides cryptographic operations for federated learning
type CryptoEngine struct {
	PrivateKey ed25519.PrivateKey `json:"-"`
	PublicKey  ed25519.PublicKey  `json:"public_key"`
	KeyID      string             `json:"key_id"`
	mutex      sync.RWMutex
}

// StorePrivateKey stores a private key for a client
func (ce *CryptoEngine) StorePrivateKey(clientID string, privateKey ed25519.PrivateKey) error {
	// Implementation placeholder
	return nil
}

// DeleteKeys removes cryptographic keys for a client
func (ce *CryptoEngine) DeleteKeys(clientID string) error {
	// Implementation placeholder
	return nil
}

// PrivacyEngine provides privacy-preserving operations
type PrivacyEngine interface {
	ApplyDifferentialPrivacy(data []byte, epsilon float64) ([]byte, error)
	ValidatePrivacyBudget(clientID string, requestedBudget float64) error
}

// TrustManager manages client trust scores and validation
type TrustManager interface {
	RegisterClient(ctx context.Context, client *FLClient) error
	UnregisterClient(ctx context.Context, clientID string) error
	UpdateTrustScore(ctx context.Context, clientID string, score float64) error
	GetTrustScore(ctx context.Context, clientID string) (float64, error)
}

// TrainingOrchestrator coordinates federated learning training rounds
type TrainingOrchestrator interface {
	StartRound(ctx context.Context, job *ActiveTrainingJob) error
	AggregateRound(ctx context.Context, updates []ModelUpdate) (*GlobalModel, error)
	ValidateParticipants(ctx context.Context, participants []*FLClient) error
}

// AggregationEngine performs model aggregation
type AggregationEngine interface {
	Aggregate(updates []ModelUpdate, algorithm AggregationAlgorithm) (*GlobalModel, error)
	ValidateUpdate(update ModelUpdate) error
	ApplyPrivacyMechanism(model *GlobalModel, mechanism PrivacyMechanism) error
}

// MetricsCollector collects and stores training metrics
type MetricsCollector interface {
	RecordClientRegistration(client *FLClient)
	RecordClientUnregistration(client *FLClient)
	RecordClientStatusChange(clientID string, oldStatus, newStatus FLClientStatus)
	RecordTrainingJobStart(job *TrainingJob)
	RecordRoundMetrics(metrics RoundMetrics)
}

// AlertManager handles system alerts and notifications
type AlertManager interface {
	SendAlert(alertType, message string, severity string) error
	ConfigureThresholds(thresholds map[string]float64) error
}

// ResourceManager manages computational resources
type ResourceManager interface {
	CheckAvailability(ctx context.Context, requirements ResourceRequirements) error
	AllocateResources(ctx context.Context, jobID string, requirements ResourceRequirements) error
	ReleaseResources(ctx context.Context, jobID string) error
}

// SchedulingEngine handles job scheduling
type SchedulingEngine interface {
	ScheduleJob(ctx context.Context, job *TrainingJob) error
	CancelJob(ctx context.Context, jobID string) error
	GetScheduledJobs(ctx context.Context) ([]*TrainingJob, error)
}

// TrainingMetrics represents training performance metrics
type TrainingMetrics struct {
	ClientID         string        `json:"client_id"`
	ModelID          string        `json:"model_id"`
	Round            int64         `json:"round"`
	Accuracy         float64       `json:"accuracy"`
	Loss             float64       `json:"loss"`
	TrainingDuration time.Duration `json:"training_duration"`
	DataSamples      int64         `json:"data_samples"`
	ComputeTime      time.Duration `json:"compute_time"`
	CommunicationTime time.Duration `json:"communication_time"`
	MemoryUsage      float64       `json:"memory_usage_mb"`
	CPUUsage         float64       `json:"cpu_usage_percent"`
	Timestamp        time.Time     `json:"timestamp"`
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Score   float64  `json:"score"`
	Errors  []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Details map[string]interface{} `json:"details"`
}

// StrategyConfig represents configuration for aggregation strategies
type StrategyConfig struct {
	Algorithm          AggregationAlgorithm `json:"algorithm"`
	WeightingMethod    string               `json:"weighting_method"`
	PrivacyMechanism   PrivacyMechanism     `json:"privacy_mechanism"`
	Parameters         map[string]float64   `json:"parameters"`
	QualityThresholds  QualityRequirements  `json:"quality_thresholds"`
}

// ComputeRequirement represents computational resource requirements
type ComputeRequirement struct {
	MinCPUCores      int     `json:"min_cpu_cores"`
	MinMemoryGB      float64 `json:"min_memory_gb"`
	MinGPUCount      int     `json:"min_gpu_count"`
	MinGPUMemoryGB   float64 `json:"min_gpu_memory_gb"`
	MinStorageGB     float64 `json:"min_storage_gb"`
	MinBandwidthMbps int64   `json:"min_bandwidth_mbps"`
	MaxLatencyMs     float64 `json:"max_latency_ms"`
	RequiredFeatures []string `json:"required_features"`
}

// ResourceLimit represents resource usage limits
type ResourceLimit struct {
	MaxCPUUsage     float64 `json:"max_cpu_usage_percent"`
	MaxMemoryUsage  float64 `json:"max_memory_usage_gb"`
	MaxNetworkUsage int64   `json:"max_network_usage_mbps"`
	MaxStorageUsage float64 `json:"max_storage_usage_gb"`
	PowerBudget     float64 `json:"power_budget_watts"`
	TimeLimit       time.Duration `json:"time_limit"`
}

// SecurityLevel represents security classification levels
type SecurityLevel string

const (
	SecurityLevelPublic       SecurityLevel = "public"
	SecurityLevelRestricted   SecurityLevel = "restricted"
	SecurityLevelConfidential SecurityLevel = "confidential"
	SecurityLevelSecret       SecurityLevel = "secret"
	SecurityLevelTopSecret    SecurityLevel = "top_secret"
)

// PrivacyLevel represents privacy protection levels
type PrivacyLevel string

const (
	PrivacyLevelNone     PrivacyLevel = "none"
	PrivacyLevelBasic    PrivacyLevel = "basic"
	PrivacyLevelStandard PrivacyLevel = "standard"
	PrivacyLevelHigh     PrivacyLevel = "high"
	PrivacyLevelMaximum  PrivacyLevel = "maximum"
)

// CertificationLevel represents certification compliance levels
type CertificationLevel string

const (
	CertificationLevelNone       CertificationLevel = "none"
	CertificationLevelBasic      CertificationLevel = "basic"
	CertificationLevelISO27001   CertificationLevel = "iso27001"
	CertificationLevelSOC2       CertificationLevel = "soc2"
	CertificationLevelFIPSL140   CertificationLevel = "fips140"
	CertificationLevelCommonCriteria CertificationLevel = "common_criteria"
)

// RRMTaskMetrics represents RRM task performance metrics
type RRMTaskMetrics struct {
	TaskType             RRMTaskType `json:"task_type"`
	SuccessRate          float64     `json:"success_rate"`
	AverageLatencyMs     float64     `json:"average_latency_ms"`
	ThroughputPerSecond  float64     `json:"throughput_per_second"`
	ResourceEfficiency   float64     `json:"resource_efficiency"`
	QualityScore         float64     `json:"quality_score"`
	CompletedTasks       int64       `json:"completed_tasks"`
	FailedTasks          int64       `json:"failed_tasks"`
	LastExecutionTime    time.Time   `json:"last_execution_time"`
}

// NetworkSliceMetrics represents network slice performance metrics
type NetworkSliceMetrics struct {
	SliceID             string  `json:"slice_id"`
	Throughput          float64 `json:"throughput_mbps"`
	Latency             float64 `json:"latency_ms"`
	PacketLossRate      float64 `json:"packet_loss_rate"`
	Availability        float64 `json:"availability_percent"`
	BandwidthUtilization float64 `json:"bandwidth_utilization_percent"`
	ActiveConnections   int64   `json:"active_connections"`
	QoSViolations       int64   `json:"qos_violations"`
	LastMeasurement     time.Time `json:"last_measurement"`
}

// TrendDirection represents trend directions for metrics
type TrendDirection string

const (
	TrendDirectionUp     TrendDirection = "up"
	TrendDirectionDown   TrendDirection = "down"
	TrendDirectionStable TrendDirection = "stable"
	TrendDirectionUnknown TrendDirection = "unknown"
)

// NetworkConditions represents current network conditions
type NetworkConditions struct {
	Bandwidth      float64 `json:"bandwidth_mbps"`
	Latency        float64 `json:"latency_ms"`
	PacketLoss     float64 `json:"packet_loss_rate"`
	Jitter         float64 `json:"jitter_ms"`
	Stability      float64 `json:"stability_score"`
	QualityScore   float64 `json:"quality_score"`
	Timestamp      time.Time `json:"timestamp"`
}

// ByzantineDetectionType represents types of Byzantine fault detection
type ByzantineDetectionType string

const (
	ByzantineDetectionStatistical   ByzantineDetectionType = "statistical"
	ByzantineDetectionGradient      ByzantineDetectionType = "gradient"
	ByzantineDetectionClustering    ByzantineDetectionType = "clustering"
	ByzantineDetectionConsensus     ByzantineDetectionType = "consensus"
	ByzantineDetectionAnomaly       ByzantineDetectionType = "anomaly"
)

// ThreatSeverity represents threat severity levels
type ThreatSeverity string

const (
	ThreatSeverityLow      ThreatSeverity = "low"
	ThreatSeverityMedium   ThreatSeverity = "medium"
	ThreatSeverityHigh     ThreatSeverity = "high"
	ThreatSeverityCritical ThreatSeverity = "critical"
)

// DetectionMethod represents detection method types
type DetectionMethod string

const (
	DetectionMethodSignature   DetectionMethod = "signature"
	DetectionMethodAnomaly     DetectionMethod = "anomaly"
	DetectionMethodHeuristic   DetectionMethod = "heuristic"
	DetectionMethodMachineLearning DetectionMethod = "ml"
)

// ResponseAction represents response action types
type ResponseAction string

const (
	ResponseActionAlert     ResponseAction = "alert"
	ResponseActionBlock     ResponseAction = "block"
	ResponseActionQuarantine ResponseAction = "quarantine"
	ResponseActionIsolate   ResponseAction = "isolate"
	ResponseActionTerminate ResponseAction = "terminate"
)

// PerformanceAnomaly represents detected performance anomalies
type PerformanceAnomaly struct {
	Type               string            `json:"type"`
	Severity           ThreatSeverity    `json:"severity"`
	MetricName         string            `json:"metric_name"`
	ExpectedValue      float64           `json:"expected_value"`
	ActualValue        float64           `json:"actual_value"`
	DeviationPercent   float64           `json:"deviation_percent"`
	Threshold          float64           `json:"threshold"`
	Duration           time.Duration     `json:"duration"`
	DetectionTime      time.Time         `json:"detection_time"`
	AffectedComponents []string          `json:"affected_components"`
	Details            map[string]interface{} `json:"details"`
}

// BehaviorInconsistency represents inconsistent client behavior
type BehaviorInconsistency struct {
	Type               string            `json:"type"`
	Description        string            `json:"description"`
	InconsistencyScore float64           `json:"inconsistency_score"`
	ExpectedPattern    interface{}       `json:"expected_pattern"`
	ObservedPattern    interface{}       `json:"observed_pattern"`
	DetectionTime      time.Time         `json:"detection_time"`
	Evidence           map[string]interface{} `json:"evidence"`
}

// StatisticalOutlier represents statistical outlier detection
type StatisticalOutlier struct {
	MetricName         string    `json:"metric_name"`
	Value              float64   `json:"value"`
	ZScore             float64   `json:"z_score"`
	Percentile         float64   `json:"percentile"`
	IsOutlier          bool      `json:"is_outlier"`
	OutlierThreshold   float64   `json:"outlier_threshold"`
	DetectionMethod    string    `json:"detection_method"`
	HistoricalData     []float64 `json:"historical_data"`
	DetectionTime      time.Time `json:"detection_time"`
}

// NetworkAnomaly represents network-level anomalies
type NetworkAnomaly struct {
	Type               string            `json:"type"`
	AnomalyScore       float64           `json:"anomaly_score"`
	NetworkMetrics     NetworkConditions `json:"network_metrics"`
	BaselineMetrics    NetworkConditions `json:"baseline_metrics"`
	DeviationDetails   map[string]float64 `json:"deviation_details"`
	DetectionTime      time.Time         `json:"detection_time"`
	AffectedEndpoints  []string          `json:"affected_endpoints"`
}

// E2LatencyAnomaly represents E2 interface latency anomalies
type E2LatencyAnomaly struct {
	MeasuredLatency    float64   `json:"measured_latency_ms"`
	ExpectedLatency    float64   `json:"expected_latency_ms"`
	LatencyThreshold   float64   `json:"latency_threshold_ms"`
	ExceedsThreshold   bool      `json:"exceeds_threshold"`
	E2NodeID           string    `json:"e2_node_id"`
	RICRequestID       string    `json:"ric_request_id"`
	MessageType        string    `json:"message_type"`
	DetectionTime      time.Time `json:"detection_time"`
}

// RRMTaskDeviation represents deviations in RRM task performance
type RRMTaskDeviation struct {
	TaskType           RRMTaskType       `json:"task_type"`
	ExpectedMetrics    RRMTaskMetrics    `json:"expected_metrics"`
	ActualMetrics      RRMTaskMetrics    `json:"actual_metrics"`
	DeviationScore     float64           `json:"deviation_score"`
	DeviationDetails   map[string]float64 `json:"deviation_details"`
	ImpactLevel        string            `json:"impact_level"`
	DetectionTime      time.Time         `json:"detection_time"`
}

// NetworkSliceViolation represents network slice policy violations
type NetworkSliceViolation struct {
	SliceID            string            `json:"slice_id"`
	ViolationType      string            `json:"violation_type"`
	PolicyRule         string            `json:"policy_rule"`
	ExpectedValue      interface{}       `json:"expected_value"`
	ActualValue        interface{}       `json:"actual_value"`
	Severity           ThreatSeverity    `json:"severity"`
	ViolationTime      time.Time         `json:"violation_time"`
	Duration           time.Duration     `json:"duration"`
	ImpactAssessment   string            `json:"impact_assessment"`
}

// HistoricalPattern represents historical behavior patterns
type HistoricalPattern struct {
	PatternType        string            `json:"pattern_type"`
	PatternData        interface{}       `json:"pattern_data"`
	ConfidenceScore    float64           `json:"confidence_score"`
	TimeWindow         time.Duration     `json:"time_window"`
	SampleSize         int               `json:"sample_size"`
	LastUpdated        time.Time         `json:"last_updated"`
	PatternMetadata    map[string]interface{} `json:"pattern_metadata"`
}

// PeerComparisonResult represents the result of comparing clients with peers
type PeerComparisonResult struct {
	ClientID           string            `json:"client_id"`
	PeerGroup          []string          `json:"peer_group"`
	ComparisonMetrics  map[string]float64 `json:"comparison_metrics"`
	DeviationScore     float64           `json:"deviation_score"`
	IsSuspicious       bool              `json:"is_suspicious"`
	SuspicionReasons   []string          `json:"suspicion_reasons"`
	ConfidenceLevel    float64           `json:"confidence_level"`
	ComparisonTime     time.Time         `json:"comparison_time"`
}

// ExternalValidationResult represents results from external validation
type ExternalValidationResult struct {
	ValidatorID        string            `json:"validator_id"`
	ValidationMethod   string            `json:"validation_method"`
	IsValid            bool              `json:"is_valid"`
	ConfidenceScore    float64           `json:"confidence_score"`
	ValidationDetails  map[string]interface{} `json:"validation_details"`
	Timestamp          time.Time         `json:"timestamp"`
	ErrorMessage       string            `json:"error_message,omitempty"`
}

// E2InterfaceContext represents O-RAN E2 interface context for federated learning
type E2InterfaceContext struct {
	// E2 Node identification
	E2NodeID           string            `json:"e2_node_id"`
	GlobalE2NodeID     *GlobalE2NodeID   `json:"global_e2_node_id"`
	E2NodeType         E2NodeType        `json:"e2_node_type"`
	
	// E2 Connection details
	ConnectionState    E2ConnectionState `json:"connection_state"`
	E2SetupComplete    bool              `json:"e2_setup_complete"`
	LastHeartbeat      time.Time         `json:"last_heartbeat"`
	
	// RAN Functions
	SupportedRANFunctions []RANFunction   `json:"supported_ran_functions"`
	SubscribedServices    []E2Subscription `json:"subscribed_services"`
	
	// Performance metrics
	LatencyMetrics     *E2LatencyMetrics `json:"latency_metrics"`
	MessageStats       *E2MessageStats   `json:"message_stats"`
	
	// Security context
	AuthenticationData *E2AuthContext    `json:"auth_context"`
	EncryptionEnabled  bool              `json:"encryption_enabled"`
	
	// Federated Learning specific
	FLParticipation    bool              `json:"fl_participation"`
	ModelSyncStatus    string            `json:"model_sync_status"`
	LastModelUpdate    time.Time         `json:"last_model_update"`
}

// GlobalE2NodeID represents global E2 node identifier
type GlobalE2NodeID struct {
	PLMNIdentity   string `json:"plmn_identity"`
	ENBIdentity    string `json:"enb_identity,omitempty"`
	GNBIdentity    string `json:"gnb_identity,omitempty"`
	NGENBIdentity  string `json:"ng_enb_identity,omitempty"`
}

// E2NodeType represents the type of E2 node
type E2NodeType string

const (
	E2NodeTypeENB    E2NodeType = "enb"
	E2NodeTypeGNB    E2NodeType = "gnb"
	E2NodeTypeNGENB  E2NodeType = "ng_enb"
	E2NodeTypeCUCP   E2NodeType = "cu_cp"
	E2NodeTypeCUUP   E2NodeType = "cu_up"
	E2NodeTypeDU     E2NodeType = "du"
)

// E2ConnectionState represents E2 connection state
type E2ConnectionState string

const (
	E2ConnectionStateConnected     E2ConnectionState = "connected"
	E2ConnectionStateDisconnected  E2ConnectionState = "disconnected"
	E2ConnectionStateConnecting    E2ConnectionState = "connecting"
	E2ConnectionStateReconnecting  E2ConnectionState = "reconnecting"
	E2ConnectionStateSetupFailed   E2ConnectionState = "setup_failed"
)

// RANFunction represents a RAN function supported by E2 node
type RANFunction struct {
	RANFunctionID          int64             `json:"ran_function_id"`
	RANFunctionDefinition  string            `json:"ran_function_definition"`
	RANFunctionRevision    int32             `json:"ran_function_revision"`
	RANFunctionOID         string            `json:"ran_function_oid"`
	Description            string            `json:"description"`
	Capabilities           []string          `json:"capabilities"`
}

// E2Subscription represents an E2 subscription
type E2Subscription struct {
	SubscriptionID     string            `json:"subscription_id"`
	RANFunctionID      int64             `json:"ran_function_id"`
	EventTrigger       *EventTrigger     `json:"event_trigger"`
	Actions            []E2Action        `json:"actions"`
	CreatedAt          time.Time         `json:"created_at"`
	Status             string            `json:"status"`
}

// EventTrigger represents E2 event trigger definition
type EventTrigger struct {
	TriggerType        string            `json:"trigger_type"`
	ReportingPeriod    time.Duration     `json:"reporting_period"`
	TriggerConditions  map[string]interface{} `json:"trigger_conditions"`
}

// E2Action represents E2 action definition
type E2Action struct {
	ActionID          int64             `json:"action_id"`
	ActionType        string            `json:"action_type"`
	ActionDefinition  interface{}       `json:"action_definition"`
	SubsequentActions []E2Action        `json:"subsequent_actions"`
}

// E2LatencyMetrics represents E2 interface latency metrics
type E2LatencyMetrics struct {
	SetupLatency       float64           `json:"setup_latency_ms"`
	SubscriptionLatency float64          `json:"subscription_latency_ms"`
	IndicationLatency  float64           `json:"indication_latency_ms"`
	ControlLatency     float64           `json:"control_latency_ms"`
	AverageRTT         float64           `json:"average_rtt_ms"`
	MaxRTT             float64           `json:"max_rtt_ms"`
	MinRTT             float64           `json:"min_rtt_ms"`
}

// E2MessageStats represents E2 message statistics
type E2MessageStats struct {
	TotalMessagesSent     int64             `json:"total_messages_sent"`
	TotalMessagesReceived int64             `json:"total_messages_received"`
	SuccessfulResponses   int64             `json:"successful_responses"`
	FailedResponses       int64             `json:"failed_responses"`
	TimeoutErrors         int64             `json:"timeout_errors"`
	MessageBreakdown      map[string]int64  `json:"message_breakdown"`
	LastResetTime         time.Time         `json:"last_reset_time"`
}

// E2AuthContext represents E2 authentication context
type E2AuthContext struct {
	AuthMethod         string            `json:"auth_method"`
	CertificateInfo    *CertificateInfo  `json:"certificate_info"`
	TokenValidUntil    time.Time         `json:"token_valid_until"`
	SecurityLevel      string            `json:"security_level"`
	AuthorizedActions  []string          `json:"authorized_actions"`
}

// CertificateInfo represents certificate information
type CertificateInfo struct {
	Subject            string            `json:"subject"`
	Issuer             string            `json:"issuer"`
	SerialNumber       string            `json:"serial_number"`
	ValidFrom          time.Time         `json:"valid_from"`
	ValidTo            time.Time         `json:"valid_to"`
	Fingerprint        string            `json:"fingerprint"`
}

// StatisticalTestResult represents results from statistical tests
type StatisticalTestResult struct {
	TestName           string            `json:"test_name"`
	TestType           string            `json:"test_type"`
	PValue             float64           `json:"p_value"`
	TestStatistic      float64           `json:"test_statistic"`
	CriticalValue      float64           `json:"critical_value"`
	IsSignificant      bool              `json:"is_significant"`
	ConfidenceLevel    float64           `json:"confidence_level"`
	TestResult         string            `json:"test_result"`
	AdditionalInfo     map[string]interface{} `json:"additional_info"`
}

// BehaviorPattern represents detected behavior patterns
type BehaviorPattern struct {
	PatternID          string            `json:"pattern_id"`
	PatternType        string            `json:"pattern_type"`
	Description        string            `json:"description"`
	Frequency          float64           `json:"frequency"`
	Confidence         float64           `json:"confidence"`
	FirstObserved      time.Time         `json:"first_observed"`
	LastObserved       time.Time         `json:"last_observed"`
	ObservationCount   int64             `json:"observation_count"`
	PatternData        interface{}       `json:"pattern_data"`
	RelatedPatterns    []string          `json:"related_patterns"`
}

// AnomalyDetector interface for anomaly detection
type AnomalyDetector interface {
	DetectAnomalies(data interface{}) ([]Anomaly, error)
	UpdateBaseline(data interface{}) error
	GetThreshold() float64
	SetThreshold(threshold float64)
}

// ClientBehaviorAnalyzer analyzes client behavior patterns
type ClientBehaviorAnalyzer interface {
	AnalyzeBehavior(clientID string, data interface{}) (*BehaviorAnalysis, error)
	GetBehaviorProfile(clientID string) (*BehaviorProfile, error)
	UpdateBehaviorModel(clientID string, data interface{}) error
}

// StatisticalAnomalyDetector performs statistical anomaly detection
type StatisticalAnomalyDetector interface {
	DetectStatisticalAnomalies(data []float64) ([]StatisticalAnomaly, error)
	PerformTest(testType string, data []float64) (*StatisticalTestResult, error)
	GetStatistics(data []float64) (*DescriptiveStatistics, error)
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	AnomalyID          string            `json:"anomaly_id"`
	AnomalyType        string            `json:"anomaly_type"`
	Severity           ThreatSeverity    `json:"severity"`
	Score              float64           `json:"score"`
	Description        string            `json:"description"`
	DetectedAt         time.Time         `json:"detected_at"`
	AffectedMetrics    []string          `json:"affected_metrics"`
	Evidence           interface{}       `json:"evidence"`
}

// BehaviorAnalysis represents behavior analysis results
type BehaviorAnalysis struct {
	ClientID           string            `json:"client_id"`
	AnalysisTime       time.Time         `json:"analysis_time"`
	BehaviorScore      float64           `json:"behavior_score"`
	Deviations         []BehaviorDeviation `json:"deviations"`
	Patterns           []BehaviorPattern `json:"patterns"`
	TrustScoreImpact   float64           `json:"trust_score_impact"`
}

// BehaviorProfile represents a client's behavior profile
type BehaviorProfile struct {
	ClientID           string            `json:"client_id"`
	ProfileVersion     int               `json:"profile_version"`
	CreatedAt          time.Time         `json:"created_at"`
	LastUpdated        time.Time         `json:"last_updated"`
	BaselineMetrics    map[string]float64 `json:"baseline_metrics"`
	TypicalPatterns    []BehaviorPattern `json:"typical_patterns"`
	Preferences        map[string]interface{} `json:"preferences"`
}

// BehaviorDeviation represents deviation from normal behavior
type BehaviorDeviation struct {
	MetricName         string            `json:"metric_name"`
	ExpectedValue      float64           `json:"expected_value"`
	ActualValue        float64           `json:"actual_value"`
	DeviationPercent   float64           `json:"deviation_percent"`
	Significance       float64           `json:"significance"`
	Impact             string            `json:"impact"`
}

// StatisticalAnomaly represents statistically detected anomaly
type StatisticalAnomaly struct {
	DataPoint          float64           `json:"data_point"`
	ZScore             float64           `json:"z_score"`
	PValue             float64           `json:"p_value"`
	IsOutlier          bool              `json:"is_outlier"`
	OutlierType        string            `json:"outlier_type"`
	ConfidenceInterval []float64         `json:"confidence_interval"`
}

// DescriptiveStatistics represents basic statistical measures
type DescriptiveStatistics struct {
	Count              int               `json:"count"`
	Mean               float64           `json:"mean"`
	Median             float64           `json:"median"`
	Mode               []float64         `json:"mode"`
	StandardDeviation  float64           `json:"standard_deviation"`
	Variance           float64           `json:"variance"`
	Minimum            float64           `json:"minimum"`
	Maximum            float64           `json:"maximum"`
	Range              float64           `json:"range"`
	Quartiles          []float64         `json:"quartiles"`
	InterquartileRange float64           `json:"interquartile_range"`
	Skewness           float64           `json:"skewness"`
	Kurtosis           float64           `json:"kurtosis"`
}

// AsyncTrainingScheduler manages asynchronous training schedules
type AsyncTrainingScheduler struct {
	SchedulerID    string            `json:"scheduler_id"`
	ActiveJobs     map[string]*TrainingJob `json:"active_jobs"`
	QueuedJobs     []*TrainingJob    `json:"queued_jobs"`
	MaxConcurrent  int               `json:"max_concurrent"`
	SchedulingMode string            `json:"scheduling_mode"`
}

// ResourceOrchestrator manages computational resources
type ResourceOrchestrator struct {
	OrchestratorID string                    `json:"orchestrator_id"`
	AvailableNodes map[string]*ComputeNode   `json:"available_nodes"`
	ResourcePool   *ResourcePool             `json:"resource_pool"`
	AllocationPolicy string                  `json:"allocation_policy"`
}

// ComputeNode represents a computational resource
type ComputeNode struct {
	NodeID       string            `json:"node_id"`
	NodeType     string            `json:"node_type"`
	Status       string            `json:"status"`
	Capabilities map[string]interface{} `json:"capabilities"`
	Location     string            `json:"location"`
}

// ResourcePool definition exists in resource_management.go

// AdaptiveThresholdManager manages dynamic thresholds
type AdaptiveThresholdManager struct {
	ManagerID      string                    `json:"manager_id"`
	Thresholds     map[string]*DynamicThreshold `json:"thresholds"`
	AdaptationRate float64                   `json:"adaptation_rate"`
	HistoryWindow  time.Duration             `json:"history_window"`
}

// DynamicThreshold represents an adaptive threshold
type DynamicThreshold struct {
	Name         string    `json:"name"`
	Value        float64   `json:"value"`
	MinValue     float64   `json:"min_value"`
	MaxValue     float64   `json:"max_value"`
	LastUpdated  time.Time `json:"last_updated"`
	UpdateCount  int       `json:"update_count"`
}

// ModelCompressor handles model compression operations
type ModelCompressor struct {
	CompressorID     string            `json:"compressor_id"`
	CompressionType  string            `json:"compression_type"`
	CompressionRatio float64           `json:"compression_ratio"`
	QualityThreshold float64           `json:"quality_threshold"`
}

// CommunicationOptimizer optimizes client-server communication
type CommunicationOptimizer struct {
	OptimizerID      string            `json:"optimizer_id"`
	OptimizationMode string            `json:"optimization_mode"`
	BandwidthLimit   float64           `json:"bandwidth_limit"`
	CompressionLevel int               `json:"compression_level"`
}

// GradientComputationEngine handles gradient calculations
type GradientComputationEngine struct {
	EngineID         string            `json:"engine_id"`
	ComputationMode  string            `json:"computation_mode"`
	BatchSize        int               `json:"batch_size"`
	LearningRate     float64           `json:"learning_rate"`
}

// ClientBehaviorMetrics tracks client behavior patterns
type ClientBehaviorMetrics struct {
	ClientID            string        `json:"client_id"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	SuccessRate         float64       `json:"success_rate"`
	DataQuality         float64       `json:"data_quality"`
	ModelAccuracy       float64       `json:"model_accuracy"`
	BehaviorConsistency float64       `json:"behavior_consistency"`
	AnomalyScore        float64       `json:"anomaly_score"`
}

// SecurityMetrics aggregates security-related metrics
type SecurityMetrics struct {
	TotalThreatDetections   int                    `json:"total_threat_detections"`
	CriticalThreats        int                    `json:"critical_threats"`
	ResolvedIncidents      int                    `json:"resolved_incidents"`
	AverageTrustScore      float64                `json:"average_trust_score"`
	SecurityPolicyVersion string                 `json:"security_policy_version"`
	LastSecurityScan       time.Time              `json:"last_security_scan"`
	VulnerabilityCount     map[string]int         `json:"vulnerability_count"`
}

// AsyncTrainingRequest represents an asynchronous training request
type AsyncTrainingRequest struct {
	RequestID       string                 `json:"request_id"`
	ClientID        string                 `json:"client_id"`
	ModelID         string                 `json:"model_id"`
	TrainingParams  map[string]interface{} `json:"training_params"`
	Priority        int                    `json:"priority"`
	SubmissionTime  time.Time              `json:"submission_time"`
	ExpectedDuration time.Duration         `json:"expected_duration"`
}

// Additional types for resource management
type ResourcePoolConfig struct {
	PoolID          string `json:"pool_id"`
	MaxCPU          float64 `json:"max_cpu"`
	MaxMemory       float64 `json:"max_memory"`
	MaxGPU          int     `json:"max_gpu"`
}

type ResourceCapacity struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	GPU    int     `json:"gpu"`
}

type UtilizationHistory struct {
	WindowSize time.Duration              `json:"window_size"`
	History    []ResourceUtilizationPoint `json:"history"`
}

type ResourceUtilizationPoint struct {
	Timestamp   time.Time         `json:"timestamp"`
	CPUUsage    float64          `json:"cpu_usage"`
	MemoryUsage float64          `json:"memory_usage"`
	GPUUsage    float64          `json:"gpu_usage"`
}

type AutoScaler struct {
	ScalerID     string  `json:"scaler_id"`
	MinReplicas  int     `json:"min_replicas"`
	MaxReplicas  int     `json:"max_replicas"`
	TargetCPU    float64 `json:"target_cpu"`
}

type ScalePredictor struct {
	PredictorID string `json:"predictor_id"`
	Algorithm   string `json:"algorithm"`
}

type CostOptimizer struct {
	OptimizerID string `json:"optimizer_id"`
	Strategy    string `json:"strategy"`
}

type QuotaManager struct {
	ManagerID string                    `json:"manager_id"`
	Quotas    map[string]*ResourceQuota `json:"quotas"`
}

type ResourceQuota struct {
	QuotaID     string  `json:"quota_id"`
	MaxCPU      float64 `json:"max_cpu"`
	MaxMemory   float64 `json:"max_memory"`
	MaxGPU      int     `json:"max_gpu"`
}

type FairnessController struct {
	ControllerID string `json:"controller_id"`
	Algorithm    string `json:"algorithm"`
}

// gRPC message types for xApp communication

// SubmitModelUpdateRequest represents a gRPC request for submitting a model update.
type SubmitModelUpdateRequest struct {
	ClientID        string       `json:"client_id"`
	ModelID         string       `json:"model_id"`
	Round           int64        `json:"round"`
	Parameters      []byte       `json:"parameters"`
	ParametersHash  string       `json:"parameters_hash"`
	DataSamplesCount int64       `json:"data_samples_count"`
	LocalMetrics    *ModelMetrics `json:"local_metrics"`
	Signature       []byte       `json:"signature"`
	Timestamp       *time.Time   `json:"timestamp"`
	Metadata        map[string]string `json:"metadata"`
}

// SubmitModelUpdateResponse represents a gRPC response for a model update submission.
type SubmitModelUpdateResponse struct {
	Status    string     `json:"status"`
	Round     int64      `json:"round"`
	Timestamp *time.Time `json:"timestamp"`
	Message   string     `json:"message"`
}

// AllocationPolicy type and constants defined in config.go

