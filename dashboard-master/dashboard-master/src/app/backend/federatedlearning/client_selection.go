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
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/kubernetes/dashboard/src/app/backend/errors"
)

// AdaptiveClientSelectionPolicy implements intelligent client selection for O-RAN environments
type AdaptiveClientSelectionPolicy struct {
	logger               *slog.Logger
	config               *ClientSelectionConfig
	
	// Selection strategies
	selectionStrategy    ClientSelectionStrategy
	diversityOptimizer   *DiversityOptimizer
	performancePredictor *ClientPerformancePredictor
	
	// O-RAN specific components
	rrmTaskAnalyzer      *RRMTaskAnalyzer
	networkSliceManager  *NetworkSliceManager
	e2LatencyPredictor   *E2LatencyPredictor
	
	// Adaptive mechanisms
	performanceTracker   *ClientPerformanceTracker
	adaptationEngine     *SelectionAdaptationEngine
	feedbackCollector    *FeedbackCollector
	
	// Resource awareness
	resourceMonitor      *ResourceMonitor
	capacityEstimator    *CapacityEstimator
	loadBalancer         *ClientLoadBalancer
	
	// Quality assurance
	qualityAssurance     *QualityAssuranceEngine
	convergencePredictor *ConvergencePredictor
	diversityMeasurer    *DiversityMeasurer
	
	// State management
	clientMetrics        sync.Map // map[string]*ClientMetrics
	selectionHistory     sync.Map // map[string]*SelectionHistory
	performanceHistory   sync.Map // map[string]*PerformanceHistory
	
	// Dynamic parameters
	currentOptimalSize   int32
	adaptiveWeights      *AdaptiveWeights
	selectionCriteria    *DynamicSelectionCriteria
	
	mutex                sync.RWMutex
}

// ClientSelectionStrategy defines the client selection approach
type ClientSelectionStrategy string

const (
	ClientSelectionRandom        ClientSelectionStrategy = "random"
	ClientSelectionPerformance   ClientSelectionStrategy = "performance_based"
	ClientSelectionDiversity     ClientSelectionStrategy = "diversity_based"
	ClientSelectionAdaptive      ClientSelectionStrategy = "adaptive"
	ClientSelectionHybrid        ClientSelectionStrategy = "hybrid"
	ClientSelectionORANOptimal   ClientSelectionStrategy = "oran_optimal"
)

// SelectionRequirements defines requirements for client selection
type SelectionRequirements struct {
	RRMTask              RRMTaskType
	MinClients           int
	MaxClients           int
	TrustThreshold       float64
	LatencyRequirement   time.Duration
	ComputeRequirement   *ComputeRequirement
	NetworkSlices        []string
	
	// Quality requirements
	MinAccuracy          float64
	MaxResourceUsage     *ResourceLimit
	DataDiversityTarget  float64
	
	// O-RAN specific requirements
	E2LatencyThreshold   time.Duration
	A1PolicyCompliance   bool
	O1ManagementSupport  bool
	
	// Temporal requirements
	SelectionDeadline    time.Time
	TrainingDuration     time.Duration
	
	// Security requirements
	SecurityLevel        SecurityLevel
	PrivacyRequirement   PrivacyLevel
	CertificationLevel   CertificationLevel
}

// ClientMetrics represents comprehensive client performance metrics
type ClientMetrics struct {
	ClientID             string
	LastUpdated          time.Time
	
	// Performance metrics
	AverageAccuracy      float64
	TrainingLatency      time.Duration
	CommunicationLatency time.Duration
	ThroughputMBps       float64
	
	// Reliability metrics
	AvailabilityScore    float64
	ReliabilityScore     float64
	UpTimePercentage     float64
	FailureRate          float64
	
	// Resource metrics
	CPUUtilization       float64
	MemoryUtilization    float64
	NetworkUtilization   float64
	StorageUtilization   float64
	
	// Quality metrics
	DataQualityScore     float64
	ModelQualityScore    float64
	DiversityContribution float64
	ConsistencyScore     float64
	
	// O-RAN specific metrics
	E2LatencyAverage     time.Duration
	E2LatencyVariance    float64
	RRMTaskPerformance   map[RRMTaskType]*RRMTaskMetrics
	NetworkSliceMetrics  map[string]*NetworkSliceMetrics
	
	// Trust and security metrics
	TrustScore           float64
	SecurityScore        float64
	ComplianceScore      float64
	ReputationScore      float64
	
	// Historical trends
	PerformanceTrend     TrendDirection
	ReliabilityTrend     TrendDirection
	QualityTrend         TrendDirection
	
	// Selection history
	SelectionCount       int64
	SuccessfulRounds     int64
	FailedRounds         int64
	LastSelectedTime     time.Time
}

