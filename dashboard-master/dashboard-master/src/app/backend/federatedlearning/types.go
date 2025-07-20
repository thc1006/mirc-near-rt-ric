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

// SecurityIncident represents a security-related event during training
type SecurityIncident struct {
	Type        string            `json:"type" bson:"type"`
	Severity    string            `json:"severity" bson:"severity"`
	ClientID    string            `json:"client_id" bson:"client_id"`
	Description string            `json:"description" bson:"description"`
	Timestamp   time.Time         `json:"timestamp" bson:"timestamp"`
	Metadata    map[string]string `json:"metadata" bson:"metadata"`
}

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