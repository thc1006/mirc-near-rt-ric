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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"

	"github.com/kubernetes/dashboard/src/app/backend/errors"
)

// AsyncFLCoordinator implements asynchronous federated learning with advanced aggregation strategies
type AsyncFLCoordinator struct {
	// Base coordinator capabilities
	*FLCoordinator

	// Asynchronous aggregation components
	aggregationStrategy    AggregationStrategy
	clientSelectionPolicy ClientSelectionPolicy
	byzantineFaultTolerance ByzantineFTEngine
	partialAggregationEnabled bool
	
	// Dynamic thresholds and timeouts
	minClientThreshold       int32
	maxClientThreshold       int32
	stragglerTimeout        time.Duration
	asyncAggregationWindow  time.Duration
	
	// Advanced scheduling and resource management
	scheduler               AsyncTrainingScheduler
	resourceOrchestrator   ResourceOrchestrator
	adaptiveThreshold      AdaptiveThresholdManager
	
	// Performance optimization
	modelCompressor        ModelCompressor
	communicationOptimizer CommunicationOptimizer
	gradientComputation    GradientComputationEngine
	
	// Concurrent job management with enhanced isolation
	jobPool                *AsyncJobPool
	jobScheduler          *JobScheduler
	priorityQueue         *PriorityQueue
	
	// Real-time monitoring and metrics
	performanceMonitor    *PerformanceMonitor
	networkMonitor        *NetworkMonitor
	latencyTracker        *LatencyTracker
	
	// Security and privacy enhancements
	homomorphicEngine     *HomomorphicEncryption
	differentialPrivacy   *DifferentialPrivacyEngine
	secureAggregation     *SecureAggregationProtocol
	
	// State management
	activeAggregations    sync.Map // map[string]*AsyncAggregationJob
	clientStates          sync.Map // map[string]*ClientState
	roundMetrics          sync.Map // map[string]*RoundMetrics
	
	// Control channels
	aggregationTrigger    chan AggregationTrigger
	priorityUpdates       chan PriorityUpdate
	resourceUpdates       chan ResourceUpdate
	
	// Atomic counters for performance
	totalJobsScheduled    int64
	successfulAggregations int64
	byzantineDetections   int64
	adaptiveAdjustments   int64
}

// AggregationStrategy defines the federated learning aggregation approach
type AggregationStrategy interface {
	// Core aggregation methods
	Aggregate(ctx context.Context, models []*ModelUpdate, weights []float64) (*GlobalModel, error)
	ComputeWeights(clients []*FLClient, metrics []*TrainingMetrics) ([]float64, error)
	ValidateUpdate(update *ModelUpdate, baseline *GlobalModel) (*ValidationResult, error)
	
	// Asynchronous specific methods
	CanAggregate(availableUpdates int, totalClients int, elapsedTime time.Duration) bool
	GetOptimalBatchSize(clientLatencies []time.Duration, resourceConstraints *ResourceConstraints) int
	AdaptStrategy(performanceMetrics *AggregationMetrics) error
	
	// Strategy information
	GetStrategy() AggregationAlgorithm
	GetConfiguration() *StrategyConfig
}

// ClientSelectionPolicy defines advanced client selection strategies
type ClientSelectionPolicy interface {
	// Client selection methods
	SelectClients(ctx context.Context, availableClients []*FLClient, requirements *SelectionRequirements) ([]*FLClient, error)
	UpdateClientPriority(clientID string, metrics *ClientMetrics) error
	GetOptimalClientCount(task RRMTaskType, networkConditions *NetworkConditions) int
	
	// Dynamic adaptation
	AdaptSelection(ctx context.Context, roundMetrics *RoundMetrics) error
	GetSelectionStrategy() ClientSelectionStrategy
}

// ByzantineFTEngine provides Byzantine fault tolerance capabilities
type ByzantineFTEngine interface {
	// Detection methods
	DetectByzantineClients(ctx context.Context, updates []*ModelUpdate, baseline *GlobalModel) ([]*ByzantineDetection, error)
	ValidateModelUpdate(update *ModelUpdate, clientHistory *ClientTrainingHistory) (*SecurityAssessment, error)
	ComputeTrustScore(clientID string, behaviorMetrics *ClientBehaviorMetrics) (float64, error)
	
	// Defense mechanisms
	FilterMaliciousUpdates(updates []*ModelUpdate, detections []*ByzantineDetection) ([]*ModelUpdate, error)
	ApplyRobustAggregation(updates []*ModelUpdate, strategy RobustAggregationStrategy) (*GlobalModel, error)
	
	// Adaptive security
	UpdateSecurityPolicy(ctx context.Context, threatLevel ThreatLevel) error
	GetSecurityMetrics() *SecurityMetrics
}