// NewAdaptiveClientSelectionPolicy creates a new adaptive client selection policy
func NewAdaptiveClientSelectionPolicy(logger *slog.Logger, config *ClientSelectionConfig) (*AdaptiveClientSelectionPolicy, error) {
	policy := &AdaptiveClientSelectionPolicy{
		logger:            logger,
		config:            config,
		selectionStrategy: config.DefaultStrategy,
		currentOptimalSize: int32(config.DefaultOptimalSize),
		adaptiveWeights:   NewAdaptiveWeights(config.AdaptiveWeightsConfig),
		selectionCriteria: NewDynamicSelectionCriteria(config.SelectionCriteriaConfig),
	}
	
	// Initialize components
	if err := policy.initializeComponents(config); err != nil {
		return nil, fmt.Errorf("failed to initialize components: %w", err)
	}
	
	// Start background adaptation process
	policy.startAdaptationProcess()
	
	return policy, nil
}

// SelectClients performs intelligent client selection based on requirements
func (p *AdaptiveClientSelectionPolicy) SelectClients(
	ctx context.Context,
	availableClients []*FLClient,
	requirements *SelectionRequirements,
) ([]*FLClient, error) {
	
	start := time.Now()
	p.logger.Info("Starting adaptive client selection",
		slog.Int("available_clients", len(availableClients)),
		slog.String("rrm_task", string(requirements.RRMTask)),
		slog.String("strategy", string(p.selectionStrategy)))
	
	if len(availableClients) == 0 {
		return nil, errors.NewInvalidInput("no clients available for selection")
	}
	
	// Validate requirements
	if err := p.validateRequirements(requirements); err != nil {
		return nil, fmt.Errorf("invalid requirements: %w", err)
	}
	
	// Pre-filter clients based on basic requirements
	eligibleClients, err := p.preFilterClients(availableClients, requirements)
	if err != nil {
		return nil, fmt.Errorf("pre-filtering failed: %w", err)
	}
	
	if len(eligibleClients) < requirements.MinClients {
		return nil, errors.NewErrorBuilder(errors.ErrServiceUnavailable).
			WithDetail("reason", "insufficient eligible clients").
			WithDetail("required", requirements.MinClients).
			WithDetail("available", len(eligibleClients)).
			Build()
	}
	
	// Determine optimal selection size
	optimalSize := p.determineOptimalSelectionSize(eligibleClients, requirements)
	
	// Perform selection based on current strategy
	var selectedClients []*FLClient
	switch p.selectionStrategy {
	case ClientSelectionRandom:
		selectedClients = p.randomSelection(eligibleClients, optimalSize)
	case ClientSelectionPerformance:
		selectedClients = p.performanceBasedSelection(eligibleClients, optimalSize, requirements)
	case ClientSelectionDiversity:
		selectedClients = p.diversityBasedSelection(eligibleClients, optimalSize, requirements)
	case ClientSelectionAdaptive:
		selectedClients = p.adaptiveSelection(eligibleClients, optimalSize, requirements)
	case ClientSelectionHybrid:
		selectedClients = p.hybridSelection(eligibleClients, optimalSize, requirements)
	case ClientSelectionORANOptimal:
		selectedClients = p.oranOptimalSelection(eligibleClients, optimalSize, requirements)
	default:
		return nil, errors.NewInvalidInput(fmt.Sprintf("unsupported selection strategy: %s", p.selectionStrategy))
	}
	
	// Post-process selection
	finalClients, err := p.postProcessSelection(selectedClients, requirements)
	if err != nil {
		return nil, fmt.Errorf("post-processing failed: %w", err)
	}
	
	// Update selection history and metrics
	p.updateSelectionHistory(finalClients, requirements, time.Since(start))
	
	// Provide feedback to adaptation engine
	p.adaptationEngine.RecordSelection(finalClients, requirements, p.selectionStrategy)
	
	p.logger.Info("Client selection completed",
		slog.Int("selected_count", len(finalClients)),
		slog.String("strategy", string(p.selectionStrategy)),
		slog.Duration("duration", time.Since(start)))
	
	return finalClients, nil
}

