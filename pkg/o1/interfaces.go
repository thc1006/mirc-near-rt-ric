// pkg/o1/interfaces.go
package o1

import "C"

type O1Interface interface {
    StartNetconfServer() error
    HandleRPCRequest(*netconf.RPCRequest) (*netconf.RPCResponse, error)
    SendNotification(*Notification) error
    ManageConfiguration(*ConfigOperation) error
    CollectPerformanceData() error
}

// Dummy structs to satisfy the interface definition for now.
// These will be replaced with real implementations.
type netconf struct{}
type RPCRequest struct{}
type RPCResponse struct{}
type Notification struct{}
type ConfigOperation struct{}
