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
	"time"
	"crypto/x509"
)

// AsyncCoordinatorConfig configuration for the asynchronous FL coordinator
type AsyncCoordinatorConfig struct {
	// Base coordinator configuration
	CoordinatorConfig       CoordinatorConfig `yaml:"coordinator"`

	// Asynchronous specific settings
	MinClientThreshold      int               `yaml:"min_client_threshold"`
	MaxClientThreshold      int               `yaml:"max_client_threshold"`
	StragglerTimeout        time.Duration     `yaml:"straggler_timeout"`
	AggregationWindow       time.Duration     `yaml:"aggregation_window"`
	PartialAggregationEnabled bool            `yaml:"partial_aggregation_enabled"`

	// Strategy configurations
	AggregationStrategy     AggregationStrategyConfig `yaml:"aggregation_strategy"`
	ClientSelectionPolicy  ClientSelectionConfig     `yaml:"client_selection"`
	ByzantineFTConfig       ByzantineFTConfig         `yaml:"byzantine_ft"`

	// Resource management
	JobPoolConfig           JobPoolConfig             `yaml:"job_pool"`
	MonitoringConfig        MonitoringConfig          `yaml:"monitoring"`
	SecurityConfig          SecurityConfig            `yaml:"security"`
}

// AggregationStrategyConfig configuration for aggregation strategies
type AggregationStrategyConfig struct {
	// Strategy selection
	DefaultStrategy         AggregationAlgorithm `yaml:"default_strategy"`
	EnableAdaptiveStrategy  bool                 `yaml:"enable_adaptive_strategy"`

	// FedAvg configuration
	FedAvgConfig           *FedAvgConfig         `yaml:"fedavg,omitempty"`

	// FedAsync configuration
	FedAsyncConfig         *FedAsyncConfig       `yaml:"fedasync,omitempty"`

	// FedProx configuration
	FedProxConfig          *FedProxConfig        `yaml:"fedprox,omitempty"`

	// Secure aggregation configuration
	SecureAggConfig        *SecureAggConfig      `yaml:"secure_agg,omitempty"`

	// Byzantine FT configuration
	ByzantineFTStrategyConfig *ByzantineFTStrategyConfig `yaml:"byzantine_ft,omitempty"`
}

// ClientSelectionConfig configuration for client selection policies
type ClientSelectionConfig struct {
	// Strategy settings
	DefaultStrategy        ClientSelectionStrategy `yaml:"default_strategy"`
	DefaultOptimalSize     int                     `yaml:"default_optimal_size"`
	MinClients            int                     `yaml:"min_clients"`
	MaxClients            int                     `yaml:"max_clients"`

	// Adaptive configuration
	AdaptiveWeightsConfig    *AdaptiveWeightsConfig    `yaml:"adaptive_weights,omitempty"`
	SelectionCriteriaConfig  *SelectionCriteriaConfig  `yaml:"selection_criteria,omitempty"`

	// Performance-based selection
	PerformanceWeights     map[string]float64    `yaml:"performance_weights"`
	QualityThresholds      map[string]float64    `yaml:"quality_thresholds"`

	// Diversity-based selection
	DiversityWeights       map[string]float64    `yaml:"diversity_weights"`
	DiversityTargets       map[string]float64    `yaml:"diversity_targets"`

	// O-RAN specific configuration
	RRMTaskWeights         map[RRMTaskType]float64 `yaml:"rrm_task_weights"`
	E2LatencyThresholds    map[string]time.Duration `yaml:"e2_latency_thresholds"`
	NetworkSliceWeights    map[string]float64      `yaml:"network_slice_weights"`
}

