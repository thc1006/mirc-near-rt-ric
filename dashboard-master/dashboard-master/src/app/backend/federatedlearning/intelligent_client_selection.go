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
	"time"
)

// ClientSelectionStrategy defines different client selection approaches
type ClientSelectionStrategy string

const (
	ClientSelectionRandom              ClientSelectionStrategy = "random"
	ClientSelectionPerformanceBased    ClientSelectionStrategy = "performance_based"
	ClientSelectionDiversityBased      ClientSelectionStrategy = "diversity_based"
	ClientSelectionNetworkAware        ClientSelectionStrategy = "network_aware"
	ClientSelectionHybrid             ClientSelectionStrategy = "hybrid"
	ClientSelectionQoSOptimized       ClientSelectionStrategy = "qos_optimized"
)

// ClientSelectionCriteria defines criteria for client selection
type ClientSelectionCriteria struct {
	Strategy                 ClientSelectionStrategy `json:"strategy"`
	MaxClientsPerRound      int                     `json:"max_clients_per_round"`
	MinClientsPerRound      int                     `json:"min_clients_per_round"`
	RequiredDataQuality     float64                 `json:"required_data_quality"`
	MaxNetworkLatency       time.Duration           `json:"max_network_latency"`
	MinComputeCapability    float64                 `json:"min_compute_capability"`
	MinBatteryLevel         float64                 `json:"min_battery_level"`
	RRMTaskRequirements     RRMTaskRequirements     `json:"rrm_task_requirements"`
	GeographicDiversity     bool                    `json:"geographic_diversity"`
	DeviceTypeDiversity     bool                    `json:"device_type_diversity"`
	NetworkTypeDiversity    bool                    `json:"network_type_diversity"`
}

// RRMTaskRequirements defines specific requirements for RRM tasks
type RRMTaskRequirements struct {
	TaskType              RRMTaskType   `json:"task_type"`
	LatencyRequirement    time.Duration `json:"latency_requirement"`
	ThroughputRequirement float64       `json:"throughput_requirement"`
	ReliabilityRequirement float64      `json:"reliability_requirement"`
	ComputeIntensity      float64       `json:"compute_intensity"`
	DataSizeRequirement   int64         `json:"data_size_requirement"`
	RealTimeProcessing    bool          `json:"real_time_processing"`
}

// ClientSelectionResult contains the results of client selection
type ClientSelectionResult struct {
	SelectedClients      []string                `json:"selected_clients"`
	SelectionReason      map[string]string       `json:"selection_reason"`
	RejectedClients      []string                `json:"rejected_clients"`
	RejectionReason      map[string]string       `json:"rejection_reason"`
	SelectionMetrics     ClientSelectionMetrics  `json:"selection_metrics"`
	SelectionTimestamp   time.Time               `json:"selection_timestamp"`
	SelectionDuration    time.Duration           `json:"selection_duration"`
}

// ClientSelectionMetrics provides metrics about the selection process
type ClientSelectionMetrics struct {
	TotalCandidates        int     `json:"total_candidates"`
	EligibleCandidates     int     `json:"eligible_candidates"`
	SelectedCount          int     `json:"selected_count"`
	AveragePerformanceScore float64 `json:"average_performance_score"`
	AverageNetworkQuality   float64 `json:"average_network_quality"`
	AverageComputeCapability float64 `json:"average_compute_capability"`
	GeographicCoverage      float64 `json:"geographic_coverage"`
	DeviceTypeDiversity     float64 `json:"device_type_diversity"`
	NetworkTypeDiversity    float64 `json:"network_type_diversity"`
}

// AvailabilityWindow represents client availability patterns
type AvailabilityWindow struct {
	TimeZone           string                    `json:"timezone"`
	DailyAvailability  map[int]AvailabilitySlot `json:"daily_availability"` // Hour -> Availability
	WeeklyPattern      map[time.Weekday]float64  `json:"weekly_pattern"`      // Day -> Availability
	SeasonalPattern    map[string]float64        `json:"seasonal_pattern"`    // Season -> Availability
	CurrentAvailability float64                  `json:"current_availability"`
	NextAvailableTime   *time.Time               `json:"next_available_time"`
}

// AvailabilitySlot represents availability during a specific time slot
type AvailabilitySlot struct {
	StartHour      int     `json:"start_hour"`
	EndHour        int     `json:"end_hour"`
	Availability   float64 `json:"availability"`
	QualityFactor  float64 `json:"quality_factor"`
}