// AsyncAggregationJob represents an asynchronous aggregation operation
type AsyncAggregationJob struct {
	ID                  string
	JobID               string
	Strategy            AggregationStrategy
	TargetModelID       string
	
	// Timing and coordination
	StartTime           time.Time
	DeadlineTime        time.Time
	LastAggregation     time.Time
	AggregationWindow   time.Duration
	
	// Client management
	ExpectedClients     map[string]*FLClient
	ReceivedUpdates     map[string]*ModelUpdate
	PendingClients      map[string]time.Time
	StragglerClients    map[string]*StragglerInfo
	
	// Aggregation control
	MinUpdatesRequired  int
	MaxUpdatesAllowed   int
	CurrentUpdateCount  int32
	PartialAggregation  bool
	
	// Quality and performance
	QualityThreshold    float64
	ConvergenceMetrics  *ConvergenceMetrics
	PerformanceTargets  *PerformanceTargets
	
	// Security and privacy
	PrivacyBudget       float64
	SecurityLevel       SecurityLevel
	TrustRequirements   *TrustRequirements
	
	// State synchronization
	mutex               sync.RWMutex
	updateChan          chan *ModelUpdate
	aggregationChan     chan *AggregationResult
	errorChan           chan error
	completionChan      chan struct{}
	
	// Resource tracking
	ComputeResources    *ResourceAllocation
	NetworkBandwidth    *BandwidthAllocation
	StorageRequirements *StorageRequirements
}

// AsyncJobPool manages concurrent asynchronous training jobs
type AsyncJobPool struct {
	maxConcurrentJobs   int32
	currentJobs         int32
	jobQueue            chan *AsyncTrainingRequest
	workerPool          chan struct{}
	
	// Resource management
	resourceManager     *AsyncResourceManager
	jobScheduler        *JobScheduler
	priorityManager     *PriorityManager
	
	// Performance optimization
	loadBalancer        *LoadBalancer
	resourceOptimizer   *ResourceOptimizer
	performanceMonitor  *JobPerformanceMonitor
	
	// State tracking
	activeJobs          sync.Map
	completedJobs       sync.Map
	failedJobs          sync.Map
	
	// Control
	ctx                 context.Context
	cancel              context.CancelFunc
	wg                  sync.WaitGroup
}

// NewAsyncFLCoordinator creates an optimized asynchronous federated learning coordinator
func NewAsyncFLCoordinator(
	logger *slog.Logger,
	kubeClient kubernetes.Interface,
	config AsyncCoordinatorConfig,
) (*AsyncFLCoordinator, error) {
	
	// Create base FL coordinator
	baseCoordinator, err := NewFLCoordinator(logger, kubeClient, config.CoordinatorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create base coordinator: %w", err)
	}
	
	coordinator := &AsyncFLCoordinator{
		FLCoordinator:            baseCoordinator,
		minClientThreshold:       int32(config.MinClientThreshold),
		maxClientThreshold:       int32(config.MaxClientThreshold),
		stragglerTimeout:        config.StragglerTimeout,
		asyncAggregationWindow:  config.AggregationWindow,
		partialAggregationEnabled: config.PartialAggregationEnabled,
		
		// Initialize channels
		aggregationTrigger:      make(chan AggregationTrigger, 100),
		priorityUpdates:        make(chan PriorityUpdate, 50),
		resourceUpdates:        make(chan ResourceUpdate, 50),
	}
	
	// Initialize aggregation strategy
	if err := coordinator.initializeAggregationStrategy(config.AggregationStrategy); err != nil {
		return nil, fmt.Errorf("failed to initialize aggregation strategy: %w", err)
	}
	
	// Initialize client selection policy
	if err := coordinator.initializeClientSelection(config.ClientSelectionPolicy); err != nil {
		return nil, fmt.Errorf("failed to initialize client selection: %w", err)
	}
	
	// Initialize Byzantine fault tolerance
	if err := coordinator.initializeByzantineFT(config.ByzantineFTConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize Byzantine FT: %w", err)
	}
	
	// Initialize async job pool
	if err := coordinator.initializeJobPool(config.JobPoolConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize job pool: %w", err)
	}
	
	// Initialize performance monitoring
	if err := coordinator.initializePerformanceMonitoring(config.MonitoringConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize performance monitoring: %w", err)
	}
	
	// Initialize security components
	if err := coordinator.initializeSecurityComponents(config.SecurityConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize security components: %w", err)
	}
	
	// Start background services
	coordinator.startAsyncBackgroundServices()
	
	return coordinator, nil
}

