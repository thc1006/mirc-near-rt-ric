package a1

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// InMemoryA1Repository provides an in-memory implementation of A1Repository
// In production, this would be replaced with a database-backed implementation
type InMemoryA1Repository struct {
	policyTypes   map[string]*A1PolicyType
	policies      map[string]*A1Policy
	enforcements  map[string]*PolicyEnforcement
	metrics       map[string]*PolicyMetrics
	mutex         sync.RWMutex
	logger        *logrus.Logger
}

// NewInMemoryA1Repository creates a new in-memory A1 repository
func NewInMemoryA1Repository() A1Repository {
	return &InMemoryA1Repository{
		policyTypes:  make(map[string]*A1PolicyType),
		policies:     make(map[string]*A1Policy),
		enforcements: make(map[string]*PolicyEnforcement),
		metrics:      make(map[string]*PolicyMetrics),
		logger:       logrus.WithField("component", "a1-repository").Logger,
	}
}

// Policy Type operations

func (repo *InMemoryA1Repository) CreatePolicyType(policyType *A1PolicyType) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.policyTypes[policyType.PolicyTypeID]; exists {
		return fmt.Errorf("policy type %s already exists", policyType.PolicyTypeID)
	}

	// Deep copy to prevent external modifications
	policyTypeCopy := *policyType
	policyTypeCopy.CreatedAt = time.Now()
	policyTypeCopy.LastModified = time.Now()

	repo.policyTypes[policyType.PolicyTypeID] = &policyTypeCopy

	repo.logger.WithField("policy_type_id", policyType.PolicyTypeID).Info("Policy type created")
	return nil
}

func (repo *InMemoryA1Repository) GetPolicyType(policyTypeID string) (*A1PolicyType, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	policyType, exists := repo.policyTypes[policyTypeID]
	if !exists {
		return nil, fmt.Errorf("policy type %s not found", policyTypeID)
	}

	// Return a copy to prevent external modifications
	policyTypeCopy := *policyType
	return &policyTypeCopy, nil
}

func (repo *InMemoryA1Repository) UpdatePolicyType(policyType *A1PolicyType) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.policyTypes[policyType.PolicyTypeID]; !exists {
		return fmt.Errorf("policy type %s not found", policyType.PolicyTypeID)
	}

	policyTypeCopy := *policyType
	policyTypeCopy.LastModified = time.Now()

	repo.policyTypes[policyType.PolicyTypeID] = &policyTypeCopy

	repo.logger.WithField("policy_type_id", policyType.PolicyTypeID).Info("Policy type updated")
	return nil
}

func (repo *InMemoryA1Repository) DeletePolicyType(policyTypeID string) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.policyTypes[policyTypeID]; !exists {
		return fmt.Errorf("policy type %s not found", policyTypeID)
	}

	delete(repo.policyTypes, policyTypeID)

	repo.logger.WithField("policy_type_id", policyTypeID).Info("Policy type deleted")
	return nil
}

func (repo *InMemoryA1Repository) ListPolicyTypes() ([]*A1PolicyType, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	policyTypes := make([]*A1PolicyType, 0, len(repo.policyTypes))
	for _, policyType := range repo.policyTypes {
		policyTypeCopy := *policyType
		policyTypes = append(policyTypes, &policyTypeCopy)
	}

	return policyTypes, nil
}

// Policy operations

func (repo *InMemoryA1Repository) CreatePolicy(policy *A1Policy) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.policies[policy.PolicyID]; exists {
		return fmt.Errorf("policy %s already exists", policy.PolicyID)
	}

	// Verify policy type exists
	if _, exists := repo.policyTypes[policy.PolicyTypeID]; !exists {
		return fmt.Errorf("policy type %s not found", policy.PolicyTypeID)
	}

	// Deep copy to prevent external modifications
	policyCopy := *policy
	policyCopy.CreatedAt = time.Now()
	policyCopy.LastModified = time.Now()
	policyCopy.Status = A1PolicyStatusNotEnforced

	repo.policies[policy.PolicyID] = &policyCopy

	// Initialize metrics
	repo.metrics[policy.PolicyID] = &PolicyMetrics{
		EnforcementLatency:   0,
		SuccessRate:          0,
		ErrorRate:            0,
		LastEnforcementTime:  time.Time{},
		TotalEnforcements:    0,
		FailedEnforcements:   0,
	}

	repo.logger.WithFields(logrus.Fields{
		"policy_id":      policy.PolicyID,
		"policy_type_id": policy.PolicyTypeID,
	}).Info("Policy created")

	return nil
}

func (repo *InMemoryA1Repository) GetPolicy(policyID string) (*A1Policy, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	policy, exists := repo.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("policy %s not found", policyID)
	}

	// Return a copy to prevent external modifications
	policyCopy := *policy
	return &policyCopy, nil
}