// PerformanceHistory tracks historical client performance
type PerformanceHistory struct {
	TotalRounds          int                    `json:"total_rounds"`
	SuccessfulRounds     int                    `json:"successful_rounds"`
	AverageLatency       time.Duration          `json:"average_latency"`
	AverageAccuracy      float64                `json:"average_accuracy"`
	AverageDataQuality   float64                `json:"average_data_quality"`
	ReliabilityScore     float64                `json:"reliability_score"`
	RecentPerformance    []PerformanceRecord    `json:"recent_performance"`
	PerformanceTrend     PerformanceTrend       `json:"performance_trend"`
}

// PerformanceRecord represents a single performance measurement
type PerformanceRecord struct {
	Timestamp      time.Time     `json:"timestamp"`
	RoundID        string        `json:"round_id"`
	TaskType       RRMTaskType   `json:"task_type"`
	Latency        time.Duration `json:"latency"`
	Accuracy       float64       `json:"accuracy"`
	DataQuality    float64       `json:"data_quality"`
	Success        bool          `json:"success"`
	FailureReason  string        `json:"failure_reason,omitempty"`
}

// PerformanceTrend indicates the trend in client performance
type PerformanceTrend string

const (
	PerformanceTrendImproving PerformanceTrend = "improving"
	PerformanceTrendStable    PerformanceTrend = "stable"
	PerformanceTrendDeclining PerformanceTrend = "declining"
	PerformanceTrendUnknown   PerformanceTrend = "unknown"
)

// clientSelectionUpdateLoop continuously updates client selection information
func (drm *DynamicResourceManager) clientSelectionUpdateLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // Update every 5 minutes
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := drm.updateClientSelectionData(ctx); err != nil {
				drm.logger.Error("Failed to update client selection data", 
					slog.String("error", err.Error()))
			}
		}
	}
}

// SelectOptimalClients selects the optimal set of clients for a federated learning round
func (drm *DynamicResourceManager) SelectOptimalClients(
	ctx context.Context,
	criteria ClientSelectionCriteria,
) (*ClientSelectionResult, error) {
	
	start := time.Now()
	
	drm.logger.Info("Starting client selection",
		slog.String("strategy", string(criteria.Strategy)),
		slog.Int("max_clients", criteria.MaxClientsPerRound),
		slog.String("rrm_task", string(criteria.RRMTaskRequirements.TaskType)))
	
	// Get all available clients
	candidateClients := drm.getCandidateClients(ctx)
	
	// Filter clients based on basic criteria
	eligibleClients := drm.filterEligibleClients(candidateClients, criteria)
	
	// Select clients based on strategy
	selectedClients, selectionReasons := drm.selectClientsByStrategy(eligibleClients, criteria)
	
	// Identify rejected clients and reasons
	rejectedClients, rejectionReasons := drm.identifyRejectedClients(candidateClients, selectedClients, criteria)
	
	// Calculate selection metrics
	metrics := drm.calculateSelectionMetrics(candidateClients, eligibleClients, selectedClients)
	
	result := &ClientSelectionResult{
		SelectedClients:      selectedClients,
		SelectionReason:      selectionReasons,
		RejectedClients:      rejectedClients,
		RejectionReason:      rejectionReasons,
		SelectionMetrics:     metrics,
		SelectionTimestamp:   time.Now(),
		SelectionDuration:    time.Since(start),
	}
	
	drm.logger.Info("Client selection completed",
		slog.Int("candidates", len(candidateClients)),
		slog.Int("eligible", len(eligibleClients)),
		slog.Int("selected", len(selectedClients)),
		slog.Duration("duration", result.SelectionDuration))
	
	return result, nil
}

// getCandidateClients retrieves all potential candidate clients
func (drm *DynamicResourceManager) getCandidateClients(ctx context.Context) []string {
	drm.mutex.RLock()
	defer drm.mutex.RUnlock()
	
	candidates := make([]string, 0, len(drm.clientResourceCache))
	for clientID := range drm.clientResourceCache {
		candidates = append(candidates, clientID)
	}
	
	return candidates
}

// filterEligibleClients filters clients based on basic eligibility criteria
func (drm *DynamicResourceManager) filterEligibleClients(
	candidates []string,
	criteria ClientSelectionCriteria,
) []string {
	
	eligible := make([]string, 0, len(candidates))
	
	drm.mutex.RLock()
	defer drm.mutex.RUnlock()
	
	for _, clientID := range candidates {
		status, exists := drm.clientResourceCache[clientID]
		if !exists {
			continue
		}
		
		// Check basic eligibility criteria
		if drm.isClientEligible(status, criteria) {
			eligible = append(eligible, clientID)
		}
	}
	
	return eligible
}

