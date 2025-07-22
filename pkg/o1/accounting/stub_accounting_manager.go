package accounting

import (
	"context"

	"github.com/hctsai1006/near-rt-ric/pkg/o1/types"
)

// StubAccountingManager is a stub implementation of the AccountingManager interface
type StubAccountingManager struct{}

// NewStubAccountingManager creates a new StubAccountingManager
func NewStubAccountingManager() types.AccountingManager {
	return &StubAccountingManager{}
}

// GetUsageData returns hardcoded usage data
func (m *StubAccountingManager) GetUsageData(ctx context.Context) (string, error) {
	return "<usage><user>admin</user><data>1024MB</data></usage>", nil
}
