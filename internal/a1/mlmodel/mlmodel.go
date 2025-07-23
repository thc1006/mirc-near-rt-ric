package mlmodel

import "context"

// MLModelStatus represents the status of an ML model
type MLModelStatus string

const (
	MLModelStatusNew       MLModelStatus = "NEW"
	MLModelStatusDeployed  MLModelStatus = "DEPLOYED"
	MLModelStatusInactive  MLModelStatus = "INACTIVE"
	MLModelStatusError     MLModelStatus = "ERROR"
)

// MLModel represents a machine learning model
type MLModel struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Version   string        `json:"version"`
	FilePath  string        `json:"file_path"`
	Status    MLModelStatus `json:"status"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

// MLModelManager defines the interface for managing ML models
type MLModelManager interface {
	UploadModel(ctx context.Context, model *MLModel) error
	DeployModel(ctx context.Context, id string) error
	GetModel(ctx context.Context, id string) (*MLModel, error)
	GetAllModels(ctx context.Context) ([]*MLModel, error)
	RollbackModel(ctx context.Context, id string) error
}

// MockMLModelManager is a mock implementation of MLModelManager
type MockMLModelManager struct{}

func (m *MockMLModelManager) UploadModel(ctx context.Context, model *MLModel) error    { return nil }
func (m *MockMLModelManager) DeployModel(ctx context.Context, id string) error          { return nil }
func (m *MockMLModelManager) GetModel(ctx context.Context, id string) (*MLModel, error) { return nil, nil }
func (m *MockMLModelManager) GetAllModels(ctx context.Context) ([]*MLModel, error)      { return nil, nil }
func (m *MockMLModelManager) RollbackModel(ctx context.Context, id string) error        { return nil }