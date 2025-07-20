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
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// NodeConditionStatus represents the status of a node condition
type NodeConditionStatus struct {
	Type               corev1.NodeConditionType `json:"type"`
	Status             corev1.ConditionStatus   `json:"status"`
	LastTransitionTime time.Time                `json:"last_transition_time"`
	Reason             string                   `json:"reason"`
	Message            string                   `json:"message"`
}

// resourceMonitoringLoop continuously monitors cluster resources
func (drm *DynamicResourceManager) resourceMonitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(drm.updateInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := drm.updateResourceMetrics(ctx); err != nil {
				drm.logger.Error("Failed to update resource metrics", 
					slog.String("error", err.Error()))
			}
		}
	}
}

// updateResourceMetrics fetches and updates all resource metrics
func (drm *DynamicResourceManager) updateResourceMetrics(ctx context.Context) error {
	start := time.Now()
	
	// Update node resource status
	if err := drm.updateNodeResourceStatus(ctx); err != nil {
		return fmt.Errorf("failed to update node resource status: %w", err)
	}
	
	// Update cluster utilization
	utilization, err := drm.calculateClusterUtilization(ctx)
	if err != nil {
		return fmt.Errorf("failed to calculate cluster utilization: %w", err)
	}
	
	drm.mutex.Lock()
	drm.currentUtilization = utilization
	drm.mutex.Unlock()
	
	// Update client capacity scores
	drm.updateClientCapacityScores()
	
	drm.logger.Debug("Resource metrics updated successfully",
		slog.Duration("duration", time.Since(start)),
		slog.Float64("cpu_utilization", utilization.OverallCPUUtilization),
		slog.Float64("memory_utilization", utilization.OverallMemoryUtilization))
	
	return nil
}

// updateNodeResourceStatus fetches current node resource status from Kubernetes
func (drm *DynamicResourceManager) updateNodeResourceStatus(ctx context.Context) error {
	// Get all nodes
	nodes, err := drm.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}
	
	// Get node metrics
	nodeMetrics, err := drm.metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		drm.logger.Warn("Failed to get node metrics, using node capacity only", 
			slog.String("error", err.Error()))
		// Continue with capacity-only information
	}
	
	// Create metrics map for quick lookup
	metricsMap := make(map[string]*metricsv1beta1.NodeMetrics)
	if nodeMetrics != nil {
		for i := range nodeMetrics.Items {
			metric := &nodeMetrics.Items[i]
			metricsMap[metric.Name] = metric
		}
	}
	
	drm.mutex.Lock()
	defer drm.mutex.Unlock()
	
	// Update each node's resource status
	for i := range nodes.Items {
		node := &nodes.Items[i]
		status := drm.calculateNodeResourceStatus(node, metricsMap[node.Name])
		drm.nodeResourceCache[node.Name] = status
	}
	
	return nil
}

// calculateNodeResourceStatus computes comprehensive node resource status
func (drm *DynamicResourceManager) calculateNodeResourceStatus(
	node *corev1.Node, 
	metrics *metricsv1beta1.NodeMetrics,
) *NodeResourceStatus {
	
	status := &NodeResourceStatus{
		NodeName:      node.Name,
		Timestamp:     time.Now(),
		NodeLabels:    node.Labels,
		NodeConditions: make([]NodeConditionStatus, 0),
	}
	
	// Extract resource capacity and allocatable
	if cpu := node.Status.Capacity[corev1.ResourceCPU]; !cpu.IsZero() {
		status.CPUCapacity = float64(cpu.MilliValue()) / 1000.0
	}
	if memory := node.Status.Capacity[corev1.ResourceMemory]; !memory.IsZero() {
		status.MemoryCapacity = memory.Value()
	}
	if cpu := node.Status.Allocatable[corev1.ResourceCPU]; !cpu.IsZero() {
		status.CPUAllocatable = float64(cpu.MilliValue()) / 1000.0
	}
	if memory := node.Status.Allocatable[corev1.ResourceMemory]; !memory.IsZero() {
		status.MemoryAllocatable = memory.Value()
	}
	
	// Handle GPU resources if available
	if gpu, exists := node.Status.Capacity["nvidia.com/gpu"]; exists && !gpu.IsZero() {
		status.GPUCapacity = int(gpu.Value())
	}
	
	// Extract current usage from metrics if available
	if metrics != nil {
		if cpu := metrics.Usage[corev1.ResourceCPU]; !cpu.IsZero() {
			status.CPUUsage = float64(cpu.MilliValue()) / 1000.0
		}
		if memory := metrics.Usage[corev1.ResourceMemory]; !memory.IsZero() {
			status.MemoryUsage = memory.Value()
		}
	}
	
	// Extract node conditions
	for _, condition := range node.Status.Conditions {
		conditionStatus := NodeConditionStatus{
			Type:               condition.Type,
			Status:             condition.Status,
			LastTransitionTime: condition.LastTransitionTime.Time,
			Reason:             condition.Reason,
			Message:            condition.Message,
		}
		status.NodeConditions = append(status.NodeConditions, conditionStatus)
	}
	
	// Calculate network bandwidth estimate (placeholder - would be enhanced with real metrics)
	status.NetworkBandwidth = drm.estimateNetworkBandwidth(node)
	status.NetworkUtilization = drm.estimateNetworkUtilization(node)
	
	// Calculate overall availability score
	status.AvailableScore = drm.calculateNodeAvailabilityScore(status)
	
	return status
}

