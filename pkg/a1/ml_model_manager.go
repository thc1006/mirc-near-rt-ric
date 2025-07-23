package a1

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/hctsai1006/near-rt-ric/internal/config"
	"github.com/sirupsen/logrus"
)

// MLModelManager manages A1 ML models
type MLModelManager struct {
	log     *logrus.Logger
	config  *config.A1Config
	db      *sql.DB
	mu      sync.RWMutex
}

// NewMLModelManager creates a new ML model manager
func NewMLModelManager(config *config.A1Config, log *logrus.Logger, db *sql.DB) *MLModelManager {
	return &MLModelManager{
		config: config,
		log:    log,
		db:     db,
	}
}

// GetMLModels returns all ML models
func (mm *MLModelManager) GetMLModels() ([]*MLModel, error) {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	rows, err := mm.db.Query("SELECT model_id, model_name, model_version, model_type, description, model_data, model_url, metadata, status, created_at, updated_at, deployed_at FROM ml_models")
	if err != nil {
		return nil, fmt.Errorf("failed to query ML models: %w", err)
	}
	defer rows.Close()

	models := make([]*MLModel, 0)
	for rows.Next() {
		model := &MLModel{}
		if err := rows.Scan(&model.ModelID, &model.ModelName, &model.ModelVersion, &model.ModelType, &model.Description, &model.ModelData, &model.ModelURL, &model.Metadata, &model.Status, &model.CreatedAt, &model.UpdatedAt, &model.DeployedAt); err != nil {
			return nil, fmt.Errorf("failed to scan ML model: %w", err)
		}
		models = append(models, model)
	}

	return models, nil
}

// DeleteMLModel deletes an ML model
func (mm *MLModelManager) DeleteMLModel(modelID string) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Check if ML model exists
	var count int
	err := mm.db.QueryRow("SELECT COUNT(*) FROM ml_models WHERE model_id = $1", modelID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing ML model: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("ML model not found")
	}

	// Delete from database
	_, err = mm.db.Exec("DELETE FROM ml_models WHERE model_id = $1", modelID)
	if err != nil {
		return fmt.Errorf("failed to delete ML model: %w", err)
	}

	mm.log.Infof("Deleted ML model %s", modelID)

	return nil
}