func (repo *InMemoryA1Repository) UpdatePolicy(policy *A1Policy) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	existingPolicy, exists := repo.policies[policy.PolicyID]
	if !exists {
		return fmt.Errorf("policy %s not found", policy.PolicyID)
	}

	// Preserve creation time and update modification time
	policyCopy := *policy
	policyCopy.CreatedAt = existingPolicy.CreatedAt
	policyCopy.LastModified = time.Now()

	repo.policies[policy.PolicyID] = &policyCopy

	repo.logger.WithField("policy_id", policy.PolicyID).Info("Policy updated")
	return nil
}

func (repo *InMemoryA1Repository) DeletePolicy(policyID string) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, exists := repo.policies[policyID]; !exists {
		return fmt.Errorf("policy %s not found", policyID)
	}

	delete(repo.policies, policyID)
	delete(repo.enforcements, policyID)
	delete(repo.metrics, policyID)

	repo.logger.WithField("policy_id", policyID).Info("Policy deleted")
	return nil
}

func (repo *InMemoryA1Repository) ListPolicies() ([]*A1Policy, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	policies := make([]*A1Policy, 0, len(repo.policies))
	for _, policy := range repo.policies {
		policyCopy := *policy
		policies = append(policies, &policyCopy)
	}

	return policies, nil
}

func (repo *InMemoryA1Repository) ListPoliciesByType(policyTypeID string) ([]*A1Policy, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	var policies []*A1Policy
	for _, policy := range repo.policies {
		if policy.PolicyTypeID == policyTypeID {
			policyCopy := *policy
			policies = append(policies, &policyCopy)
		}
	}

	return policies, nil
}

// Policy enforcement tracking

func (repo *InMemoryA1Repository) RecordEnforcement(enforcement *PolicyEnforcement) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	// Verify policy exists
	policy, exists := repo.policies[enforcement.PolicyID]
	if !exists {
		return fmt.Errorf("policy %s not found", enforcement.PolicyID)
	}

	// Update policy status
	policy.Status = enforcement.Status
	policy.LastModified = time.Now()

	// Store enforcement record
	enforcementCopy := *enforcement
	enforcementCopy.EnforcementTime = time.Now()
	repo.enforcements[enforcement.PolicyID] = &enforcementCopy

	// Update metrics
	metrics, exists := repo.metrics[enforcement.PolicyID]
	if !exists {
		metrics = &PolicyMetrics{}
		repo.metrics[enforcement.PolicyID] = metrics
	}

	metrics.TotalEnforcements++
	metrics.LastEnforcementTime = time.Now()

	if enforcement.Status == A1PolicyStatusEnforced {
		metrics.SuccessRate = float64(metrics.TotalEnforcements-metrics.FailedEnforcements) / float64(metrics.TotalEnforcements)
	} else if enforcement.Status == A1PolicyStatusError {
		metrics.FailedEnforcements++
		metrics.SuccessRate = float64(metrics.TotalEnforcements-metrics.FailedEnforcements) / float64(metrics.TotalEnforcements)
	}

	metrics.ErrorRate = float64(metrics.FailedEnforcements) / float64(metrics.TotalEnforcements)

	if enforcement.Metrics != nil {
		metrics.EnforcementLatency = enforcement.Metrics.EnforcementLatency
	}

	repo.logger.WithFields(logrus.Fields{
		"policy_id": enforcement.PolicyID,
		"status":    enforcement.Status,
	}).Info("Policy enforcement recorded")

	return nil
}

func (repo *InMemoryA1Repository) GetEnforcementStatus(policyID string) (*PolicyEnforcement, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	enforcement, exists := repo.enforcements[policyID]
	if !exists {
		return nil, fmt.Errorf("no enforcement record for policy %s", policyID)
	}

	// Return a copy
	enforcementCopy := *enforcement
	return &enforcementCopy, nil
}

func (repo *InMemoryA1Repository) GetPolicyMetrics(policyID string) (*PolicyMetrics, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	metrics, exists := repo.metrics[policyID]
	if !exists {
		return nil, fmt.Errorf("no metrics for policy %s", policyID)
	}

	// Return a copy
	metricsCopy := *metrics
	return &metricsCopy, nil
}

// Additional utility methods

func (repo *InMemoryA1Repository) GetPolicyTypeStatus(policyTypeID string) (*A1PolicyTypeStatus, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	// Verify policy type exists
	if _, exists := repo.policyTypes[policyTypeID]; !exists {
		return nil, fmt.Errorf("policy type %s not found", policyTypeID)
	}

	// Count policies of this type
	var policyInstanceIDs []string
	for _, policy := range repo.policies {
		if policy.PolicyTypeID == policyTypeID {
			policyInstanceIDs = append(policyInstanceIDs, policy.PolicyID)
		}
	}

	return &A1PolicyTypeStatus{
		PolicyTypeID:      policyTypeID,
		NumberOfPolicies:  len(policyInstanceIDs),
		PolicyInstanceIDs: policyInstanceIDs,
	}, nil
}