// calculateClusterUtilization computes overall cluster resource utilization
func (drm *DynamicResourceManager) calculateClusterUtilization(ctx context.Context) (*ClusterUtilization, error) {
	utilization := &ClusterUtilization{
		Timestamp: time.Now(),
	}
	
	// Get current resource usage across all nodes
	var totalCPUCapacity, totalCPUUsage float64
	var totalMemoryCapacity, totalMemoryUsage int64
	var totalGPUCapacity, totalGPUUsage int
	var availableNodeCount int
	
	drm.mutex.RLock()
	for _, nodeStatus := range drm.nodeResourceCache {
		if drm.isNodeReady(nodeStatus) {
			availableNodeCount++
			totalCPUCapacity += nodeStatus.CPUAllocatable
			totalCPUUsage += nodeStatus.CPUUsage
			totalMemoryCapacity += nodeStatus.MemoryAllocatable
			totalMemoryUsage += nodeStatus.MemoryUsage
			totalGPUCapacity += nodeStatus.GPUCapacity
			totalGPUUsage += nodeStatus.GPUUsage
		}
	}
	drm.mutex.RUnlock()
	
	// Calculate utilization percentages
	if totalCPUCapacity > 0 {
		utilization.OverallCPUUtilization = totalCPUUsage / totalCPUCapacity
	}
	if totalMemoryCapacity > 0 {
		utilization.OverallMemoryUtilization = float64(totalMemoryUsage) / float64(totalMemoryCapacity)
	}
	if totalGPUCapacity > 0 {
		utilization.OverallGPUUtilization = float64(totalGPUUsage) / float64(totalGPUCapacity)
	}
	
	// Count active and queued jobs
	drm.mutex.RLock()
	utilization.ActiveJobCount = len(drm.activeJobs)
	if drm.jobSchedulingQueue != nil {
		utilization.QueuedJobCount = drm.jobSchedulingQueue.Size()
	}
	drm.mutex.RUnlock()
	
	utilization.AvailableNodeCount = availableNodeCount
	
	// Add prediction if ML model is available
	if drm.resourcePredictorMLModel != nil {
		prediction, err := drm.resourcePredictorMLModel.PredictUtilization(ctx, utilization)
		if err != nil {
			drm.logger.Warn("Failed to generate resource prediction", 
				slog.String("error", err.Error()))
		} else {
			utilization.PredictedUtilization = prediction
		}
	}
	
	return utilization, nil
}

// updateClientCapacityScores recalculates client capacity scores
func (drm *DynamicResourceManager) updateClientCapacityScores() {
	drm.mutex.Lock()
	defer drm.mutex.Unlock()
	
	for clientID, clientStatus := range drm.clientResourceCache {
		score := drm.calculateClientCapacityScore(clientStatus)
		drm.clientCapacityScores[clientID] = score
	}
}

// calculateClientCapacityScore computes a unified capacity score for a client
func (drm *DynamicResourceManager) calculateClientCapacityScore(status *ClientResourceStatus) float64 {
	if status == nil {
		return 0.0
	}
	
	// Compute component scores (0-1 scale)
	computeScore := drm.normalizeComputeScore(status.ComputeCapacity)
	networkScore := drm.normalizeNetworkScore(status.NetworkCondition)
	energyScore := drm.normalizeEnergyScore(status.EnergyConstraints)
	availabilityScore := drm.normalizeAvailabilityScore(status.AvailabilityWindow)
	performanceScore := drm.normalizePerformanceScore(status.HistoricalPerformance)
	
	// Weighted combination
	weights := struct {
		compute     float64
		network     float64
		energy      float64
		availability float64
		performance float64
	}{
		compute:     0.35,
		network:     0.25,
		energy:      0.15,
		availability: 0.15,
		performance: 0.10,
	}
	
	totalScore := weights.compute*computeScore +
		weights.network*networkScore +
		weights.energy*energyScore +
		weights.availability*availabilityScore +
		weights.performance*performanceScore
	
	return totalScore
}