// StartAsyncTraining initiates an asynchronous federated learning training job
func (c *AsyncFLCoordinator) StartAsyncTraining(ctx context.Context, jobSpec AsyncTrainingJobSpec) (*TrainingJob, error) {
	c.logger.Info("Starting asynchronous FL training job",
		slog.String("model_id", jobSpec.ModelID),
		slog.String("rrm_task", string(jobSpec.RRMTask)),
		slog.String("aggregation_strategy", string(jobSpec.AggregationStrategy)))
	
	// Validate job specification with enhanced checks
	if err := c.validateAsyncTrainingJobSpec(jobSpec); err != nil {
		return nil, errors.NewErrorBuilder(errors.ErrInvalidInput).
			WithDetail("validation_error", err.Error()).
			WithCaller().
			Build()
	}
	
	// Check resource availability with predictive analysis
	resourceRequirements := c.calculateResourceRequirements(jobSpec)
	if err := c.resourceOrchestrator.ReserveResources(ctx, resourceRequirements); err != nil {
		return nil, errors.NewErrorBuilder(errors.ErrServiceUnavailable).
			WithDetail("resource_error", err.Error()).
			WithCaller().
			Build()
	}
	
	// Create enhanced training job
	job := &TrainingJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "federatedlearning.o-ran.org/v1alpha1",
			Kind:       "AsyncTrainingJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("async-fl-job-%s", uuid.New().String()[:8]),
			Namespace: "federated-learning",
			Labels: map[string]string{
				"app":                  "federated-learning",
				"component":            "async-coordinator",
				"rrm-task":            string(jobSpec.RRMTask),
				"model-id":            jobSpec.ModelID,
				"aggregation-strategy": string(jobSpec.AggregationStrategy),
				"priority":            string(jobSpec.Priority),
			},
			Annotations: map[string]string{
				"fl.o-ran.org/async-enabled":     "true",
				"fl.o-ran.org/partial-agg":      fmt.Sprintf("%t", jobSpec.PartialAggregationEnabled),
				"fl.o-ran.org/min-clients":      fmt.Sprintf("%d", jobSpec.MinClientThreshold),
				"fl.o-ran.org/straggler-timeout": jobSpec.StragglerTimeout.String(),
			},
		},
		Spec: TrainingJobSpec{
			ModelID:           jobSpec.ModelID,
			RRMTask:          jobSpec.RRMTask,
			TrainingConfig:   jobSpec.TrainingConfig,
			ClientSelector:   jobSpec.ClientSelector,
			Resources:        resourceRequirements,
			Privacy:          jobSpec.PrivacyConfig,
			Security:         jobSpec.SecurityConfig,
		},
		Status: TrainingJobStatus{
			Phase:      TrainingStatusInitializing,
			LastUpdate: metav1.Now(),
		},
	}
	
	// Select clients using advanced selection policy
	selectedClients, err := c.clientSelectionPolicy.SelectClients(ctx, nil, &SelectionRequirements{
		RRMTask:           jobSpec.RRMTask,
		MinClients:        int(c.minClientThreshold),
		MaxClients:        int(c.maxClientThreshold),
		TrustThreshold:    jobSpec.TrustThreshold,
		LatencyRequirement: jobSpec.LatencyRequirement,
		ComputeRequirement: jobSpec.ComputeRequirement,
		NetworkSlices:     jobSpec.NetworkSlices,
	})
	if err != nil {
		c.resourceOrchestrator.ReleaseResources(ctx, resourceRequirements)
		return nil, fmt.Errorf("failed to select clients: %w", err)
	}
	
	if len(selectedClients) < int(c.minClientThreshold) {
		c.resourceOrchestrator.ReleaseResources(ctx, resourceRequirements)
		return nil, errors.NewErrorBuilder(errors.ErrServiceUnavailable).
			WithDetail("reason", "insufficient participants for async training").
			WithDetail("required", c.minClientThreshold).
			WithDetail("available", len(selectedClients)).
			WithCaller().
			Build()
	}
	
	// Create asynchronous aggregation job
	asyncJob := &AsyncAggregationJob{
		ID:                 uuid.New().String(),
		JobID:              job.Name,
		Strategy:           c.aggregationStrategy,
		TargetModelID:      jobSpec.ModelID,
		StartTime:          time.Now(),
		DeadlineTime:       time.Now().Add(jobSpec.TrainingConfig.MaxDuration),
		AggregationWindow:  c.asyncAggregationWindow,
		ExpectedClients:    make(map[string]*FLClient),
		ReceivedUpdates:    make(map[string]*ModelUpdate),
		PendingClients:     make(map[string]time.Time),
		StragglerClients:   make(map[string]*StragglerInfo),
		MinUpdatesRequired: int(c.minClientThreshold),
		MaxUpdatesAllowed:  int(c.maxClientThreshold),
		PartialAggregation: c.partialAggregationEnabled,
		QualityThreshold:   jobSpec.QualityThreshold,
		PrivacyBudget:      jobSpec.PrivacyConfig.Budget,
		SecurityLevel:      jobSpec.SecurityConfig.Level,
		updateChan:         make(chan *ModelUpdate, len(selectedClients)*2),
		aggregationChan:    make(chan *AggregationResult, 10),
		errorChan:          make(chan error, 100),
		completionChan:     make(chan struct{}),
		ComputeResources:   resourceRequirements.ComputeAllocation,
		NetworkBandwidth:   resourceRequirements.NetworkAllocation,
		StorageRequirements: resourceRequirements.StorageAllocation,
	}
	
	// Set up client expectations
	for _, client := range selectedClients {
		asyncJob.ExpectedClients[client.ID] = client
		asyncJob.PendingClients[client.ID] = time.Now()
	}
	
	// Initialize convergence tracking
	asyncJob.ConvergenceMetrics = &ConvergenceMetrics{
		TargetAccuracy:    jobSpec.TargetAccuracy,
		ConvergenceTolerance: jobSpec.ConvergenceTolerance,
		MaxRounds:         jobSpec.TrainingConfig.MaxRounds,
		EarlyStoppingEnabled: jobSpec.EarlyStoppingEnabled,
	}
	
	// Initialize performance targets
	asyncJob.PerformanceTargets = &PerformanceTargets{
		MaxLatency:        jobSpec.LatencyRequirement,
		MinThroughput:     jobSpec.ThroughputRequirement,
		MaxResourceUsage:  jobSpec.ResourceLimits,
		QualityThreshold:  jobSpec.QualityThreshold,
	}
	
	// Store async job
	c.activeAggregations.Store(asyncJob.ID, asyncJob)
	
	// Submit to job pool for asynchronous execution
	if err := c.jobPool.SubmitJob(ctx, &AsyncTrainingRequest{
		Job:        job,
		AsyncJob:   asyncJob,
		Clients:    selectedClients,
		StartTime:  time.Now(),
		Priority:   jobSpec.Priority,
	}); err != nil {
		c.activeAggregations.Delete(asyncJob.ID)
		c.resourceOrchestrator.ReleaseResources(ctx, resourceRequirements)
		return nil, fmt.Errorf("failed to submit async job: %w", err)
	}
	
	// Update job status
	job.Status.Phase = TrainingStatusRunning
	job.Status.ParticipatingClients = make([]string, 0, len(selectedClients))
	for _, client := range selectedClients {
		job.Status.ParticipatingClients = append(job.Status.ParticipatingClients, client.ID)
	}
	job.Status.LastUpdate = metav1.Now()
	
	// Start asynchronous orchestration
	c.wg.Add(1)
	go c.orchestrateAsyncTraining(ctx, asyncJob)
	
	// Update metrics
	atomic.AddInt64(&c.totalJobsScheduled, 1)
	
	// Emit training start event
	c.performanceMonitor.RecordAsyncJobStart(job, asyncJob)
	
	c.logger.Info("Successfully started async FL training job",
		slog.String("job_id", job.Name),
		slog.String("async_job_id", asyncJob.ID),
		slog.Int("participants", len(selectedClients)),
		slog.String("strategy", string(jobSpec.AggregationStrategy)),
		slog.Duration("aggregation_window", c.asyncAggregationWindow))
	
	return job, nil
}