// isClientEligible checks if a client meets basic eligibility requirements
func (drm *DynamicResourceManager) isClientEligible(
	status *ClientResourceStatus,
	criteria ClientSelectionCriteria,
) bool {
	
	// Check network latency requirement
	if status.NetworkCondition.Latency > criteria.MaxNetworkLatency {
		return false
	}
	
	// Check compute capability requirement
	if status.ComputeCapacity.ComputeScore < criteria.MinComputeCapability {
		return false
	}
	
	// Check battery level requirement (for mobile devices)
	if !status.EnergyConstraints.IsPluggedIn && 
		status.EnergyConstraints.BatteryLevel < criteria.MinBatteryLevel {
		return false
	}
	
	// Check data quality requirement
	if status.HistoricalPerformance.AverageDataQuality < criteria.RequiredDataQuality {
		return false
	}
	
	// Check current availability
	if status.AvailabilityWindow.CurrentAvailability < 0.5 { // At least 50% available
		return false
	}
	
	// Check task-specific requirements
	if !drm.meetsTaskRequirements(status, criteria.RRMTaskRequirements) {
		return false
	}
	
	return true
}

// meetsTaskRequirements checks if client meets specific RRM task requirements
func (drm *DynamicResourceManager) meetsTaskRequirements(
	status *ClientResourceStatus,
	requirements RRMTaskRequirements,
) bool {
	
	// Check latency requirement
	if status.NetworkCondition.Latency > requirements.LatencyRequirement {
		return false
	}
	
	// Check throughput requirement
	if status.NetworkCondition.Bandwidth < requirements.ThroughputRequirement {
		return false
	}
	
	// Check reliability requirement
	if status.HistoricalPerformance.ReliabilityScore < requirements.ReliabilityRequirement {
		return false
	}
	
	// Check compute intensity requirement
	if status.ComputeCapacity.ComputeScore < requirements.ComputeIntensity {
		return false
	}
	
	// For real-time processing tasks, require higher capabilities
	if requirements.RealTimeProcessing {
		if status.NetworkCondition.Latency > 100*time.Millisecond {
			return false
		}
		if status.ComputeCapacity.ComputeScore < 0.7 {
			return false
		}
	}
	
	return true
}

// selectClientsByStrategy selects clients based on the specified strategy
func (drm *DynamicResourceManager) selectClientsByStrategy(
	eligibleClients []string,
	criteria ClientSelectionCriteria,
) ([]string, map[string]string) {
	
	switch criteria.Strategy {
	case ClientSelectionPerformanceBased:
		return drm.selectByPerformance(eligibleClients, criteria)
	case ClientSelectionNetworkAware:
		return drm.selectByNetworkCondition(eligibleClients, criteria)
	case ClientSelectionDiversityBased:
		return drm.selectByDiversity(eligibleClients, criteria)
	case ClientSelectionQoSOptimized:
		return drm.selectByQoS(eligibleClients, criteria)
	case ClientSelectionHybrid:
		return drm.selectByHybridApproach(eligibleClients, criteria)
	default:
		return drm.selectRandomly(eligibleClients, criteria)
	}
}

// selectByPerformance selects clients based on historical performance
func (drm *DynamicResourceManager) selectByPerformance(
	eligibleClients []string,
	criteria ClientSelectionCriteria,
) ([]string, map[string]string) {
	
	type clientScore struct {
		clientID string
		score    float64
	}
	
	// Calculate performance scores
	scores := make([]clientScore, 0, len(eligibleClients))
	
	drm.mutex.RLock()
	for _, clientID := range eligibleClients {
		if status, exists := drm.clientResourceCache[clientID]; exists {
			score := drm.calculatePerformanceScore(status)
			scores = append(scores, clientScore{clientID: clientID, score: score})
		}
	}
	drm.mutex.RUnlock()
	
	// Sort by score (highest first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// Select top performers
	maxClients := min(criteria.MaxClientsPerRound, len(scores))
	selected := make([]string, maxClients)
	reasons := make(map[string]string)
	
	for i := 0; i < maxClients; i++ {
		selected[i] = scores[i].clientID
		reasons[scores[i].clientID] = fmt.Sprintf("High performance score: %.3f", scores[i].score)
	}
	
	return selected, reasons
}