// Helper functions for score normalization
func (drm *DynamicResourceManager) normalizeComputeScore(capability ComputeCapability) float64 {
	// Normalize based on typical ranges
	cpuScore := float64(capability.CPUCores) / 16.0 // Assume 16 cores as reference
	memoryScore := capability.MemoryGB / 32.0       // Assume 32GB as reference
	
	score := (cpuScore + memoryScore) / 2.0
	if capability.HasGPU {
		gpuScore := capability.GPUMemoryGB / 16.0 // Assume 16GB GPU memory as reference
		score = (score + gpuScore) / 2.0
	}
	
	return min(score, 1.0)
}

func (drm *DynamicResourceManager) normalizeNetworkScore(condition NetworkCondition) float64 {
	// Lower latency is better
	latencyScore := 1.0 - min(float64(condition.Latency.Milliseconds())/1000.0, 1.0)
	
	// Higher bandwidth is better
	bandwidthScore := min(condition.Bandwidth/1000.0, 1.0) // 1Gbps as reference
	
	// Lower packet loss is better
	packetLossScore := 1.0 - min(condition.PacketLoss/10.0, 1.0) // 10% as maximum
	
	return (latencyScore + bandwidthScore + packetLossScore) / 3.0
}

func (drm *DynamicResourceManager) normalizeEnergyScore(constraints EnergyConstraints) float64 {
	if constraints.IsPluggedIn {
		return 1.0 // No energy constraints when plugged in
	}
	
	batteryScore := constraints.BatteryLevel / 100.0
	return batteryScore
}

func (drm *DynamicResourceManager) normalizeAvailabilityScore(window AvailabilityWindow) float64 {
	// This would be implemented based on the AvailabilityWindow structure
	// For now, return a placeholder
	return 0.8
}

func (drm *DynamicResourceManager) normalizePerformanceScore(performance PerformanceHistory) float64 {
	// This would be implemented based on historical performance data
	// For now, return a placeholder
	return 0.75
}

// Helper functions
func (drm *DynamicResourceManager) isNodeReady(status *NodeResourceStatus) bool {
	for _, condition := range status.NodeConditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (drm *DynamicResourceManager) estimateNetworkBandwidth(node *corev1.Node) float64 {
	// Placeholder implementation - would be enhanced with real network metrics
	if nodeType, exists := node.Labels["node.kubernetes.io/instance-type"]; exists {
		switch nodeType {
		case "m5.large", "m5.xlarge":
			return 1000.0 // 1 Gbps
		case "m5.2xlarge", "m5.4xlarge":
			return 10000.0 // 10 Gbps
		default:
			return 1000.0
		}
	}
	return 1000.0
}

func (drm *DynamicResourceManager) estimateNetworkUtilization(node *corev1.Node) float64 {
	// Placeholder implementation - would be enhanced with real network metrics
	return 0.3 // 30% utilization estimate
}

func (drm *DynamicResourceManager) calculateNodeAvailabilityScore(status *NodeResourceStatus) float64 {
	score := 1.0
	
	// Check CPU availability
	if status.CPUAllocatable > 0 {
		cpuAvailable := (status.CPUAllocatable - status.CPUUsage) / status.CPUAllocatable
		score *= cpuAvailable
	}
	
	// Check memory availability
	if status.MemoryAllocatable > 0 {
		memoryAvailable := float64(status.MemoryAllocatable-status.MemoryUsage) / float64(status.MemoryAllocatable)
		score *= memoryAvailable
	}
	
	// Penalize for node conditions
	for _, condition := range status.NodeConditions {
		if condition.Type == corev1.NodeReady && condition.Status != corev1.ConditionTrue {
			score *= 0.5
		}
		if condition.Type == corev1.NodeMemoryPressure && condition.Status == corev1.ConditionTrue {
			score *= 0.7
		}
		if condition.Type == corev1.NodeDiskPressure && condition.Status == corev1.ConditionTrue {
			score *= 0.8
		}
	}
	
	return max(score, 0.0)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}