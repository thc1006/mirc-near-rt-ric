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
	"time"
)

// autoScalingLoop continuously monitors and adjusts resource allocation
func (drm *DynamicResourceManager) autoScalingLoop(ctx context.Context) {
	ticker := time.NewTicker(drm.autoScalingConfig.ScaleUpThreshold / 2) // Check more frequently than threshold
	defer ticker.Stop()
	
	drm.logger.Info("Auto-scaling loop started",
		slog.Float64("cpu_target", drm.autoScalingConfig.CPUTargetUtilization),
		slog.Float64("memory_target", drm.autoScalingConfig.MemoryTargetUtilization))
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := drm.evaluateAndScale(ctx); err != nil {
				drm.logger.Error("Auto-scaling evaluation failed", 
					slog.String("error", err.Error()))
			}
		}
	}
}

// evaluateAndScale evaluates current conditions and makes scaling decisions
func (drm *DynamicResourceManager) evaluateAndScale(ctx context.Context) error {
	// Get current utilization
	utilization, err := drm.GetClusterUtilization(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cluster utilization: %w", err)
	}
	
	// Determine if scaling action is needed
	scaleDecision := drm.analyzeScalingNeeds(utilization)
	
	// Check cooldown periods
	if !drm.isScalingAllowed(scaleDecision.Action) {
		drm.logger.Debug("Scaling action suppressed due to cooldown",
			slog.String("action", string(scaleDecision.Action)),
			slog.Duration("time_since_last", time.Since(drm.lastScaleAction)))
		return nil
	}
	
	// Execute scaling action if needed
	if scaleDecision.Action != ScaleActionHold {
		if err := drm.executeScalingAction(ctx, scaleDecision); err != nil {
			return fmt.Errorf("failed to execute scaling action: %w", err)
		}
	}
	
	return nil
}

// ScalingDecision represents a scaling decision with context
type ScalingDecision struct {
	Action                ScaleActionType `json:"action"`
	CurrentConcurrentJobs int            `json:"current_concurrent_jobs"`
	TargetConcurrentJobs  int            `json:"target_concurrent_jobs"`
	Trigger               string         `json:"trigger"`
	Confidence           float64        `json:"confidence"`
	Urgency              float64        `json:"urgency"`
	Reasoning            []string       `json:"reasoning"`
}

// analyzeScalingNeeds analyzes current conditions and determines scaling requirements
func (drm *DynamicResourceManager) analyzeScalingNeeds(utilization *ClusterUtilization) *ScalingDecision {
	decision := &ScalingDecision{
		Action:                ScaleActionHold,
		CurrentConcurrentJobs: utilization.ActiveJobCount,
		TargetConcurrentJobs:  utilization.ActiveJobCount,
		Confidence:           0.5,
		Urgency:              0.0,
		Reasoning:            make([]string, 0),
	}
	
	// Analyze CPU utilization
	cpuPressure := drm.analyzeCPUPressure(utilization, decision)
	
	// Analyze memory utilization
	memoryPressure := drm.analyzeMemoryPressure(utilization, decision)
	
	// Analyze queue length
	queuePressure := drm.analyzeQueuePressure(utilization, decision)
	
	// Analyze predictions if available
	predictionPressure := drm.analyzePredictionPressure(utilization, decision)
	
	// Combine all pressure signals
	overallPressure := drm.combinePressureSignals(cpuPressure, memoryPressure, queuePressure, predictionPressure)
	
	// Determine scaling action based on overall pressure
	decision = drm.determineScalingAction(decision, overallPressure, utilization)
	
	drm.logger.Debug("Scaling analysis completed",
		slog.String("action", string(decision.Action)),
		slog.Float64("cpu_pressure", cpuPressure),
		slog.Float64("memory_pressure", memoryPressure),
		slog.Float64("queue_pressure", queuePressure),
		slog.Float64("overall_pressure", overallPressure),
		slog.Float64("confidence", decision.Confidence))
	
	return decision
}

