package policy

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

// PostgresPolicyManager implements the PolicyManager interface using a PostgreSQL database
type PostgresPolicyManager struct {
	db     *sql.DB
	logger *logrus.Logger
}

// NewPostgresPolicyManager creates a new PostgresPolicyManager
func NewPostgresPolicyManager(db *sql.DB) *PostgresPolicyManager {
	return &PostgresPolicyManager{
		db:     db,
		logger: logrus.WithField("component", "postgres-policy-manager").Logger,
	}
}

// CreatePoliciesTable creates the policies table in the database if it doesn't exist
func (m *PostgresPolicyManager) CreatePoliciesTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS policies (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		version TEXT NOT NULL,
		content TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL
	);`
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create policies table: %w", err)
	}
	m.logger.Info("Policies table created or already exists")
	return nil
}

// CreatePolicy adds a new policy to the database
func (m *PostgresPolicyManager) CreatePolicy(ctx context.Context, policy *Policy) error {
	query := `
	INSERT INTO policies (id, name, version, content, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7);`
	_, err := m.db.ExecContext(ctx, query, policy.ID, policy.Name, policy.Version, policy.Content, policy.Status, policy.CreatedAt, policy.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create policy: %w", err)
	}
	m.logger.Infof("Successfully created policy with ID: %s", policy.ID)
	return nil
}

// GetPolicy retrieves a policy from the database by its ID
func (m *PostgresPolicyManager) GetPolicy(ctx context.Context, id string) (*Policy, error) {
	query := `SELECT id, name, version, content, status, created_at, updated_at FROM policies WHERE id = $1;`
	row := m.db.QueryRowContext(ctx, query, id)
	policy := &Policy{}
	err := row.Scan(&policy.ID, &policy.Name, &policy.Version, &policy.Content, &policy.Status, &policy.CreatedAt, &policy.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy with ID %s: %w", id, err)
	}
	return policy, nil
}

// UpdatePolicy updates an existing policy in the database
func (m *PostgresPolicyManager) UpdatePolicy(ctx context.Context, policy *Policy) error {
	query := `
	UPDATE policies
	SET name = $2, version = $3, content = $4, status = $5, updated_at = $6
	WHERE id = $1;`
	_, err := m.db.ExecContext(ctx, query, policy.ID, policy.Name, policy.Version, policy.Content, policy.Status, policy.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update policy with ID %s: %w", policy.ID, err)
	}
	m.logger.Infof("Successfully updated policy with ID: %s", policy.ID)
	return nil
}

// DeletePolicy removes a policy from the database by its ID
func (m *PostgresPolicyManager) DeletePolicy(ctx context.Context, id string) error {
	query := `DELETE FROM policies WHERE id = $1;`
	_, err := m.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete policy with ID %s: %w", id, err)
	}
	m.logger.Infof("Successfully deleted policy with ID: %s", id)
	return nil
}

// GetAllPolicies retrieves all policies from the database
func (m *PostgresPolicyManager) GetAllPolicies(ctx context.Context) ([]*Policy, error) {
	query := `SELECT id, name, version, content, status, created_at, updated_at FROM policies;`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all policies: %w", err)
	}
	defer rows.Close()

	var policies []*Policy
	for rows.Next() {
		policy := &Policy{}
		err := rows.Scan(&policy.ID, &policy.Name, &policy.Version, &policy.Content, &policy.Status, &policy.CreatedAt, &policy.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan policy row: %w", err)
		}
		policies = append(policies, policy)
	}
	return policies, nil
}
