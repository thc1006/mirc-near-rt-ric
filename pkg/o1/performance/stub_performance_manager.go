package performance

import (
	"context"

	"github.com/hctsai1006/near-rt-ric/pkg/o1/types"
)

// StubPerformanceManager is a stub implementation of the PerformanceManager interface
type StubPerformanceManager struct{}

// NewStubPerformanceManager creates a new StubPerformanceManager
func NewStubPerformanceManager() types.PerformanceManager {
	return &StubPerformanceManager{}
}

// GetKPIs returns hardcoded KPI data
func (m *StubPerformanceManager) GetKPIs(ctx context.Context) (string, error) {
	return "<kpis><throughput>100Gbps</throughput></kpis>", nil
}