// orchestrateAsyncTraining manages the asynchronous training process
func (c *AsyncFLCoordinator) orchestrateAsyncTraining(ctx context.Context, asyncJob *AsyncAggregationJob) {
	defer c.wg.Done()
	defer close(asyncJob.completionChan)
	
	jobCtx, cancel := context.WithDeadline(ctx, asyncJob.DeadlineTime)
	defer cancel()
	
	c.logger.Info("Starting async training orchestration",
		slog.String("async_job_id", asyncJob.ID),
		slog.Int("expected_clients", len(asyncJob.ExpectedClients)))
	
	// Start aggregation timer
	aggregationTimer := time.NewTicker(asyncJob.AggregationWindow)
	defer aggregationTimer.Stop()
	
	// Start straggler detection
	stragglerTimer := time.NewTicker(c.stragglerTimeout)
	defer stragglerTimer.Stop()
	
	// Performance monitoring
	performanceTicker := time.NewTicker(30 * time.Second)
	defer performanceTicker.Stop()
	
	roundNumber := int64(0)
	lastAggregationTime := time.Now()
	
	for {
		select {
		case <-jobCtx.Done():
			c.handleAsyncJobCompletion(asyncJob, "deadline_exceeded")
			return
			
		case update := <-asyncJob.updateChan:
			if err := c.processAsyncModelUpdate(jobCtx, asyncJob, update); err != nil {
				c.logger.Error("Failed to process async model update",
					slog.String("async_job_id", asyncJob.ID),
					slog.String("client_id", update.ClientID),
					slog.String("error", err.Error()))
				continue
			}
			
			// Check if we can trigger aggregation
			if c.canTriggerAggregation(asyncJob) {
				aggregationTimer.Reset(0) // Trigger immediate aggregation
			}
			
		case <-aggregationTimer.C:
			if c.shouldAggregate(asyncJob) {
				roundNumber++
				c.logger.Info("Triggering async aggregation",
					slog.String("async_job_id", asyncJob.ID),
					slog.Int64("round", roundNumber),
					slog.Int("available_updates", len(asyncJob.ReceivedUpdates)))
				
				if err := c.performAsyncAggregation(jobCtx, asyncJob, roundNumber); err != nil {
					c.logger.Error("Async aggregation failed",
						slog.String("async_job_id", asyncJob.ID),
						slog.Int64("round", roundNumber),
						slog.String("error", err.Error()))
				} else {
					lastAggregationTime = time.Now()
					atomic.AddInt64(&c.successfulAggregations, 1)
				}
			}
			
		case <-stragglerTimer.C:
			c.handleStragglers(jobCtx, asyncJob)
			
		case <-performanceTicker.C:
			c.monitorAsyncJobPerformance(asyncJob)
			
		case result := <-asyncJob.aggregationChan:
			if err := c.handleAggregationResult(jobCtx, asyncJob, result); err != nil {
				c.logger.Error("Failed to handle aggregation result",
					slog.String("async_job_id", asyncJob.ID),
					slog.String("error", err.Error()))
			}
			
		case err := <-asyncJob.errorChan:
			c.logger.Error("Async job error",
				slog.String("async_job_id", asyncJob.ID),
				slog.String("error", err.Error()))
			
			// Check if this is a critical error
			if c.isCriticalError(err) {
				c.handleAsyncJobCompletion(asyncJob, "critical_error")
				return
			}
		}
		
		// Check convergence and completion conditions
		if c.checkAsyncJobCompletion(asyncJob, roundNumber) {
			c.handleAsyncJobCompletion(asyncJob, "converged")
			return
		}
	}
}