// ByzantineFTConfig configuration for Byzantine fault tolerance
type ByzantineFTConfig struct {
	// Detection settings
	EnableStatisticalDetection   bool                    `yaml:"enable_statistical_detection"`
	EnableBehaviorAnalysis       bool                    `yaml:"enable_behavior_analysis"`
	EnablePerformanceAnalysis    bool                    `yaml:"enable_performance_analysis"`
	EnableORANSpecificDetection  bool                    `yaml:"enable_oran_specific_detection"`

	// Detection configurations
	AnomalyDetectionConfig       *AnomalyDetectionConfig       `yaml:"anomaly_detection,omitempty"`
	BehaviorAnalysisConfig       *BehaviorAnalysisConfig       `yaml:"behavior_analysis,omitempty"`
	StatisticalDetectionConfig   *StatisticalDetectionConfig   `yaml:"statistical_detection,omitempty"`

	// Trust and reputation
	TrustScoreWeights           TrustScoreWeights             `yaml:"trust_score_weights"`
	ReputationConfig            *ReputationConfig             `yaml:"reputation,omitempty"`

	// Defense mechanisms
	RobustAggregationStrategies []RobustAggregationStrategy   `yaml:"robust_aggregation_strategies"`
	FilteringThresholds         map[string]float64            `yaml:"filtering_thresholds"`

	// Security policies
	SecurityPolicyConfig        *SecurityPolicyConfig         `yaml:"security_policy,omitempty"`
	IncidentResponseConfig      *IncidentResponseConfig       `yaml:"incident_response,omitempty"`
}

// JobPoolConfig configuration for async job pool management
type JobPoolConfig struct {
	// Pool sizing
	MaxConcurrentJobs       int                     `yaml:"max_concurrent_jobs"`
	WorkerPoolSize          int                     `yaml:"worker_pool_size"`
	JobQueueSize            int                     `yaml:"job_queue_size"`

	// Resource management
	ResourceManagerConfig   *ResourceManagerConfig  `yaml:"resource_manager,omitempty"`
	LoadBalancerConfig      *LoadBalancerConfig     `yaml:"load_balancer,omitempty"`

	// Performance optimization
	EnableJobPrioritization bool                    `yaml:"enable_job_prioritization"`
	EnableResourceOptimization bool                 `yaml:"enable_resource_optimization"`
	EnablePerformanceMonitoring bool                `yaml:"enable_performance_monitoring"`

	// Scheduling
	SchedulingPolicy        SchedulingPolicy        `yaml:"scheduling_policy"`
	PriorityLevels          []JobPriority          `yaml:"priority_levels"`
}

// MonitoringConfig configuration for performance monitoring
type MonitoringConfig struct {
	// Monitoring settings
	EnablePrometheusMetrics   bool                  `yaml:"enable_prometheus_metrics"`
	EnableOpenTelemetry      bool                  `yaml:"enable_opentelemetry"`
	EnableRealTimeDashboard  bool                  `yaml:"enable_realtime_dashboard"`

	// Metrics collection
	MetricsCollectionInterval time.Duration         `yaml:"metrics_collection_interval"`
	MetricsRetentionPeriod   time.Duration         `yaml:"metrics_retention_period"`
	MetricsAggregationWindow time.Duration         `yaml:"metrics_aggregation_window"`

	// Performance tracking
	LatencyTrackerConfig     *LatencyTrackerConfig  `yaml:"latency_tracker,omitempty"`
	ThroughputTrackerConfig  *ThroughputTrackerConfig `yaml:"throughput_tracker,omitempty"`
	ResourceTrackerConfig    *ResourceTrackerConfig `yaml:"resource_tracker,omitempty"`

	// Alerting
	AlertingConfig           *AlertingConfig        `yaml:"alerting,omitempty"`
	ThresholdConfig          *ThresholdConfig       `yaml:"thresholds,omitempty"`

	// Dashboard configuration
	DashboardConfig          *DashboardConfig       `yaml:"dashboard,omitempty"`
	VisualizationConfig      *VisualizationConfig   `yaml:"visualization,omitempty"`

	// O-RAN specific monitoring
	E2MonitoringConfig       *E2MonitoringConfig    `yaml:"e2_monitoring,omitempty"`
	RRMMonitoringConfig      *RRMMonitoringConfig   `yaml:"rrm_monitoring,omitempty"`
	NetworkSliceMonitoringConfig *NetworkSliceMonitoringConfig `yaml:"network_slice_monitoring,omitempty"`

	// History and storage
	HistoryConfig            *HistoryConfig         `yaml:"history,omitempty"`
	StorageConfig            *StorageConfig         `yaml:"storage,omitempty"`
}

