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
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// DynamicResourceManager provides adaptive resource allocation and auto-scaling
type DynamicResourceManager struct {
	logger                  *slog.Logger
	kubeClient             kubernetes.Interface
	metricsClient          metrics.Interface
	resourcePredictorMLModel MLModel
	qosSchedulingEngine    QoSScheduler
	autoScalingConfig      AutoScalingConfig
	
	// Resource monitoring and tracking
	nodeResourceCache      map[string]*NodeResourceStatus
	clientResourceCache    map[string]*ClientResourceStatus
	jobSchedulingQueue     *PriorityJobQueue
	activeJobs            map[string]*ScheduledJob
	
	// Auto-scaling state
	lastScaleAction       time.Time
	scaleActionHistory    []ScaleActionRecord
	currentUtilization    *ClusterUtilization
	
	// QoS and network monitoring
	networkLatencyMatrix  map[string]map[string]time.Duration
	clientCapacityScores  map[string]float64
	
	mutex                 sync.RWMutex
	stopCh               chan struct{}
	updateInterval       time.Duration
}

// AutoScalingConfig defines auto-scaling behavior parameters
type AutoScalingConfig struct {
	// CPU and Memory thresholds
	CPUTargetUtilization    float64       `json:"cpu_target_utilization"`
	MemoryTargetUtilization float64       `json:"memory_target_utilization"`
	
	// Scaling timing constraints
	ScaleUpThreshold       time.Duration `json:"scale_up_threshold"`
	ScaleDownThreshold     time.Duration `json:"scale_down_threshold"`
	ScaleUpCooldown       time.Duration `json:"scale_up_cooldown"`
	ScaleDownCooldown     time.Duration `json:"scale_down_cooldown"`
	
	// Job limits and constraints
	MaxConcurrentJobsLimit int           `json:"max_concurrent_jobs_limit"`
	MinConcurrentJobs     int           `json:"min_concurrent_jobs"`
	
	// Prediction and adaptation
	PredictionWindow      time.Duration `json:"prediction_window"`
	AdaptationAggression  float64       `json:"adaptation_aggression"`
	
	// QoS-specific thresholds
	LatencySensitiveThreshold float64    `json:"latency_sensitive_threshold"`
	ThroughputOptimizedThreshold float64 `json:"throughput_optimized_threshold"`
}

// NodeResourceStatus tracks real-time node resource availability
type NodeResourceStatus struct {
	NodeName            string                 `json:"node_name"`
	Timestamp           time.Time              `json:"timestamp"`
	CPUCapacity         float64                `json:"cpu_capacity"`
	CPUAllocatable      float64                `json:"cpu_allocatable"`
	CPUUsage           float64                `json:"cpu_usage"`
	MemoryCapacity      int64                  `json:"memory_capacity"`
	MemoryAllocatable   int64                  `json:"memory_allocatable"`
	MemoryUsage        int64                  `json:"memory_usage"`
	GPUCapacity        int                    `json:"gpu_capacity"`
	GPUUsage           int                    `json:"gpu_usage"`
	NetworkBandwidth   float64                `json:"network_bandwidth"`
	NetworkUtilization float64                `json:"network_utilization"`
	NodeLabels         map[string]string      `json:"node_labels"`
	NodeConditions     []NodeConditionStatus  `json:"node_conditions"`
	AvailableScore     float64                `json:"available_score"`
}

// ClientResourceStatus tracks client-side resource capabilities
type ClientResourceStatus struct {
	ClientID           string            `json:"client_id"`
	LastUpdate         time.Time         `json:"last_update"`
	ComputeCapacity    ComputeCapability `json:"compute_capacity"`
	NetworkCondition   NetworkCondition  `json:"network_condition"`
	EnergyConstraints  EnergyConstraints `json:"energy_constraints"`
	AvailabilityWindow AvailabilityWindow `json:"availability_window"`
	QoSPreferences     QoSPreferences    `json:"qos_preferences"`
	HistoricalPerformance PerformanceHistory `json:"historical_performance"`
}

