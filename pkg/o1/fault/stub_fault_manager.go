package fault

import (
	"context"

	"github.com/hctsai1006/near-rt-ric/pkg/o1/types"
)

// StubFaultManager is a stub implementation of the FaultManager interface
type StubFaultManager struct{}

// NewStubFaultManager creates a new StubFaultManager
func NewStubFaultManager() types.FaultManager {
	return &StubFaultManager{}
}

// GetAlarms returns a hardcoded list of alarms
func (m *StubFaultManager) GetAlarms(ctx context.Context) ([]*types.Alarm, error) {
	return []*types.Alarm{
		{ID: "1", Severity: "critical", Message: "Link down"},
		{ID: "2", Severity: "major", Message: "High CPU usage"},
	}, nil
}
