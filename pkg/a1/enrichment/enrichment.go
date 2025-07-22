package enrichment

import (
	"context"
	"time"
)

// EnrichmentInfo represents enrichment information
type EnrichmentInfo struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Source    string        `json:"source"`
	Content   []byte        `json:"content"`
	Timestamp time.Time     `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
}

// EnrichmentManager defines the interface for managing enrichment information
type EnrichmentManager interface {
	UpdateInfo(ctx context.Context, info *EnrichmentInfo) error
	GetInfo(ctx context.Context, id string) (*EnrichmentInfo, error)
	GetAllInfo(ctx context.Context) ([]*EnrichmentInfo, error)
}

// MockEnrichmentManager is a mock implementation of EnrichmentManager
type MockEnrichmentManager struct{}

func (m *MockEnrichmentManager) UpdateInfo(ctx context.Context, info *EnrichmentInfo) error {
	return nil
}
func (m *MockEnrichmentManager) GetInfo(ctx context.Context, id string) (*EnrichmentInfo, error) {
	return nil, nil
}
func (m *MockEnrichmentManager) GetAllInfo(ctx context.Context) ([]*EnrichmentInfo, error) {
	return nil, nil
}