// oranOptimalSelection performs O-RAN specific optimal client selection
func (p *AdaptiveClientSelectionPolicy) oranOptimalSelection(
	clients []*FLClient,
	targetSize int,
	requirements *SelectionRequirements,
) []*FLClient {
	
	// Score each client based on O-RAN specific criteria
	clientScores := make([]ClientScore, len(clients))
	
	for i, client := range clients {
		score := p.computeORANOptimalScore(client, requirements)
		clientScores[i] = ClientScore{
			Client: client,
			Score:  score,
		}
	}
	
	// Sort by score (descending)
	sort.Slice(clientScores, func(i, j int) bool {
		return clientScores[i].Score > clientScores[j].Score
	})
	
	// Apply diversity constraints
	selectedClients := p.applyDiversityConstraints(clientScores, targetSize, requirements)
	
	// Apply load balancing
	balancedClients := p.applyLoadBalancing(selectedClients, requirements)
	
	// Ensure network slice distribution
	finalClients := p.ensureNetworkSliceDistribution(balancedClients, requirements)
	
	return finalClients
}

// computeORANOptimalScore calculates O-RAN specific optimization score
func (p *AdaptiveClientSelectionPolicy) computeORANOptimalScore(client *FLClient, requirements *SelectionRequirements) float64 {
	p.mutex.RLock()
	weights := p.adaptiveWeights
	p.mutex.RUnlock()
	
	score := 0.0
	
	// Base performance score
	if metrics, exists := p.clientMetrics.Load(client.ID); exists {
		clientMetrics := metrics.(*ClientMetrics)
		
		// Accuracy weight
		score += weights.Accuracy * clientMetrics.AverageAccuracy
		
		// Latency weight (inverse relationship)
		latencyScore := 1.0 - math.Min(1.0, clientMetrics.TrainingLatency.Seconds()/60.0) // Normalize to 1 minute
		score += weights.Latency * latencyScore
		
		// Reliability weight
		score += weights.Reliability * clientMetrics.ReliabilityScore
		
		// Trust weight
		score += weights.Trust * clientMetrics.TrustScore
		
		// Resource efficiency weight
		resourceScore := 1.0 - (clientMetrics.CPUUtilization + clientMetrics.MemoryUtilization) / 2.0
		score += weights.ResourceEfficiency * resourceScore
	}
	
	// O-RAN specific scoring
	
	// RRM task compatibility
	rrmScore := p.computeRRMTaskCompatibility(client, requirements.RRMTask)
	score += weights.RRMTaskCompatibility * rrmScore
	
	// E2 latency score
	e2Score := p.computeE2LatencyScore(client, requirements.E2LatencyThreshold)
	score += weights.E2Latency * e2Score
	
	// Network slice affinity
	sliceScore := p.computeNetworkSliceAffinity(client, requirements.NetworkSlices)
	score += weights.NetworkSliceAffinity * sliceScore
	
	// Data diversity contribution
	diversityScore := p.computeDataDiversityScore(client, requirements)
	score += weights.DataDiversity * diversityScore
	
	// Security compliance score
	securityScore := p.computeSecurityComplianceScore(client, requirements.SecurityLevel)
	score += weights.SecurityCompliance * securityScore
	
	// Geographic diversity bonus
	geoScore := p.computeGeographicDiversityScore(client)
	score += weights.GeographicDiversity * geoScore
	
	// Load balancing penalty (to avoid overloading popular clients)
	loadPenalty := p.computeLoadBalancingPenalty(client)
	score -= weights.LoadBalancingPenalty * loadPenalty
	
	// Trend bonus (favor improving clients)
	trendBonus := p.computeTrendBonus(client)
	score += weights.TrendBonus * trendBonus
	
	return score
}