// ComputeCapability represents client computational resources
type ComputeCapability struct {
	CPUCores           int     `json:"cpu_cores"`
	CPUFrequency       float64 `json:"cpu_frequency"`
	MemoryGB          float64 `json:"memory_gb"`
	HasGPU            bool    `json:"has_gpu"`
	GPUMemoryGB       float64 `json:"gpu_memory_gb"`
	StorageGB         float64 `json:"storage_gb"`
	ComputeScore      float64 `json:"compute_score"`
}

// NetworkCondition represents network connectivity quality
type NetworkCondition struct {
	Latency           time.Duration `json:"latency"`
	Bandwidth         float64       `json:"bandwidth_mbps"`
	PacketLoss        float64       `json:"packet_loss_percent"`
	Jitter            time.Duration `json:"jitter"`
	ConnectionType    string        `json:"connection_type"`
	NetworkQuality    float64       `json:"network_quality_score"`
}

// EnergyConstraints represents client energy limitations
type EnergyConstraints struct {
	BatteryLevel      float64       `json:"battery_level_percent"`
	PowerConsumption  float64       `json:"power_consumption_watts"`
	IsPluggedIn       bool          `json:"is_plugged_in"`
	EnergyBudget      float64       `json:"energy_budget_wh"`
	MaxPowerDraw      float64       `json:"max_power_draw_watts"`
}

// QoSPreferences defines client Quality of Service preferences
type QoSPreferences struct {
	MaxLatency        time.Duration `json:"max_latency"`
	MinThroughput     float64       `json:"min_throughput"`
	ReliabilityLevel  float64       `json:"reliability_level"`
	CostSensitivity   float64       `json:"cost_sensitivity"`
	EnergyEfficiency  float64       `json:"energy_efficiency_priority"`
}

// ClusterUtilization represents overall cluster resource usage
type ClusterUtilization struct {
	Timestamp              time.Time `json:"timestamp"`
	OverallCPUUtilization  float64   `json:"overall_cpu_utilization"`
	OverallMemoryUtilization float64 `json:"overall_memory_utilization"`
	OverallGPUUtilization  float64   `json:"overall_gpu_utilization"`
	NetworkUtilization     float64   `json:"network_utilization"`
	ActiveJobCount         int       `json:"active_job_count"`
	QueuedJobCount         int       `json:"queued_job_count"`
	AvailableNodeCount     int       `json:"available_node_count"`
	PredictedUtilization   *PredictedUtilization `json:"predicted_utilization"`
}

// PredictedUtilization contains ML-based resource usage predictions
type PredictedUtilization struct {
	NextPeriodCPU     float64   `json:"next_period_cpu"`
	NextPeriodMemory  float64   `json:"next_period_memory"`
	PeakTime          time.Time `json:"peak_time"`
	PredictionWindow  time.Duration `json:"prediction_window"`
	Confidence        float64   `json:"confidence"`
}

// ScaleActionRecord tracks scaling decisions for analysis
type ScaleActionRecord struct {
	Timestamp      time.Time        `json:"timestamp"`
	Action         ScaleActionType  `json:"action"`
	FromValue      int             `json:"from_value"`
	ToValue        int             `json:"to_value"`
	Trigger        string          `json:"trigger"`
	CPUUtilization float64         `json:"cpu_utilization"`
	MemoryUtilization float64      `json:"memory_utilization"`
	ActiveJobs     int             `json:"active_jobs"`
	QueuedJobs     int             `json:"queued_jobs"`
}

// ScaleActionType defines possible scaling actions
type ScaleActionType string

const (
	ScaleActionUp        ScaleActionType = "scale_up"
	ScaleActionDown      ScaleActionType = "scale_down"
	ScaleActionHold      ScaleActionType = "hold"
	ScaleActionEmergency ScaleActionType = "emergency_scale"
)