// analyzeCPUPressure evaluates CPU utilization pressure
func (drm *DynamicResourceManager) analyzeCPUPressure(utilization *ClusterUtilization, decision *ScalingDecision) float64 {
	target := drm.autoScalingConfig.CPUTargetUtilization
	current := utilization.OverallCPUUtilization
	
	// Calculate pressure as deviation from target
	pressure := (current - target) / target
	
	if current > target*1.2 { // 20% above target
		decision.Reasoning = append(decision.Reasoning, 
			fmt.Sprintf("CPU utilization (%.1f%%) significantly exceeds target (%.1f%%)", 
				current*100, target*100))
	} else if current < target*0.7 { // 30% below target
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("CPU utilization (%.1f%%) significantly below target (%.1f%%)", 
				current*100, target*100))
	}
	
	return pressure
}

// analyzeMemoryPressure evaluates memory utilization pressure
func (drm *DynamicResourceManager) analyzeMemoryPressure(utilization *ClusterUtilization, decision *ScalingDecision) float64 {
	target := drm.autoScalingConfig.MemoryTargetUtilization
	current := utilization.OverallMemoryUtilization
	
	// Calculate pressure as deviation from target
	pressure := (current - target) / target
	
	if current > target*1.15 { // 15% above target (memory is more critical)
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Memory utilization (%.1f%%) significantly exceeds target (%.1f%%)", 
				current*100, target*100))
	} else if current < target*0.6 { // 40% below target
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Memory utilization (%.1f%%) significantly below target (%.1f%%)", 
				current*100, target*100))
	}
	
	return pressure
}

// analyzeQueuePressure evaluates job queue pressure
func (drm *DynamicResourceManager) analyzeQueuePressure(utilization *ClusterUtilization, decision *ScalingDecision) float64 {
	queueLength := utilization.QueuedJobCount
	activeJobs := utilization.ActiveJobCount
	
	// Calculate queue pressure based on queue length relative to active jobs
	var pressure float64
	if activeJobs > 0 {
		queueRatio := float64(queueLength) / float64(activeJobs)
		pressure = queueRatio - 0.3 // Target 30% queue to active ratio
		
		if queueRatio > 0.5 {
			decision.Reasoning = append(decision.Reasoning,
				fmt.Sprintf("High queue pressure: %d jobs queued vs %d active", 
					queueLength, activeJobs))
		}
	} else if queueLength > 0 {
		pressure = float64(queueLength) / 5.0 // Normalize by expected baseline
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Jobs in queue (%d) with no active jobs", queueLength))
	}
	
	return pressure
}

// analyzePredictionPressure evaluates ML prediction pressure
func (drm *DynamicResourceManager) analyzePredictionPressure(utilization *ClusterUtilization, decision *ScalingDecision) float64 {
	if utilization.PredictedUtilization == nil {
		return 0.0
	}
	
	pred := utilization.PredictedUtilization
	cpuTarget := drm.autoScalingConfig.CPUTargetUtilization
	memoryTarget := drm.autoScalingConfig.MemoryTargetUtilization
	
	// Calculate predicted pressure
	cpuPredPressure := (pred.NextPeriodCPU - cpuTarget) / cpuTarget
	memoryPredPressure := (pred.NextPeriodMemory - memoryTarget) / memoryTarget
	
	pressure := (cpuPredPressure + memoryPredPressure) / 2.0
	
	// Weight by prediction confidence
	pressure *= pred.Confidence
	
	if pressure > 0.2 {
		decision.Reasoning = append(decision.Reasoning,
			fmt.Sprintf("Predicted high utilization: CPU %.1f%%, Memory %.1f%% (confidence %.1f%%)",
				pred.NextPeriodCPU*100, pred.NextPeriodMemory*100, pred.Confidence*100))
	}
	
	return pressure
}

// combinePressureSignals combines multiple pressure signals into overall pressure
func (drm *DynamicResourceManager) combinePressureSignals(cpu, memory, queue, prediction float64) float64 {
	// Weight different pressure signals
	weights := struct {
		cpu        float64
		memory     float64
		queue      float64
		prediction float64
	}{
		cpu:        0.35,
		memory:     0.35, // Memory is critical
		queue:      0.20,
		prediction: 0.10,
	}
	
	overall := weights.cpu*cpu + 
		weights.memory*memory + 
		weights.queue*queue + 
		weights.prediction*prediction
	
	return overall
}