// selectByNetworkCondition selects clients based on network conditions
func (drm *DynamicResourceManager) selectByNetworkCondition(
	eligibleClients []string,
	criteria ClientSelectionCriteria,
) ([]string, map[string]string) {
	
	type clientNetworkScore struct {
		clientID string
		score    float64
		latency  time.Duration
	}
	
	// Calculate network scores
	scores := make([]clientNetworkScore, 0, len(eligibleClients))
	
	drm.mutex.RLock()
	for _, clientID := range eligibleClients {
		if status, exists := drm.clientResourceCache[clientID]; exists {
			score := drm.calculateNetworkScore(status.NetworkCondition)
			scores = append(scores, clientNetworkScore{
				clientID: clientID,
				score:    score,
				latency:  status.NetworkCondition.Latency,
			})
		}
	}
	drm.mutex.RUnlock()
	
	// Sort by network score (highest first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// Select clients with best network conditions
	maxClients := min(criteria.MaxClientsPerRound, len(scores))
	selected := make([]string, maxClients)
	reasons := make(map[string]string)
	
	for i := 0; i < maxClients; i++ {
		selected[i] = scores[i].clientID
		reasons[scores[i].clientID] = fmt.Sprintf("Excellent network: %.3f score, %dms latency", 
			scores[i].score, scores[i].latency.Milliseconds())
	}
	
	return selected, reasons
}

// selectByDiversity selects clients to maximize diversity
func (drm *DynamicResourceManager) selectByDiversity(
	eligibleClients []string,
	criteria ClientSelectionCriteria,
) ([]string, map[string]string) {
	
	selected := make([]string, 0, criteria.MaxClientsPerRound)
	reasons := make(map[string]string)
	
	// Track diversity dimensions
	deviceTypes := make(map[string]bool)
	networkTypes := make(map[string]bool)
	locations := make(map[string]bool)
	
	drm.mutex.RLock()
	defer drm.mutex.RUnlock()
	
	// Select clients to maximize diversity
	for _, clientID := range eligibleClients {
		if len(selected) >= criteria.MaxClientsPerRound {
			break
		}
		
		status, exists := drm.clientResourceCache[clientID]
		if !exists {
			continue
		}
		
		// Extract diversity characteristics
		deviceType := drm.getDeviceType(status)
		networkType := status.NetworkCondition.ConnectionType
		location := drm.getLocation(status)
		
		// Check if this client adds diversity
		addsDiversity := false
		diversityReasons := make([]string, 0)
		
		if criteria.DeviceTypeDiversity && !deviceTypes[deviceType] {
			addsDiversity = true
			deviceTypes[deviceType] = true
			diversityReasons = append(diversityReasons, fmt.Sprintf("device:%s", deviceType))
		}
		
		if criteria.NetworkTypeDiversity && !networkTypes[networkType] {
			addsDiversity = true
			networkTypes[networkType] = true
			diversityReasons = append(diversityReasons, fmt.Sprintf("network:%s", networkType))
		}
		
		if criteria.GeographicDiversity && !locations[location] {
			addsDiversity = true
			locations[location] = true
			diversityReasons = append(diversityReasons, fmt.Sprintf("location:%s", location))
		}
		
		// Select if adds diversity or if we haven't reached minimum
		if addsDiversity || len(selected) < criteria.MinClientsPerRound {
			selected = append(selected, clientID)
			if len(diversityReasons) > 0 {
				reasons[clientID] = fmt.Sprintf("Diversity: %v", diversityReasons)
			} else {
				reasons[clientID] = "Required for minimum count"
			}
		}
	}
	
	return selected, reasons
}

// selectByQoS selects clients optimized for QoS requirements
func (drm *DynamicResourceManager) selectByQoS(
	eligibleClients []string,
	criteria ClientSelectionCriteria,
) ([]string, map[string]string) {
	
	type clientQoSScore struct {
		clientID string
		score    float64
		metrics  map[string]float64
	}
	
	// Calculate QoS scores based on task requirements
	scores := make([]clientQoSScore, 0, len(eligibleClients))
	
	drm.mutex.RLock()
	for _, clientID := range eligibleClients {
		if status, exists := drm.clientResourceCache[clientID]; exists {
			score, metrics := drm.calculateQoSScore(status, criteria.RRMTaskRequirements)
			scores = append(scores, clientQoSScore{
				clientID: clientID,
				score:    score,
				metrics:  metrics,
			})
		}
	}
	drm.mutex.RUnlock()
	
	// Sort by QoS score (highest first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})
	
	// Select clients with best QoS characteristics
	maxClients := min(criteria.MaxClientsPerRound, len(scores))
	selected := make([]string, maxClients)
	reasons := make(map[string]string)
	
	for i := 0; i < maxClients; i++ {
		selected[i] = scores[i].clientID
		reasons[scores[i].clientID] = fmt.Sprintf("QoS optimized: %.3f score", scores[i].score)
	}
	
	return selected, reasons
}