// SecurityConfig configuration for security and privacy
type SecurityConfig struct {
	// Security settings
	EnableHomomorphicEncryption bool                    `yaml:"enable_homomorphic_encryption"`
	EnableDifferentialPrivacy   bool                    `yaml:"enable_differential_privacy"`
	EnableSecureAggregation     bool                    `yaml:"enable_secure_aggregation"`

	// Authentication and authorization
	AuthenticationConfig        *AuthenticationConfig   `yaml:"authentication,omitempty"`
	AuthorizationConfig         *AuthorizationConfig    `yaml:"authorization,omitempty"`
	CertificateConfig           *CertificateConfig      `yaml:"certificate,omitempty"`

	// Encryption configuration
	EncryptionConfig            *EncryptionConfig       `yaml:"encryption,omitempty"`
	HomomorphicConfig           *HomomorphicConfig      `yaml:"homomorphic,omitempty"`

	// Privacy configuration
	PrivacyConfig               *PrivacyConfig          `yaml:"privacy,omitempty"`
	BudgetConfig                *PrivacyBudgetConfig    `yaml:"budget,omitempty"`

	// Secure aggregation
	SecureAggregationConfig     *SecureAggregationConfig `yaml:"secure_aggregation,omitempty"`

	// Security monitoring
	SecurityMonitorConfig       *SecurityMonitorConfig  `yaml:"security_monitor,omitempty"`
	ComplianceConfig            *ComplianceConfig       `yaml:"compliance,omitempty"`

	// O-RAN specific security
	E2SecurityConfig            *E2SecurityConfig       `yaml:"e2_security,omitempty"`
	NetworkSliceSecurityConfig  *NetworkSliceSecurityConfig `yaml:"network_slice_security,omitempty"`

	// Session management
	SessionTimeout              time.Duration           `yaml:"session_timeout"`
	TokenExpirationTime         time.Duration           `yaml:"token_expiration_time"`
}

// ResourceManagerConfig configuration for resource management
type ResourceManagerConfig struct {
	// Resource pools
	ResourcePoolConfig          *ResourcePoolConfig     `yaml:"resource_pool,omitempty"`
	ComputePoolConfig           *ComputePoolConfig      `yaml:"compute_pool,omitempty"`
	NetworkPoolConfig           *NetworkPoolConfig      `yaml:"network_pool,omitempty"`
	StoragePoolConfig           *StoragePoolConfig      `yaml:"storage_pool,omitempty"`

	// Allocation policies
	AllocationPolicy            AllocationPolicy        `yaml:"allocation_policy"`
	QuotaConfig                 *QuotaConfig           `yaml:"quota,omitempty"`
	FairnessConfig              *FairnessConfig        `yaml:"fairness,omitempty"`

	// Optimization
	EnableResourceOptimization  bool                    `yaml:"enable_resource_optimization"`
	OptimizationInterval        time.Duration           `yaml:"optimization_interval"`
	OptimizationStrategy        OptimizationStrategy    `yaml:"optimization_strategy"`

	// Monitoring and metrics
	UtilizationThresholds       map[string]float64      `yaml:"utilization_thresholds"`
	AllocationTimeouts          map[string]time.Duration `yaml:"allocation_timeouts"`

	// O-RAN specific resources
	E2ProcessingUnits           int64                   `yaml:"e2_processing_units"`
	A1PolicyEngines             int64                   `yaml:"a1_policy_engines"`
	O1ManagementUnits           int64                   `yaml:"o1_management_units"`
	RRMComputeUnits             int64                   `yaml:"rrm_compute_units"`
}

// TrustScoreWeights configuration for trust score computation
type TrustScoreWeights struct {
	Current                 float64 `yaml:"current"`
	Behavior                float64 `yaml:"behavior"`
	Reputation              float64 `yaml:"reputation"`
	Historical              float64 `yaml:"historical"`
	E2Compliance            float64 `yaml:"e2_compliance"`
	RRMTask                 float64 `yaml:"rrm_task"`
}

