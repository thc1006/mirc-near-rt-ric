package mlmodel

import (
	"context"
	"fmt"
	"time"

	"github.com/hctsai1006/near-rt-ric/pkg/a1/db"
	"github.com/sirupsen/logrus"
)

// PostgresMLModelManager implements the MLModelManager interface using a PostgreSQL database
type PostgresMLModelManager struct {
	db     *db.DB
	logger *logrus.Logger
}

// NewPostgresMLModelManager creates a new PostgresMLModelManager
func NewPostgresMLModelManager(db *db.DB) *PostgresMLModelManager {
	return &PostgresMLModelManager{
		db:     db,
		logger: logrus.WithField("component", "postgres-mlmodel-manager").Logger,
	}
}

// CreateMLModelsTable creates the ml_models table in the database if it doesn't exist
func (m *PostgresMLModelManager) CreateMLModelsTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS ml_models (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		version TEXT NOT NULL,
		file_path TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL
	);`
	_, err := m.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create ml_models table: %w", err)
	}
	m.logger.Info("ML models table created or already exists")
	return nil
}

// UploadModel adds a new ML model to the database
func (m *PostgresMLModelManager) UploadModel(ctx context.Context, model *MLModel) error {
	query := `
	INSERT INTO ml_models (id, name, version, file_path, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7);`
	_, err := m.db.ExecContext(ctx, query, model.ID, model.Name, model.Version, model.FilePath, model.Status, model.CreatedAt, model.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to upload ml model: %w", err)
	}
	m.logger.Infof("Successfully uploaded ml model with ID: %s", model.ID)
	return nil
}

// GetModel retrieves an ML model from the database by its ID
func (m *PostgresMLModelManager) GetModel(ctx context.Context, id string) (*MLModel, error) {
	query := `SELECT id, name, version, file_path, status, created_at, updated_at FROM ml_models WHERE id = $1;`
	row := m.db.QueryRowContext(ctx, query, id)
	model := &MLModel{}
	err := row.Scan(&model.ID, &model.Name, &model.Version, &model.FilePath, &model.Status, &model.CreatedAt, &model.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get ml model with ID %s: %w", id, err)
	}
	return model, nil
}

// GetAllModels retrieves all ML models from the database
func (m *PostgresMLModelManager) GetAllModels(ctx context.Context) ([]*MLModel, error) {
	query := `SELECT id, name, version, file_path, status, created_at, updated_at FROM ml_models;`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all ml models: %w", err)
	}
	defer rows.Close()

	var models []*MLModel
	for rows.Next() {
		model := &MLModel{}
		err := rows.Scan(&model.ID, &model.Name, &model.Version, &model.FilePath, &model.Status, &model.CreatedAt, &model.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ml model row: %w", err)
		}
		models = append(models, model)
	}
	return models, nil
}

// DeployModel updates the status of an ML model to "DEPLOYED"
func (m *PostgresMLModelManager) DeployModel(ctx context.Context, id string) error {
	query := `UPDATE ml_models SET status = $1, updated_at = $2 WHERE id = $3;`
	_, err := m.db.ExecContext(ctx, query, MLModelStatusDeployed, time.Now().Format(time.RFC3339), id)
	if err != nil {
		return fmt.Errorf("failed to deploy ml model with ID %s: %w", id, err)
	}
	m.logger.Infof("Successfully deployed ml model with ID: %s", id)
	return nil
}

// RollbackModel updates the status of an ML model to "INACTIVE"
func (m *PostgresMLModelManager) RollbackModel(ctx context.Context, id string) error {
	query := `UPDATE ml_models SET status = $1, updated_at = $2 WHERE id = $3;`
	_, err := m.db.ExecContext(ctx, query, MLModelStatusInactive, time.Now().Format(time.RFC3339), id)
	if err != nil {
		return fmt.Errorf("failed to rollback ml model with ID %s: %w", id, err)
	}
	m.logger.Infof("Successfully rolled back ml model with ID: %s", id)
	return nil
}
