package security

import (
	"context"

	"github.com/hctsai1006/near-rt-ric/pkg/o1/types"
)

// StubSecurityManager is a stub implementation of the SecurityManager interface
type StubSecurityManager struct{}

// NewStubSecurityManager creates a new StubSecurityManager
func NewStubSecurityManager() types.SecurityManager {
	return &StubSecurityManager{}
}

// GetSecurityLogs returns a hardcoded list of security logs
func (m *StubSecurityManager) GetSecurityLogs(ctx context.Context) ([]string, error) {
	return []string{
		"Login successful for user admin",
		"Login failed for user guest",
	}, nil
}
