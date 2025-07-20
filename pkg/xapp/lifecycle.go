// pkg/xapp/lifecycle.go
package xapp

import "fmt"

// LifecycleManager manages the lifecycle of xApps
type LifecycleManager struct {
	xapps map[string]interface{}
}

// NewLifecycleManager creates a new LifecycleManager
func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{xapps: make(map[string]interface{})}
}

// RegisterxApp registers a new xApp
func (m *LifecycleManager) RegisterxApp(name string, xapp interface{}) {
	m.xapps[name] = xapp
	fmt.Printf("Registered xApp: %s\n", name)
}

// UnregisterxApp unregisters an xApp
func (m *LifecycleManager) UnregisterxApp(name string) {
	delete(m.xapps, name)
	fmt.Printf("Unregistered xApp: %s\n", name)
}