// AdaptiveWeightsConfig configuration for adaptive weight computation
type AdaptiveWeightsConfig struct {
	// Weight parameters
	Accuracy                float64 `yaml:"accuracy"`
	Latency                 float64 `yaml:"latency"`
	Reliability             float64 `yaml:"reliability"`
	Trust                   float64 `yaml:"trust"`
	ResourceEfficiency      float64 `yaml:"resource_efficiency"`

	// O-RAN specific weights
	RRMTaskCompatibility    float64 `yaml:"rrm_task_compatibility"`
	E2Latency               float64 `yaml:"e2_latency"`
	NetworkSliceAffinity    float64 `yaml:"network_slice_affinity"`
	DataDiversity           float64 `yaml:"data_diversity"`
	SecurityCompliance      float64 `yaml:"security_compliance"`
	GeographicDiversity     float64 `yaml:"geographic_diversity"`
	LoadBalancingPenalty    float64 `yaml:"load_balancing_penalty"`
	TrendBonus              float64 `yaml:"trend_bonus"`

	// Adaptation parameters
	LearningRate            float64 `yaml:"learning_rate"`
	AdaptationInterval      time.Duration `yaml:"adaptation_interval"`
	PerformanceWindow       time.Duration `yaml:"performance_window"`
}

// LatencyTrackerConfig configuration for latency tracking
type LatencyTrackerConfig struct {
	// Measurement settings
	MeasurementInterval     time.Duration `yaml:"measurement_interval"`
	BufferSize              int           `yaml:"buffer_size"`
	RetentionPeriod         time.Duration `yaml:"retention_period"`

	// Statistical analysis
	EnableStatisticalAnalysis bool        `yaml:"enable_statistical_analysis"`
	EnableTrendAnalysis      bool         `yaml:"enable_trend_analysis"`
	EnablePredictiveModeling bool         `yaml:"enable_predictive_modeling"`

	// Thresholds
	LatencyThresholds       *LatencyThresholds `yaml:"latency_thresholds,omitempty"`
	SLAConfig               *SLAConfig         `yaml:"sla,omitempty"`

	// O-RAN specific latency tracking
	E2LatencyThresholds     map[RRMTaskType]time.Duration `yaml:"e2_latency_thresholds"`
	RRMTaskLatencyLimits    map[RRMTaskType]time.Duration `yaml:"rrm_task_latency_limits"`
	NetworkSliceLatencyLimits map[string]time.Duration   `yaml:"network_slice_latency_limits"`
}

// DashboardConfig configuration for real-time dashboard
type DashboardConfig struct {
	// Dashboard settings
	EnableWebUI             bool          `yaml:"enable_web_ui"`
	Port                    int           `yaml:"port"`
	UpdateInterval          time.Duration `yaml:"update_interval"`

	// Data management
	DataCacheSize           int           `yaml:"data_cache_size"`
	CompressionEnabled      bool          `yaml:"compression_enabled"`
	RateLimitEnabled        bool          `yaml:"rate_limit_enabled"`

	// WebSocket configuration
	WebSocketConfig         *WebSocketConfig `yaml:"websocket,omitempty"`
	BroadcasterConfig       *BroadcasterConfig `yaml:"broadcaster,omitempty"`

	// Dashboard views
	OverviewDashboardConfig *OverviewDashboardConfig `yaml:"overview,omitempty"`
	JobDashboardConfig      *JobDashboardConfig      `yaml:"job,omitempty"`
	ClientDashboardConfig   *ClientDashboardConfig   `yaml:"client,omitempty"`
	SecurityDashboardConfig *SecurityDashboardConfig `yaml:"security,omitempty"`

	// O-RAN specific dashboards
	RRMDashboardConfig      *RRMDashboardConfig      `yaml:"rrm,omitempty"`
	E2DashboardConfig       *E2DashboardConfig       `yaml:"e2,omitempty"`
	NetworkSliceDashboardConfig *NetworkSliceDashboardConfig `yaml:"network_slice,omitempty"`
}

