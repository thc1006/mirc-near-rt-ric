package types

import (
	"context"
)

// FCAPSManager is a marker interface for all FCAPS managers
type FCAPSManager interface{}

// FaultManager defines the interface for fault management
type FaultManager interface {
	FCAPSManager
	GetAlarms(ctx context.Context) ([]*Alarm, error)
}

// ConfigurationManager defines the interface for configuration management
type ConfigurationManager interface {
	FCAPSManager
	GetConfig(ctx context.Context, target string) (string, error)
	SetConfig(ctx context.Context, target string, config string) error
}

// AccountingManager defines the interface for accounting management
type AccountingManager interface {
	FCAPSManager
	GetUsageData(ctx context.Context) (string, error)
}

// PerformanceManager defines the interface for performance management
type PerformanceManager interface {
	FCAPSManager
	GetKPIs(ctx context.Context) (string, error)
}

// SecurityManager defines the interface for security management
type SecurityManager interface {
	FCAPSManager
	GetSecurityLogs(ctx context.Context) ([]string, error)
}

// Alarm represents a fault alarm
type Alarm struct {
	ID       string
	Severity string
	Message  string
}