// NewDynamicResourceManager creates a new dynamic resource manager
func NewDynamicResourceManager(
	logger *slog.Logger,
	kubeClient kubernetes.Interface,
	metricsClient metrics.Interface,
	config AutoScalingConfig,
) (*DynamicResourceManager, error) {
	
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if kubeClient == nil {
		return nil, fmt.Errorf("kubernetes client cannot be nil")
	}
	if metricsClient == nil {
		return nil, fmt.Errorf("metrics client cannot be nil")
	}
	
	// Validate configuration
	if err := validateAutoScalingConfig(config); err != nil {
		return nil, fmt.Errorf("invalid auto-scaling configuration: %w", err)
	}
	
	drm := &DynamicResourceManager{
		logger:                 logger,
		kubeClient:            kubeClient,
		metricsClient:         metricsClient,
		autoScalingConfig:     config,
		nodeResourceCache:     make(map[string]*NodeResourceStatus),
		clientResourceCache:   make(map[string]*ClientResourceStatus),
		activeJobs:           make(map[string]*ScheduledJob),
		networkLatencyMatrix: make(map[string]map[string]time.Duration),
		clientCapacityScores: make(map[string]float64),
		scaleActionHistory:   make([]ScaleActionRecord, 0),
		stopCh:              make(chan struct{}),
		updateInterval:      30 * time.Second,
	}
	
	// Initialize components
	if err := drm.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}
	
	logger.Info("Dynamic Resource Manager initialized successfully",
		slog.Int("max_concurrent_jobs", config.MaxConcurrentJobsLimit),
		slog.Float64("cpu_target", config.CPUTargetUtilization),
		slog.Float64("memory_target", config.MemoryTargetUtilization))
	
	return drm, nil
}

// Start begins the dynamic resource management monitoring and control loops
func (drm *DynamicResourceManager) Start(ctx context.Context) error {
	drm.logger.Info("Starting Dynamic Resource Manager")
	
	// Start monitoring goroutines
	go drm.resourceMonitoringLoop(ctx)
	go drm.autoScalingLoop(ctx)
	go drm.clientSelectionUpdateLoop(ctx)
	go drm.qosSchedulingLoop(ctx)
	
	// Start ML model if available
	if drm.resourcePredictorMLModel != nil {
		go drm.predictionLoop(ctx)
	}
	
	<-ctx.Done()
	drm.logger.Info("Dynamic Resource Manager stopped")
	return nil
}

// Stop gracefully shuts down the dynamic resource manager
func (drm *DynamicResourceManager) Stop() {
	drm.logger.Info("Stopping Dynamic Resource Manager")
	close(drm.stopCh)
}

// GetClusterUtilization returns current cluster resource utilization
func (drm *DynamicResourceManager) GetClusterUtilization(ctx context.Context) (*ClusterUtilization, error) {
	drm.mutex.RLock()
	defer drm.mutex.RUnlock()
	
	if drm.currentUtilization == nil {
		// Perform immediate calculation if not cached
		utilization, err := drm.calculateClusterUtilization(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate cluster utilization: %w", err)
		}
		return utilization, nil
	}
	
	return drm.currentUtilization, nil
}

// validateAutoScalingConfig validates the auto-scaling configuration
func validateAutoScalingConfig(config AutoScalingConfig) error {
	if config.CPUTargetUtilization <= 0 || config.CPUTargetUtilization > 1 {
		return fmt.Errorf("CPU target utilization must be between 0 and 1")
	}
	if config.MemoryTargetUtilization <= 0 || config.MemoryTargetUtilization > 1 {
		return fmt.Errorf("memory target utilization must be between 0 and 1")
	}
	if config.MaxConcurrentJobsLimit <= 0 {
		return fmt.Errorf("max concurrent jobs limit must be positive")
	}
	if config.ScaleUpThreshold <= 0 {
		return fmt.Errorf("scale up threshold must be positive")
	}
	if config.ScaleDownThreshold <= 0 {
		return fmt.Errorf("scale down threshold must be positive")
	}
	return nil
}

// initializeComponents initializes all sub-components
func (drm *DynamicResourceManager) initializeComponents() error {
	// Initialize job queue with priority support
	drm.jobSchedulingQueue = NewPriorityJobQueue()
	
	// Initialize QoS scheduling engine
	qosEngine, err := NewQoSScheduler(drm.logger, drm.autoScalingConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize QoS scheduler: %w", err)
	}
	drm.qosSchedulingEngine = qosEngine
	
	// Initialize ML model for resource prediction (optional)
	mlModel, err := NewResourcePredictorMLModel(drm.logger)
	if err != nil {
		drm.logger.Warn("Failed to initialize ML model, using heuristic prediction", 
			slog.String("error", err.Error()))
	} else {
		drm.resourcePredictorMLModel = mlModel
	}
	
	return nil
}