// PrivacyConfig configuration for differential privacy
type PrivacyConfig struct {
	// Privacy parameters
	DefaultEpsilon          float64       `yaml:"default_epsilon"`
	DefaultDelta            float64       `yaml:"default_delta"`
	DefaultSensitivity      float64       `yaml:"default_sensitivity"`

	// Noise mechanisms
	LaplaceMechanismConfig  *LaplaceMechanismConfig  `yaml:"laplace,omitempty"`
	GaussianMechanismConfig *GaussianMechanismConfig `yaml:"gaussian,omitempty"`
	ExponentialMechanismConfig *ExponentialMechanismConfig `yaml:"exponential,omitempty"`

	// Advanced privacy techniques
	LocalDPConfig           *LocalDPConfig           `yaml:"local_dp,omitempty"`
	RenyiDPConfig           *RenyiDPConfig           `yaml:"renyi_dp,omitempty"`
	AdaptiveDPConfig        *AdaptiveDPConfig        `yaml:"adaptive_dp,omitempty"`

	// Budget management
	TotalBudget             float64                  `yaml:"total_budget"`
	BudgetAllocationStrategy BudgetAllocationStrategy `yaml:"budget_allocation_strategy"`
	BudgetRefreshInterval   time.Duration            `yaml:"budget_refresh_interval"`

	// O-RAN specific privacy
	RRMTaskPrivacyBudgets   map[RRMTaskType]float64  `yaml:"rrm_task_privacy_budgets"`
	NetworkSlicePrivacyBudgets map[string]float64    `yaml:"network_slice_privacy_budgets"`
	E2InterfacePrivacyBudget float64                 `yaml:"e2_interface_privacy_budget"`
}

// HomomorphicConfig configuration for homomorphic encryption
type HomomorphicConfig struct {
	// Encryption scheme
	EncryptionScheme        HomomorphicScheme `yaml:"encryption_scheme"`
	KeySize                 int               `yaml:"key_size"`
	SecurityLevel           int               `yaml:"security_level"`

	// Performance optimization
	BatchProcessingEnabled  bool              `yaml:"batch_processing_enabled"`
	ParallelComputationEnabled bool           `yaml:"parallel_computation_enabled"`
	KeyCacheSize            int               `yaml:"key_cache_size"`

	// Key management
	KeyRotationEnabled      bool              `yaml:"key_rotation_enabled"`
	KeyRotationInterval     time.Duration     `yaml:"key_rotation_interval"`
	KeyGenerationMethod     KeyGenerationMethod `yaml:"key_generation_method"`

	// Scheme-specific configurations
	PaillierConfig          *PaillierConfig   `yaml:"paillier,omitempty"`
	BFVConfig               *BFVConfig        `yaml:"bfv,omitempty"`
	CKKSConfig              *CKKSConfig       `yaml:"ckks,omitempty"`
}

// Default configurations
func DefaultAsyncCoordinatorConfig() AsyncCoordinatorConfig {
	return AsyncCoordinatorConfig{
		MinClientThreshold:        3,
		MaxClientThreshold:        50,
		StragglerTimeout:          30 * time.Second,
		AggregationWindow:         10 * time.Second,
		PartialAggregationEnabled: true,

		AggregationStrategy: AggregationStrategyConfig{
			DefaultStrategy:        AggregationFedAvg,
			EnableAdaptiveStrategy: true,
			FedAvgConfig: &FedAvgConfig{
				AdaptiveWeights:     true,
				QualityThreshold:    0.7,
				ParallelAggregation: true,
				CompressionEnabled:  true,
				QuantizationLevel:   8,
			},
		},

		ClientSelectionPolicy: ClientSelectionConfig{
			DefaultStrategy:    ClientSelectionAdaptive,
			DefaultOptimalSize: 10,
			MinClients:         3,
			MaxClients:         50,
			PerformanceWeights: map[string]float64{
				"accuracy":    0.3,
				"reliability": 0.25,
				"latency":     0.2,
				"trust":       0.25,
			},
		},

		ByzantineFTConfig: ByzantineFTConfig{
			EnableStatisticalDetection:  true,
			EnableBehaviorAnalysis:      true,
			EnablePerformanceAnalysis:   true,
			EnableORANSpecificDetection: true,
			TrustScoreWeights: TrustScoreWeights{
				Current:      0.3,
				Behavior:     0.25,
				Reputation:   0.2,
				Historical:   0.15,
				E2Compliance: 0.05,
				RRMTask:      0.05,
			},
		},

		JobPoolConfig: JobPoolConfig{
			MaxConcurrentJobs:           10,
			WorkerPoolSize:              20,
			JobQueueSize:                100,
			EnableJobPrioritization:     true,
			EnableResourceOptimization:  true,
			EnablePerformanceMonitoring: true,
			SchedulingPolicy:            SchedulingPolicyFIFO,
		},

		MonitoringConfig: MonitoringConfig{
			EnablePrometheusMetrics:   true,
			EnableOpenTelemetry:       true,
			EnableRealTimeDashboard:   true,
			MetricsCollectionInterval: 10 * time.Second,
			MetricsRetentionPeriod:    24 * time.Hour,
			MetricsAggregationWindow:  1 * time.Minute,
		},

		SecurityConfig: SecurityConfig{
			EnableHomomorphicEncryption: false,
			EnableDifferentialPrivacy:   true,
			EnableSecureAggregation:     false,
			SessionTimeout:              1 * time.Hour,
			TokenExpirationTime:         30 * time.Minute,
			PrivacyConfig: &PrivacyConfig{
				DefaultEpsilon:           1.0,
				DefaultDelta:             1e-5,
				DefaultSensitivity:       1.0,
				TotalBudget:              10.0,
				BudgetAllocationStrategy: BudgetAllocationUniform,
				BudgetRefreshInterval:    24 * time.Hour,
			},
		},
	}
}