// selectByHybridApproach uses a combination of all strategies
func (drm *DynamicResourceManager) selectByHybridApproach(
	eligibleClients []string,
	criteria ClientSelectionCriteria,
) ([]string, map[string]string) {
	
	type clientHybridScore struct {
		clientID         string
		totalScore       float64
		performanceScore float64
		networkScore     float64
		qosScore         float64
		diversityBonus   float64
	}
	
	// Calculate hybrid scores
	scores := make([]clientHybridScore, 0, len(eligibleClients))
	
	drm.mutex.RLock()
	for _, clientID := range eligibleClients {
		if status, exists := drm.clientResourceCache[clientID]; exists {
			perfScore := drm.calculatePerformanceScore(status)
			netScore := drm.calculateNetworkScore(status.NetworkCondition)
			qosScore, _ := drm.calculateQoSScore(status, criteria.RRMTaskRequirements)
			
			// Weighted combination
			totalScore := 0.4*perfScore + 0.3*netScore + 0.3*qosScore
			
			scores = append(scores, clientHybridScore{
				clientID:         clientID,
				totalScore:       totalScore,
				performanceScore: perfScore,
				networkScore:     netScore,
				qosScore:         qosScore,
			})
		}
	}
	drm.mutex.RUnlock()
	
	// Apply diversity bonus
	drm.applyDiversityBonus(scores, criteria)
	
	// Sort by total score (highest first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].totalScore > scores[j].totalScore
	})
	
	// Select top clients
	maxClients := min(criteria.MaxClientsPerRound, len(scores))
	selected := make([]string, maxClients)
	reasons := make(map[string]string)
	
	for i := 0; i < maxClients; i++ {
		selected[i] = scores[i].clientID
		reasons[scores[i].clientID] = fmt.Sprintf(
			"Hybrid: %.3f (perf:%.2f, net:%.2f, qos:%.2f, div:%.2f)",
			scores[i].totalScore, scores[i].performanceScore, 
			scores[i].networkScore, scores[i].qosScore, scores[i].diversityBonus)
	}
	
	return selected, reasons
}

// selectRandomly selects clients randomly from eligible candidates
func (drm *DynamicResourceManager) selectRandomly(
	eligibleClients []string,
	criteria ClientSelectionCriteria,
) ([]string, map[string]string) {
	
	// Shuffle the list
	shuffled := make([]string, len(eligibleClients))
	copy(shuffled, eligibleClients)
	
	// Fisher-Yates shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		j := int(time.Now().UnixNano()) % (i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	
	// Select the first N clients
	maxClients := min(criteria.MaxClientsPerRound, len(shuffled))
	selected := shuffled[:maxClients]
	
	reasons := make(map[string]string)
	for _, clientID := range selected {
		reasons[clientID] = "Random selection"
	}
	
	return selected, reasons
}

// Helper methods for score calculations and client management
func (drm *DynamicResourceManager) calculatePerformanceScore(status *ClientResourceStatus) float64 {
	history := status.HistoricalPerformance
	
	// Combine multiple performance factors
	successRate := float64(history.SuccessfulRounds) / float64(max(history.TotalRounds, 1))
	reliabilityScore := history.ReliabilityScore
	dataQualityScore := history.AverageDataQuality
	accuracyScore := history.AverageAccuracy
	
	// Apply trend bonus/penalty
	trendMultiplier := 1.0
	switch history.PerformanceTrend {
	case PerformanceTrendImproving:
		trendMultiplier = 1.1
	case PerformanceTrendDeclining:
		trendMultiplier = 0.9
	}
	
	score := (0.3*successRate + 0.25*reliabilityScore + 0.25*dataQualityScore + 0.2*accuracyScore) * trendMultiplier
	return min(score, 1.0)
}

func (drm *DynamicResourceManager) calculateNetworkScore(condition NetworkCondition) float64 {
	return condition.NetworkQuality
}