// UpdateClientPriority updates the priority of a client based on recent metrics
func (p *AdaptiveClientSelectionPolicy) UpdateClientPriority(clientID string, metrics *ClientMetrics) error {
	p.logger.Debug("Updating client priority",
		slog.String("client_id", clientID),
		slog.Float64("accuracy", metrics.AverageAccuracy),
		slog.Float64("reliability", metrics.ReliabilityScore))
	
	// Store updated metrics
	p.clientMetrics.Store(clientID, metrics)
	
	// Update performance tracker
	if err := p.performanceTracker.UpdateMetrics(clientID, metrics); err != nil {
		return fmt.Errorf("failed to update performance tracker: %w", err)
	}
	
	// Trigger adaptation if significant change detected
	if p.isSignificantChange(clientID, metrics) {
		p.adaptationEngine.TriggerAdaptation("client_priority_change", clientID)
	}
	
	return nil
}

// GetOptimalClientCount determines the optimal number of clients for a task
func (p *AdaptiveClientSelectionPolicy) GetOptimalClientCount(task RRMTaskType, networkConditions *NetworkConditions) int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	// Base optimal count from configuration
	baseCount := p.config.DefaultOptimalSize
	
	// Task-specific adjustments
	taskMultiplier := p.getTaskComplexityMultiplier(task)
	adjustedCount := int(float64(baseCount) * taskMultiplier)
	
	// Network condition adjustments
	if networkConditions != nil {
		networkMultiplier := p.getNetworkConditionMultiplier(networkConditions)
		adjustedCount = int(float64(adjustedCount) * networkMultiplier)
	}
	
	// Apply learned optimal size if available
	if p.currentOptimalSize > 0 {
		learnedFactor := 0.3 // Weight for learned optimal size
		configFactor := 0.7  // Weight for configuration-based size
		
		finalCount := int(learnedFactor*float64(p.currentOptimalSize) + configFactor*float64(adjustedCount))
		adjustedCount = finalCount
	}
	
	// Ensure bounds
	minCount := p.config.MinClients
	maxCount := p.config.MaxClients
	
	if adjustedCount < minCount {
		adjustedCount = minCount
	} else if adjustedCount > maxCount {
		adjustedCount = maxCount
	}
	
	p.logger.Debug("Computed optimal client count",
		slog.String("task", string(task)),
		slog.Int("base_count", baseCount),
		slog.Float64("task_multiplier", taskMultiplier),
		slog.Int("final_count", adjustedCount))
	
	return adjustedCount
}

// AdaptSelection adapts the selection strategy based on round metrics
func (p *AdaptiveClientSelectionPolicy) AdaptSelection(ctx context.Context, roundMetrics *RoundMetrics) error {
	p.logger.Info("Adapting selection strategy",
		slog.Float64("convergence_rate", roundMetrics.ConvergenceRate),
		slog.Duration("round_duration", roundMetrics.RoundDuration),
		slog.Int("participant_count", roundMetrics.ParticipantCount))
	
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Analyze round performance
	adaptationSignal := p.analyzeRoundPerformance(roundMetrics)
	
	// Update adaptive weights based on performance
	if err := p.updateAdaptiveWeights(adaptationSignal); err != nil {
		return fmt.Errorf("failed to update adaptive weights: %w", err)
	}
	
	// Adjust optimal client count
	if err := p.adjustOptimalClientCount(adaptationSignal); err != nil {
		return fmt.Errorf("failed to adjust optimal client count: %w", err)
	}
	
	// Consider strategy change if performance is poor
	if p.shouldChangeStrategy(adaptationSignal) {
		newStrategy := p.selectBetterStrategy(adaptationSignal)
		if newStrategy != p.selectionStrategy {
			p.logger.Info("Changing selection strategy",
				slog.String("old_strategy", string(p.selectionStrategy)),
				slog.String("new_strategy", string(newStrategy)))
			p.selectionStrategy = newStrategy
		}
	}
	
	// Update selection criteria
	if err := p.updateSelectionCriteria(adaptationSignal); err != nil {
		return fmt.Errorf("failed to update selection criteria: %w", err)
	}
	
	return nil
}

// GetSelectionStrategy returns the current selection strategy
func (p *AdaptiveClientSelectionPolicy) GetSelectionStrategy() ClientSelectionStrategy {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.selectionStrategy
}