// determineScalingAction determines the appropriate scaling action
func (drm *DynamicResourceManager) determineScalingAction(
	decision *ScalingDecision, 
	pressure float64, 
	utilization *ClusterUtilization,
) *ScalingDecision {
	
	currentJobs := utilization.ActiveJobCount
	maxJobs := drm.autoScalingConfig.MaxConcurrentJobsLimit
	minJobs := drm.autoScalingConfig.MinConcurrentJobs
	
	// Determine action based on pressure
	if pressure > 0.3 { // High pressure - scale up
		decision.Action = ScaleActionUp
		decision.Urgency = min(pressure, 1.0)
		decision.Confidence = 0.8
		
		// Calculate target jobs
		scaleUpFactor := 1.0 + (pressure * drm.autoScalingConfig.AdaptationAggression)
		targetJobs := int(math.Ceil(float64(currentJobs) * scaleUpFactor))
		decision.TargetConcurrentJobs = min(targetJobs, maxJobs)
		
		decision.Trigger = "high_resource_pressure"
		
	} else if pressure < -0.2 { // Low pressure - scale down
		decision.Action = ScaleActionDown
		decision.Urgency = min(math.Abs(pressure), 1.0)
		decision.Confidence = 0.7
		
		// Calculate target jobs
		scaleDownFactor := 1.0 + (pressure * drm.autoScalingConfig.AdaptationAggression)
		targetJobs := int(math.Floor(float64(currentJobs) * scaleDownFactor))
		decision.TargetConcurrentJobs = max(targetJobs, minJobs)
		
		decision.Trigger = "low_resource_pressure"
		
	} else {
		// Pressure within acceptable range
		decision.Action = ScaleActionHold
		decision.Trigger = "stable_resource_utilization"
		decision.Confidence = 0.9
	}
	
	// Emergency scaling for critical situations
	if pressure > 0.8 || utilization.OverallMemoryUtilization > 0.95 {
		decision.Action = ScaleActionEmergency
		decision.Urgency = 1.0
		decision.Confidence = 1.0
		decision.Trigger = "emergency_resource_pressure"
		decision.TargetConcurrentJobs = min(currentJobs/2, minJobs) // Aggressive scale down
	}
	
	return decision
}

// isScalingAllowed checks if scaling action is allowed based on cooldown periods
func (drm *DynamicResourceManager) isScalingAllowed(action ScaleActionType) bool {
	timeSinceLastScale := time.Since(drm.lastScaleAction)
	
	switch action {
	case ScaleActionUp:
		return timeSinceLastScale >= drm.autoScalingConfig.ScaleUpCooldown
	case ScaleActionDown:
		return timeSinceLastScale >= drm.autoScalingConfig.ScaleDownCooldown
	case ScaleActionEmergency:
		return true // Emergency actions override cooldowns
	default:
		return true
	}
}

// executeScalingAction performs the actual scaling operation
func (drm *DynamicResourceManager) executeScalingAction(ctx context.Context, decision *ScalingDecision) error {
	drm.logger.Info("Executing scaling action",
		slog.String("action", string(decision.Action)),
		slog.Int("from_jobs", decision.CurrentConcurrentJobs),
		slog.Int("to_jobs", decision.TargetConcurrentJobs),
		slog.String("trigger", decision.Trigger),
		slog.Float64("confidence", decision.Confidence))
	
	switch decision.Action {
	case ScaleActionUp:
		return drm.scaleUp(ctx, decision)
	case ScaleActionDown:
		return drm.scaleDown(ctx, decision)
	case ScaleActionEmergency:
		return drm.emergencyScale(ctx, decision)
	default:
		return fmt.Errorf("unknown scaling action: %s", decision.Action)
	}
}