// Additional helper methods would be implemented here...
// Including methods for:
// - processAsyncModelUpdate
// - canTriggerAggregation
// - shouldAggregate
// - performAsyncAggregation
// - handleStragglers
// - monitorAsyncJobPerformance
// - handleAggregationResult
// - isCriticalError
// - checkAsyncJobCompletion
// - handleAsyncJobCompletion

// SubmitModelUpdate handles asynchronous model updates from clients
func (c *AsyncFLCoordinator) SubmitModelUpdate(ctx context.Context, update *ModelUpdate) error {
	// Find the relevant async job
	var targetJob *AsyncAggregationJob
	c.activeAggregations.Range(func(key, value interface{}) bool {
		job := value.(*AsyncAggregationJob)
		job.mutex.RLock()
		if _, exists := job.ExpectedClients[update.ClientID]; exists {
			targetJob = job
			job.mutex.RUnlock()
			return false // Stop iteration
		}
		job.mutex.RUnlock()
		return true // Continue iteration
	})
	
	if targetJob == nil {
		return errors.NewNotFound("active_job", update.ClientID)
	}
	
	// Validate update security and privacy
	if err := c.validateAsyncModelUpdate(ctx, targetJob, update); err != nil {
		return errors.NewErrorBuilder(errors.ErrInvalidInput).
			WithDetail("validation_error", err.Error()).
			WithCaller().
			Build()
	}
	
	// Submit to job's update channel
	select {
	case targetJob.updateChan <- update:
		c.logger.Debug("Async model update submitted",
			slog.String("async_job_id", targetJob.ID),
			slog.String("client_id", update.ClientID),
			slog.Int("update_size_bytes", len(update.ModelData)))
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return errors.NewServiceUnavailable("update_channel_full")
	}
}

