package a1

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hctsai1006/near-rt-ric/internal/config"
	"github.com/hctsai1006/near-rt-ric/pkg/common/monitoring"
	"github.com/sirupsen/logrus"
)

type PolicyEventHandler interface {
	OnPolicyTypeCreated(policyType *PolicyType)
	OnPolicyTypeDeleted(policyTypeID PolicyTypeID)
	OnPolicyInstanceCreated(policy *PolicyInstance)
	OnPolicyInstanceDeleted(policyID PolicyID)
}

// PolicyManager manages A1 policies
type PolicyManager struct {
	log         *logrus.Logger
	config      *config.A1Config
	metrics     *monitoring.MetricsCollector
	db          *sql.DB
	mu          sync.RWMutex
	eventHandler PolicyEventHandler
}

// NewPolicyManager creates a new policy manager
func NewPolicyManager(config *config.A1Config, log *logrus.Logger, metrics *monitoring.MetricsCollector, db *sql.DB) *PolicyManager {
	return &PolicyManager{
		config:      config,
		log:         log,
		metrics:     metrics,
		db:          db,
	}
}

// CreatePolicyType creates a new policy type
func (pm *PolicyManager) CreatePolicyType(req *PolicyTypeCreateRequest) (*PolicyType, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if policy type already exists
	var count int
	err := pm.db.QueryRow("SELECT COUNT(*) FROM policy_types WHERE policy_type_id = $1", req.PolicyTypeID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing policy type: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("policy type already exists")
	}

	policyType := &PolicyType{
		PolicyTypeID: req.PolicyTypeID,
		Name:         req.Name,
		Description:  req.Description,
		PolicySchema: req.PolicySchema,
		CreateSchema: req.CreateSchema,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Insert into database
	_, err = pm.db.Exec(
		"INSERT INTO policy_types (policy_type_id, name, description, policy_schema, create_schema, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		policyType.PolicyTypeID, policyType.Name, policyType.Description, policyType.PolicySchema, policyType.CreateSchema, policyType.CreatedAt, policyType.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert policy type: %w", err)
	}

	pm.log.Infof("Created policy type %s", req.PolicyTypeID)

	if pm.eventHandler != nil {
		pm.eventHandler.OnPolicyTypeCreated(policyType)
	}

	return policyType, nil
}

// GetPolicyType returns a specific policy type
func (pm *PolicyManager) GetPolicyType(policyTypeID PolicyTypeID) (*PolicyType, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	policyType := &PolicyType{}
	row := pm.db.QueryRow("SELECT policy_type_id, name, description, policy_schema, create_schema, created_at, updated_at FROM policy_types WHERE policy_type_id = $1", policyTypeID)
	err := row.Scan(&policyType.PolicyTypeID, &policyType.Name, &policyType.Description, &policyType.PolicySchema, &policyType.CreateSchema, &policyType.CreatedAt, &policyType.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("policy type not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get policy type: %w", err)
	}

	return policyType, nil
}

// GetAllPolicyTypes returns all policy types
func (pm *PolicyManager) GetAllPolicyTypes() ([]*PolicyType, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	rows, err := pm.db.Query("SELECT policy_type_id, name, description, policy_schema, create_schema, created_at, updated_at FROM policy_types")
	if err != nil {
		return nil, fmt.Errorf("failed to query policy types: %w", err)
	}
	defer rows.Close()

	policyTypes := make([]*PolicyType, 0)
	for rows.Next() {
		policyType := &PolicyType{}
		if err := rows.Scan(&policyType.PolicyTypeID, &policyType.Name, &policyType.Description, &policyType.PolicySchema, &policyType.CreateSchema, &policyType.CreatedAt, &policyType.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan policy type: %w", err)
		}
		policyTypes = append(policyTypes, policyType)
	}

	return policyTypes, nil
}

// DeletePolicyType deletes a policy type
func (pm *PolicyManager) DeletePolicyType(policyTypeID PolicyTypeID) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if policy type exists
	var count int
	err := pm.db.QueryRow("SELECT COUNT(*) FROM policy_types WHERE policy_type_id = $1", policyTypeID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing policy type: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("policy type not found")
	}

	// Check for active policy instances
	err = pm.db.QueryRow("SELECT COUNT(*) FROM policy_instances WHERE policy_type_id = $1", policyTypeID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check active policy instances: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("active policy instances exist for this policy type")
	}

	// Delete from database
	_, err = pm.db.Exec("DELETE FROM policy_types WHERE policy_type_id = $1", policyTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete policy type: %w", err)
	}

	pm.log.Infof("Deleted policy type %s", policyTypeID)

	if pm.eventHandler != nil {
		pm.eventHandler.OnPolicyTypeDeleted(policyTypeID)
	}

	return nil
}