// Validation methods
func (c *AsyncCoordinatorConfig) Validate() error {
	if c.MinClientThreshold < 1 {
		return fmt.Errorf("min_client_threshold must be at least 1")
	}
	if c.MaxClientThreshold < c.MinClientThreshold {
		return fmt.Errorf("max_client_threshold must be >= min_client_threshold")
	}
	if c.StragglerTimeout <= 0 {
		return fmt.Errorf("straggler_timeout must be positive")
	}
	if c.AggregationWindow <= 0 {
		return fmt.Errorf("aggregation_window must be positive")
	}
	return nil
}

// Enum types for configuration
type AggregationAlgorithm string

const (
	AggregationFedAvg    AggregationAlgorithm = "fedavg"
	AggregationFedAsync  AggregationAlgorithm = "fedasync"
	AggregationFedProx   AggregationAlgorithm = "fedprox"
	AggregationSecure    AggregationAlgorithm = "secure"
	AggregationByzantine AggregationAlgorithm = "byzantine"
)

type ClientSelectionStrategy string

const (
	ClientSelectionRandom      ClientSelectionStrategy = "random"
	ClientSelectionPerformance ClientSelectionStrategy = "performance"
	ClientSelectionDiversity   ClientSelectionStrategy = "diversity"
	ClientSelectionAdaptive    ClientSelectionStrategy = "adaptive"
	ClientSelectionHybrid      ClientSelectionStrategy = "hybrid"
	ClientSelectionORANOptimal ClientSelectionStrategy = "oran_optimal"
)

type SchedulingPolicy string

const (
	SchedulingPolicyFIFO     SchedulingPolicy = "fifo"
	SchedulingPolicyPriority SchedulingPolicy = "priority"
	SchedulingPolicyRoundRobin SchedulingPolicy = "round_robin"
	SchedulingPolicyAdaptive SchedulingPolicy = "adaptive"
)

type AllocationPolicy string

const (
	AllocationPolicyFirstFit  AllocationPolicy = "first_fit"
	AllocationPolicyBestFit   AllocationPolicy = "best_fit"
	AllocationPolicyWorstFit  AllocationPolicy = "worst_fit"
	AllocationPolicyOptimal   AllocationPolicy = "optimal"
)

type OptimizationStrategy string

const (
	OptimizationGreedy     OptimizationStrategy = "greedy"
	OptimizationGenetic    OptimizationStrategy = "genetic"
	OptimizationSimulated  OptimizationStrategy = "simulated_annealing"
	OptimizationReinforcement OptimizationStrategy = "reinforcement_learning"
)

type BudgetAllocationStrategy string

const (
	BudgetAllocationUniform   BudgetAllocationStrategy = "uniform"
	BudgetAllocationAdaptive  BudgetAllocationStrategy = "adaptive"
	BudgetAllocationOptimal   BudgetAllocationStrategy = "optimal"
)

type HomomorphicScheme string

const (
	HomomorphicPaillier HomomorphicScheme = "paillier"
	HomomorphicBFV      HomomorphicScheme = "bfv"
	HomomorphicCKKS     HomomorphicScheme = "ckks"
)

type KeyGenerationMethod string

const (
	KeyGenerationSecure   KeyGenerationMethod = "secure"
	KeyGenerationFast     KeyGenerationMethod = "fast"
	KeyGenerationBalanced KeyGenerationMethod = "balanced"
)