// Helper methods implementation

func (p *AdaptiveClientSelectionPolicy) preFilterClients(clients []*FLClient, requirements *SelectionRequirements) ([]*FLClient, error) {
	var eligible []*FLClient
	
	for _, client := range clients {
		// Check basic availability
		if client.Status != FLClientStatusIdle && client.Status != FLClientStatusRegistered {
			continue
		}
		
		// Check trust threshold
		if client.TrustScore < requirements.TrustThreshold {
			continue
		}
		
		// Check RRM task compatibility
		if !p.isRRMTaskCompatible(client, requirements.RRMTask) {
			continue
		}
		
		// Check network slice requirements
		if !p.isNetworkSliceCompatible(client, requirements.NetworkSlices) {
			continue
		}
		
		// Check compute requirements
		if !p.meetsComputeRequirements(client, requirements.ComputeRequirement) {
			continue
		}
		
		// Check latency requirements
		if !p.meetsLatencyRequirements(client, requirements.LatencyRequirement) {
			continue
		}
		
		// Check security requirements
		if !p.meetsSecurityRequirements(client, requirements.SecurityLevel) {
			continue
		}
		
		eligible = append(eligible, client)
	}
	
	return eligible, nil
}

func (p *AdaptiveClientSelectionPolicy) determineOptimalSelectionSize(clients []*FLClient, requirements *SelectionRequirements) int {
	// Start with requirements
	minSize := requirements.MinClients
	maxSize := requirements.MaxClients
	
	// Get optimal count for the task
	optimalCount := p.GetOptimalClientCount(requirements.RRMTask, nil)
	
	// Adjust based on available clients
	availableCount := len(clients)
	
	// Choose the minimum of optimal, max requirement, and available
	targetSize := optimalCount
	if targetSize > maxSize {
		targetSize = maxSize
	}
	if targetSize > availableCount {
		targetSize = availableCount
	}
	if targetSize < minSize {
		targetSize = minSize
	}
	
	return targetSize
}

func (p *AdaptiveClientSelectionPolicy) randomSelection(clients []*FLClient, targetSize int) []*FLClient {
	if len(clients) <= targetSize {
		return clients
	}
	
	// Create a copy and shuffle
	clientsCopy := make([]*FLClient, len(clients))
	copy(clientsCopy, clients)
	
	// Shuffle using Fisher-Yates algorithm
	for i := len(clientsCopy) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		clientsCopy[i], clientsCopy[j] = clientsCopy[j], clientsCopy[i]
	}
	
	return clientsCopy[:targetSize]
}

func (p *AdaptiveClientSelectionPolicy) performanceBasedSelection(clients []*FLClient, targetSize int, requirements *SelectionRequirements) []*FLClient {
	// Score clients based on performance metrics
	type clientScore struct {
		client *FLClient
		score  float64
	}
	
	scores := make([]clientScore, len(clients))
	for i, client := range clients {
		score := p.computePerformanceScore(client, requirements)
		scores[i] = clientScore{client: client, score: score}
	}
	
	// Sort by score (descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// Select top performers
	selected := make([]*FLClient, 0, targetSize)
	for i := 0; i < targetSize && i < len(scores); i++ {
		selected = append(selected, scores[i].client)
	}
	
	return selected
}

// Additional helper method implementations would continue here...
// Including computePerformanceScore, diversityBasedSelection, etc.

func (p *AdaptiveClientSelectionPolicy) computePerformanceScore(client *FLClient, requirements *SelectionRequirements) float64 {
	score := 0.0
	
	// Trust score component (0-1)
	score += client.TrustScore * 0.3
	
	// Get cached metrics if available
	if metrics, exists := p.clientMetrics.Load(client.ID); exists {
		clientMetrics := metrics.(*ClientMetrics)
		
		// Accuracy component (0-1)
		score += clientMetrics.AverageAccuracy * 0.25
		
		// Reliability component (0-1)
		score += clientMetrics.ReliabilityScore * 0.25
		
		// Latency component (inverse, 0-1)
		latencyScore := 1.0 - math.Min(1.0, clientMetrics.TrainingLatency.Seconds()/300.0) // Normalize to 5 minutes
		score += latencyScore * 0.2
	}
	
	return score
}