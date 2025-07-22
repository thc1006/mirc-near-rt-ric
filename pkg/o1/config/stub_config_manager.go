package config

import (
	"context"

	"github.com/hctsai1006/near-rt-ric/pkg/o1/types"
)

// StubConfigurationManager is a stub implementation of the ConfigurationManager interface
type StubConfigurationManager struct{}

// NewStubConfigurationManager creates a new StubConfigurationManager
func NewStubConfigurationManager() types.ConfigurationManager {
	return &StubConfigurationManager{}
}

// GetConfig returns a hardcoded configuration string
func (m *StubConfigurationManager) GetConfig(ctx context.Context, target string) (string, error) {
	return "<config><interface><name>eth0</name></interface></config>", nil
}

// SetConfig is a no-op for the stub implementation
func (m *StubConfigurationManager) SetConfig(ctx context.Context, target string, config string) error {
	return nil
}