// GetAsyncJobStatus returns the status of an asynchronous training job
func (c *AsyncFLCoordinator) GetAsyncJobStatus(ctx context.Context, jobID string) (*AsyncJobStatus, error) {
	var targetJob *AsyncAggregationJob
	c.activeAggregations.Range(func(key, value interface{}) bool {
		job := value.(*AsyncAggregationJob)
		if job.JobID == jobID {
			targetJob = job
			return false
		}
		return true
	})
	
	if targetJob == nil {
		return nil, errors.NewNotFound("async_job", jobID)
	}
	
	targetJob.mutex.RLock()
	defer targetJob.mutex.RUnlock()
	
	status := &AsyncJobStatus{
		ID:                    targetJob.ID,
		JobID:                targetJob.JobID,
		Strategy:             targetJob.Strategy.GetStrategy(),
		StartTime:            targetJob.StartTime,
		LastAggregation:      targetJob.LastAggregation,
		ExpectedClients:      len(targetJob.ExpectedClients),
		ReceivedUpdates:      len(targetJob.ReceivedUpdates),
		PendingClients:       len(targetJob.PendingClients),
		StragglerClients:     len(targetJob.StragglerClients),
		CurrentUpdateCount:   int(atomic.LoadInt32(&targetJob.CurrentUpdateCount)),
		PartialAggregation:   targetJob.PartialAggregation,
		QualityThreshold:     targetJob.QualityThreshold,
		ConvergenceMetrics:   targetJob.ConvergenceMetrics,
		PerformanceTargets:   targetJob.PerformanceTargets,
		ComputeUtilization:   c.resourceOrchestrator.GetUtilization(targetJob.ComputeResources),
		NetworkUtilization:   c.resourceOrchestrator.GetNetworkUtilization(targetJob.NetworkBandwidth),
	}
	
	return status, nil
}

// Shutdown gracefully stops the async FL coordinator
func (c *AsyncFLCoordinator) Shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down async FL coordinator")
	
	// Stop job pool
	if c.jobPool != nil {
		if err := c.jobPool.Shutdown(ctx); err != nil {
			c.logger.Error("Failed to shutdown job pool", slog.String("error", err.Error()))
		}
	}
	
	// Complete all active aggregations
	c.activeAggregations.Range(func(key, value interface{}) bool {
		job := value.(*AsyncAggregationJob)
		close(job.updateChan)
		close(job.aggregationChan)
		close(job.errorChan)
		return true
	})
	
	// Release resources
	if c.resourceOrchestrator != nil {
		if err := c.resourceOrchestrator.Shutdown(ctx); err != nil {
			c.logger.Error("Failed to shutdown resource orchestrator", slog.String("error", err.Error()))
		}
	}
	
	// Shutdown base coordinator
	if err := c.FLCoordinator.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown base coordinator: %w", err)
	}
	
	c.logger.Info("Async FL coordinator shutdown completed")
	return nil
}