// GetPolicyInstancesByType returns all policy instances for a policy type
func (pm *PolicyManager) GetPolicyInstancesByType(policyTypeID PolicyTypeID) ([]*PolicyInstance, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	rows, err := pm.db.Query("SELECT policy_id, policy_type_id, policy_data, status, status_reason, target_near_rt_ric, created_by, created_at, last_updated_by, last_updated_at, enforced_at FROM policy_instances WHERE policy_type_id = $1", policyTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query policy instances by type: %w", err)
	}
	defer rows.Close()

	instances := make([]*PolicyInstance, 0)
	for rows.Next() {
		instance := &PolicyInstance{}
		if err := rows.Scan(&instance.PolicyID, &instance.PolicyTypeID, &instance.PolicyData, &instance.Status, &instance.StatusReason, &instance.TargetNearRTRIC, &instance.CreatedBy, &instance.CreatedAt, &instance.LastUpdatedBy, &instance.LastUpdatedAt, &instance.EnforcedAt); err != nil {
			return nil, fmt.Errorf("failed to scan policy instance: %w", err)
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// GetPolicyStatus returns the status of a policy instance
func (pm *PolicyManager) GetPolicyStatus(policyID PolicyID) (*PolicyStatusInfo, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	statusInfo := &PolicyStatusInfo{}
	row := pm.db.QueryRow("SELECT policy_id, policy_type_id, status, status_reason, last_updated_at FROM policy_instances WHERE policy_id = $1", policyID)
	err := row.Scan(&statusInfo.PolicyID, &statusInfo.PolicyTypeID, &statusInfo.Status, &statusInfo.StatusReason, &statusInfo.LastUpdate)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("policy instance not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get policy status: %w", err)
	}

	return statusInfo, nil
}

// GetStatistics returns statistics about the policy manager
func (pm *PolicyManager) GetStatistics() (*A1Statistics, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := &A1Statistics{}

	// Total policy types
	err := pm.db.QueryRow("SELECT COUNT(*) FROM policy_types").Scan(&stats.TotalPolicyTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to get total policy types: %w", err)
	}

	// Total policy instances
	err = pm.db.QueryRow("SELECT COUNT(*) FROM policy_instances").Scan(&stats.TotalPolicyInstances)
	if err != nil {
		return nil, fmt.Errorf("failed to get total policy instances: %w", err)
	}

	// Enforced policies
	err = pm.db.QueryRow("SELECT COUNT(*) FROM policy_instances WHERE status = $1", PolicyStatusEnforced).Scan(&stats.EnforcedPolicies)
	if err != nil {
		return nil, fmt.Errorf("failed to get enforced policies: %w", err)
	}

	// Failed policies (assuming a 'FAILED' status exists or can be derived)
	// For now, we'll just set it to 0 or implement proper status tracking.
	stats.FailedPolicies = 0 

	// Placeholder for request metrics (these would typically come from a metrics collector, not DB)
	stats.TotalRequests = 0
	stats.SuccessfulRequests = 0
	stats.FailedRequests = 0
	stats.AverageResponseTime = 0
	stats.Uptime = 0
	stats.LastUpdate = time.Now()

	return stats, nil
}

// CreatePolicyInstance creates a new policy instance
func (pm *PolicyManager) CreatePolicyInstance(policyTypeID PolicyTypeID, policyID PolicyID, req *PolicyInstanceCreateRequest, userID string) (*PolicyInstance, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if policy instance already exists
	var count int
	err := pm.db.QueryRow("SELECT COUNT(*) FROM policy_instances WHERE policy_id = $1", policyID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing policy instance: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("policy instance already exists")
	}

	// Check if policy type exists
	err = pm.db.QueryRow("SELECT COUNT(*) FROM policy_types WHERE policy_type_id = $1", policyTypeID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing policy type: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("policy type not found")
	}

	instance := &PolicyInstance{
		PolicyID:      policyID,
		PolicyTypeID:  policyTypeID,
		PolicyData:    req.PolicyData,
		Status:        PolicyStatusNotEnforced,
		CreatedBy:     userID,
		CreatedAt:     time.Now(),
		LastUpdatedBy: userID,
		LastUpdatedAt: time.Now(),
	}

	// Insert into database
	_, err = pm.db.Exec(
		"INSERT INTO policy_instances (policy_id, policy_type_id, policy_data, status, created_by, created_at, last_updated_by, last_updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		instance.PolicyID, instance.PolicyTypeID, instance.PolicyData, instance.Status, instance.CreatedBy, instance.CreatedAt, instance.LastUpdatedBy, instance.LastUpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert policy instance: %w", err)
	}

	pm.log.Infof("Created policy instance %s of type %s by user %s", policyID, policyTypeID, userID)

	if pm.eventHandler != nil {
		pm.eventHandler.OnPolicyInstanceCreated(instance)
	}

	return instance, nil
}

// DeletePolicyInstance deletes a policy instance
func (pm *PolicyManager) DeletePolicyInstance(policyID PolicyID, userID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if policy instance exists
	var count int
	err := pm.db.QueryRow("SELECT COUNT(*) FROM policy_instances WHERE policy_id = $1", policyID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing policy instance: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("policy instance not found")
	}

	// Delete from database
	_, err = pm.db.Exec("DELETE FROM policy_instances WHERE policy_id = $1", policyID)
	if err != nil {
		return fmt.Errorf("failed to delete policy instance: %w", err)
	}

	pm.log.Infof("Deleted policy instance %s by user %s", policyID, userID)

	if pm.eventHandler != nil {
		pm.eventHandler.OnPolicyInstanceDeleted(policyID)
	}

	return nil
}

// GetPolicyInstance retrieves a policy instance
func (pm *PolicyManager) GetPolicyInstance(policyID PolicyID) (*PolicyInstance, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	instance := &PolicyInstance{}
	row := pm.db.QueryRow("SELECT policy_id, policy_type_id, policy_data, status, status_reason, target_near_rt_ric, created_by, created_at, last_updated_by, last_updated_at, enforced_at FROM policy_instances WHERE policy_id = $1", policyID)
	err := row.Scan(&instance.PolicyID, &instance.PolicyTypeID, &instance.PolicyData, &instance.Status, &instance.StatusReason, &instance.TargetNearRTRIC, &instance.CreatedBy, &instance.CreatedAt, &instance.LastUpdatedBy, &instance.LastUpdatedAt, &instance.EnforcedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("policy instance not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get policy instance: %w", err)
	}

	return instance, nil
}