// scaleUp increases the number of concurrent jobs
func (drm *DynamicResourceManager) scaleUp(ctx context.Context, decision *ScalingDecision) error {
	targetJobs := decision.TargetConcurrentJobs
	currentJobs := decision.CurrentConcurrentJobs
	
	if targetJobs <= currentJobs {
		return nil // Nothing to scale up
	}
	
	// Start additional jobs from the queue
	jobsToStart := targetJobs - currentJobs
	startedJobs := 0
	
	drm.mutex.Lock()
	defer drm.mutex.Unlock()
	
	for i := 0; i < jobsToStart && drm.jobSchedulingQueue.Size() > 0; i++ {
		job := drm.jobSchedulingQueue.Pop()
		if job != nil {
			if err := drm.startJob(ctx, job); err != nil {
				drm.logger.Error("Failed to start job during scale up",
					slog.String("job_id", job.JobID),
					slog.String("error", err.Error()))
				// Re-queue the job
				drm.jobSchedulingQueue.Push(job)
				continue
			}
			startedJobs++
		}
	}
	
	// Record the scaling action
	drm.recordScalingAction(decision, startedJobs)
	
	drm.logger.Info("Scale up completed",
		slog.Int("jobs_started", startedJobs),
		slog.Int("requested", jobsToStart))
	
	return nil
}

// scaleDown decreases the number of concurrent jobs
func (drm *DynamicResourceManager) scaleDown(ctx context.Context, decision *ScalingDecision) error {
	targetJobs := decision.TargetConcurrentJobs
	currentJobs := decision.CurrentConcurrentJobs
	
	if targetJobs >= currentJobs {
		return nil // Nothing to scale down
	}
	
	// Stop some running jobs gracefully
	jobsToStop := currentJobs - targetJobs
	stoppedJobs := 0
	
	drm.mutex.Lock()
	defer drm.mutex.Unlock()
	
	// Select jobs to stop (prefer lower priority jobs)
	jobsToStopList := drm.selectJobsForTermination(jobsToStop)
	
	for _, job := range jobsToStopList {
		if err := drm.stopJob(ctx, job, "scaling_down"); err != nil {
			drm.logger.Error("Failed to stop job during scale down",
				slog.String("job_id", job.JobID),
				slog.String("error", err.Error()))
			continue
		}
		stoppedJobs++
	}
	
	// Record the scaling action
	drm.recordScalingAction(decision, -stoppedJobs)
	
	drm.logger.Info("Scale down completed",
		slog.Int("jobs_stopped", stoppedJobs),
		slog.Int("requested", jobsToStop))
	
	return nil
}

// emergencyScale performs emergency scaling for critical situations
func (drm *DynamicResourceManager) emergencyScale(ctx context.Context, decision *ScalingDecision) error {
	drm.logger.Warn("Executing emergency scaling",
		slog.String("trigger", decision.Trigger),
		slog.Int("target_jobs", decision.TargetConcurrentJobs))
	
	// Emergency scale down - stop jobs immediately
	return drm.scaleDown(ctx, decision)
}

// recordScalingAction records a scaling action for historical analysis
func (drm *DynamicResourceManager) recordScalingAction(decision *ScalingDecision, actualChange int) {
	record := ScaleActionRecord{
		Timestamp:         time.Now(),
		Action:           decision.Action,
		FromValue:        decision.CurrentConcurrentJobs,
		ToValue:          decision.CurrentConcurrentJobs + actualChange,
		Trigger:          decision.Trigger,
		ActiveJobs:       decision.CurrentConcurrentJobs,
		QueuedJobs:       0, // Would be filled from current state
	}
	
	if drm.currentUtilization != nil {
		record.CPUUtilization = drm.currentUtilization.OverallCPUUtilization
		record.MemoryUtilization = drm.currentUtilization.OverallMemoryUtilization
		record.QueuedJobs = drm.currentUtilization.QueuedJobCount
	}
	
	drm.scaleActionHistory = append(drm.scaleActionHistory, record)
	drm.lastScaleAction = time.Now()
	
	// Keep only recent history (last 100 actions)
	if len(drm.scaleActionHistory) > 100 {
		drm.scaleActionHistory = drm.scaleActionHistory[1:]
	}
}

// Helper functions that would be implemented elsewhere
func (drm *DynamicResourceManager) startJob(ctx context.Context, job *QueuedJob) error {
	// Implementation for starting a job
	return nil
}

func (drm *DynamicResourceManager) stopJob(ctx context.Context, job *ScheduledJob, reason string) error {
	// Implementation for stopping a job
	return nil
}

func (drm *DynamicResourceManager) selectJobsForTermination(count int) []*ScheduledJob {
	// Implementation for selecting jobs to terminate
	jobs := make([]*ScheduledJob, 0)
	// Would select lowest priority jobs first
	return jobs
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}