func (drm *DynamicResourceManager) calculateQoSScore(
	status *ClientResourceStatus,
	requirements RRMTaskRequirements,
) (float64, map[string]float64) {
	
	metrics := make(map[string]float64)
	
	// Latency score (lower is better)
	latencyScore := 1.0 - min(float64(status.NetworkCondition.Latency.Milliseconds())/1000.0, 1.0)
	metrics["latency"] = latencyScore
	
	// Reliability score
	reliabilityScore := status.HistoricalPerformance.ReliabilityScore
	metrics["reliability"] = reliabilityScore
	
	// Compute capability score
	computeScore := status.ComputeCapacity.ComputeScore
	metrics["compute"] = computeScore
	
	// Energy stability score
	energyScore := 1.0
	if !status.EnergyConstraints.IsPluggedIn {
		energyScore = status.EnergyConstraints.BatteryLevel / 100.0
	}
	metrics["energy"] = energyScore
	
	// Weighted combination based on task requirements
	weights := drm.getQoSWeights(requirements)
	totalScore := weights.latency*latencyScore + 
		weights.reliability*reliabilityScore + 
		weights.compute*computeScore + 
		weights.energy*energyScore
	
	return totalScore, metrics
}

// updateClientSelectionData updates client selection related data
func (drm *DynamicResourceManager) updateClientSelectionData(ctx context.Context) error {
	// Update network latency matrix
	if err := drm.updateNetworkLatencyMatrix(ctx); err != nil {
		drm.logger.Warn("Failed to update network latency matrix", 
			slog.String("error", err.Error()))
	}
	
	// Update client performance trends
	drm.updateClientPerformanceTrends()
	
	// Update availability windows
	drm.updateClientAvailabilityWindows()
	
	return nil
}

// Helper structures and methods
type qosWeights struct {
	latency     float64
	reliability float64
	compute     float64
	energy      float64
}

func (drm *DynamicResourceManager) getQoSWeights(requirements RRMTaskRequirements) qosWeights {
	switch requirements.TaskType {
	case RRMTaskHandoverOptimization:
		return qosWeights{latency: 0.5, reliability: 0.3, compute: 0.15, energy: 0.05}
	case RRMTaskResourceAllocation:
		return qosWeights{latency: 0.3, reliability: 0.25, compute: 0.35, energy: 0.1}
	case RRMTaskAnomalyDetection:
		return qosWeights{latency: 0.4, reliability: 0.35, compute: 0.2, energy: 0.05}
	default:
		return qosWeights{latency: 0.25, reliability: 0.25, compute: 0.25, energy: 0.25}
	}
}

// Additional helper methods would be implemented here...
func (drm *DynamicResourceManager) getDeviceType(status *ClientResourceStatus) string {
	// Implementation to determine device type from client status
	return "mobile" // Placeholder
}

func (drm *DynamicResourceManager) getLocation(status *ClientResourceStatus) string {
	// Implementation to determine geographic location from client status
	return "region1" // Placeholder
}

func (drm *DynamicResourceManager) applyDiversityBonus(scores []clientHybridScore, criteria ClientSelectionCriteria) {
	// Implementation to apply diversity bonuses to hybrid scores
}

func (drm *DynamicResourceManager) identifyRejectedClients(
	candidates []string,
	selected []string,
	criteria ClientSelectionCriteria,
) ([]string, map[string]string) {
	
	selectedMap := make(map[string]bool)
	for _, clientID := range selected {
		selectedMap[clientID] = true
	}
	
	rejected := make([]string, 0)
	reasons := make(map[string]string)
	
	for _, clientID := range candidates {
		if !selectedMap[clientID] {
			rejected = append(rejected, clientID)
			reasons[clientID] = "Not selected by strategy"
		}
	}
	
	return rejected, reasons
}

func (drm *DynamicResourceManager) calculateSelectionMetrics(
	candidates []string,
	eligible []string,
	selected []string,
) ClientSelectionMetrics {
	
	return ClientSelectionMetrics{
		TotalCandidates:    len(candidates),
		EligibleCandidates: len(eligible),
		SelectedCount:      len(selected),
		// Other metrics would be calculated here
	}
}

func (drm *DynamicResourceManager) updateNetworkLatencyMatrix(ctx context.Context) error {
	// Implementation for updating network latency measurements between clients
	return nil
}

func (drm *DynamicResourceManager) updateClientPerformanceTrends() {
	// Implementation for analyzing and updating client performance trends
}

func (drm *DynamicResourceManager) updateClientAvailabilityWindows() {
	// Implementation for updating client availability predictions
}