func (repo *InMemoryA1Repository) GetStatistics() map[string]interface{} {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	// Count policies by status
	statusCounts := make(map[A1PolicyStatus]int)
	for _, policy := range repo.policies {
		statusCounts[policy.Status]++
	}

	// Count policies by type
	typeCounts := make(map[string]int)
	for _, policy := range repo.policies {
		typeCounts[policy.PolicyTypeID]++
	}

	// Calculate average metrics
	var totalLatency time.Duration
	var totalSuccessRate, totalErrorRate float64
	metricsCount := 0

	for _, metrics := range repo.metrics {
		if metrics.TotalEnforcements > 0 {
			totalLatency += metrics.EnforcementLatency
			totalSuccessRate += metrics.SuccessRate
			totalErrorRate += metrics.ErrorRate
			metricsCount++
		}
	}

	avgLatency := time.Duration(0)
	avgSuccessRate := float64(0)
	avgErrorRate := float64(0)

	if metricsCount > 0 {
		avgLatency = totalLatency / time.Duration(metricsCount)
		avgSuccessRate = totalSuccessRate / float64(metricsCount)
		avgErrorRate = totalErrorRate / float64(metricsCount)
	}

	return map[string]interface{}{
		"total_policy_types":     len(repo.policyTypes),
		"total_policies":         len(repo.policies),
		"policies_by_status":     statusCounts,
		"policies_by_type":       typeCounts,
		"average_latency_ms":     avgLatency.Milliseconds(),
		"average_success_rate":   avgSuccessRate,
		"average_error_rate":     avgErrorRate,
		"total_enforcements":     func() int64 {
			var total int64
			for _, metrics := range repo.metrics {
				total += metrics.TotalEnforcements
			}
			return total
		}(),
	}
}

// Export/Import functions for backup and restore

func (repo *InMemoryA1Repository) Export() ([]byte, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	exportData := struct {
		PolicyTypes  map[string]*A1PolicyType    `json:"policy_types"`
		Policies     map[string]*A1Policy        `json:"policies"`
		Enforcements map[string]*PolicyEnforcement `json:"enforcements"`
		Metrics      map[string]*PolicyMetrics   `json:"metrics"`
		ExportTime   time.Time                   `json:"export_time"`
	}{
		PolicyTypes:  repo.policyTypes,
		Policies:     repo.policies,
		Enforcements: repo.enforcements,
		Metrics:      repo.metrics,
		ExportTime:   time.Now(),
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal export data: %w", err)
	}

	return data, nil
}

func (repo *InMemoryA1Repository) Import(data []byte) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	var importData struct {
		PolicyTypes  map[string]*A1PolicyType    `json:"policy_types"`
		Policies     map[string]*A1Policy        `json:"policies"`
		Enforcements map[string]*PolicyEnforcement `json:"enforcements"`
		Metrics      map[string]*PolicyMetrics   `json:"metrics"`
		ExportTime   time.Time                   `json:"export_time"`
	}

	if err := json.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("failed to unmarshal import data: %w", err)
	}

	// Clear existing data
	repo.policyTypes = make(map[string]*A1PolicyType)
	repo.policies = make(map[string]*A1Policy)
	repo.enforcements = make(map[string]*PolicyEnforcement)
	repo.metrics = make(map[string]*PolicyMetrics)

	// Import data
	repo.policyTypes = importData.PolicyTypes
	repo.policies = importData.Policies
	repo.enforcements = importData.Enforcements
	repo.metrics = importData.Metrics

	repo.logger.WithFields(logrus.Fields{
		"policy_types": len(repo.policyTypes),
		"policies":     len(repo.policies),
		"export_time":  importData.ExportTime,
	}).Info("Data imported successfully")

	return nil
}

// Cleanup removes old enforcement records and metrics
func (repo *InMemoryA1Repository) Cleanup(maxAge time.Duration) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	cleaned := 0

	// Clean old enforcement records
	for policyID, enforcement := range repo.enforcements {
		if enforcement.EnforcementTime.Before(cutoff) {
			// Only clean if the policy no longer exists
			if _, exists := repo.policies[policyID]; !exists {
				delete(repo.enforcements, policyID)
				delete(repo.metrics, policyID)
				cleaned++
			}
		}
	}

	if cleaned > 0 {
		repo.logger.WithField("cleaned_records", cleaned).Info("Cleanup completed")
	}

	return nil
}