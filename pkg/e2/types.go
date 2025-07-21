package e2

import (
	"time"
)

// E2APConstants defines E2AP protocol constants
const (
	E2APVersion      = 3
	E2SetupRequestID = 1
	E2SetupResponseID = 2
	E2SetupFailureID  = 3
	
	RICSubscriptionRequestID  = 12
	RICSubscriptionResponseID = 13
	RICSubscriptionFailureID  = 14
	
	RICIndicationID = 5
	
	// Standard SCTP port for E2 interface
	E2SCTPPort = 36421
)

// E2NodeType represents the type of E2 node
type E2NodeType int

const (
	E2NodeTypeUnknown E2NodeType = iota
	E2NodeTypeCU      // Centralized Unit
	E2NodeTypeDU      // Distributed Unit  
	E2NodeTypeENB     // eNodeB (4G)
	E2NodeTypeGNB     // gNodeB (5G)
)

// E2Node represents an E2 node with O-RAN compliant attributes
type E2Node struct {
	ID               string
	Type             E2NodeType
	PLMNIdentity     []byte
	Address          string
	Port             int
	FunctionList     []E2NodeFunction
	ConnectionStatus ConnectionStatus
	SetupComplete    bool
	LastHeartbeat    time.Time
}

// E2NodeFunction represents a RAN function supported by the E2 node
type E2NodeFunction struct {
	OID              string // Object Identifier
	Description      string
	Instance         int
	ServiceModelOID  string
}

// ConnectionStatus represents the status of E2 connection
type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
	StatusSetupInProgress
	StatusOperational
	StatusError
)

// E2APMessage represents a generic E2AP message structure
type E2APMessage struct {
	MessageType     E2APMessageType
	ProcedureCode   int32
	Criticality     Criticality
	TransactionID   *int32
	Payload         []byte
	Timestamp       time.Time
}

// E2APMessageType defines the type of E2AP message
type E2APMessageType int

const (
	MessageTypeInitiatingMessage E2APMessageType = iota
	MessageTypeSuccessfulOutcome
	MessageTypeUnsuccessfulOutcome
)

// Criticality defines the criticality level of E2AP elements
type Criticality int

const (
	CriticalityReject Criticality = iota
	CriticalityIgnore
	CriticalityNotify
)

// E2SetupRequest represents an E2 Setup Request message
type E2SetupRequest struct {
	TransactionID        int32
	GlobalE2NodeID       GlobalE2NodeID
	RANFunctions         []RANFunction
	E2NodeConfigUpdate   *E2NodeConfigUpdate
}

// E2SetupResponse represents an E2 Setup Response message  
type E2SetupResponse struct {
	TransactionID              int32
	GlobalRICID                GlobalRICID
	RANFunctionsAccepted       []RANFunctionIDItem
	RANFunctionsRejected       []RANFunctionIDCause
	E2NodeConfigUpdate         *E2NodeConfigUpdateAck
}

// GlobalE2NodeID represents the global E2 node identifier
type GlobalE2NodeID struct {
	PLMNIdentity  []byte
	E2NodeType    E2NodeType
	NodeIdentity  []byte
}

// GlobalRICID represents the global RIC identifier
type GlobalRICID struct {
	PLMNIdentity []byte
	RICInstance  []byte
}

// RANFunction represents a RAN function
type RANFunction struct {
	RANFunctionID          int32
	RANFunctionDefinition  []byte
	RANFunctionRevision    int32
}

// RANFunctionIDItem represents an accepted RAN function
type RANFunctionIDItem struct {
	RANFunctionID       int32
	RANFunctionRevision int32
}

// RANFunctionIDCause represents a rejected RAN function with cause
type RANFunctionIDCause struct {
	RANFunctionID int32
	Cause         Cause
}

// Cause represents the cause of rejection or failure
type Cause struct {
	CauseType  CauseType
	CauseValue int32
}

// CauseType defines the type of cause
type CauseType int

const (
	CauseTypeRIC CauseType = iota
	CauseTypeRANFunction
	CauseTypeTransport
	CauseTypeProtocol
	CauseTypeMisc
)

// E2NodeConfigUpdate represents E2 node configuration update
type E2NodeConfigUpdate struct {
	RANFunctionsAdded    []RANFunction
	RANFunctionsModified []RANFunction
	RANFunctionsDeleted  []RANFunctionID
}

// E2NodeConfigUpdateAck represents acknowledgment of config update
type E2NodeConfigUpdateAck struct {
	RANFunctionsAccepted []RANFunctionIDItem
	RANFunctionsRejected []RANFunctionIDCause
}

// RANFunctionID represents a RAN function identifier
type RANFunctionID struct {
	ID int32
}

// RICSubscription represents a RIC subscription
type RICSubscription struct {
	RICRequestID      RICRequestID
	RANFunctionID     int32
	RICSubscriptionDetails RICSubscriptionDetails
}

// RICRequestID represents a RIC request identifier
type RICRequestID struct {
	RICRequestorID int32
	RICInstanceID  int32
}

// RICSubscriptionDetails contains subscription details
type RICSubscriptionDetails struct {
	RICEventTriggerDefinition []byte
	RICActions               []RICAction
}

// RICAction represents a RIC action
type RICAction struct {
	RICActionID         int32
	RICActionType       RICActionType
	RICActionDefinition *[]byte
	RICSubsequentAction *RICSubsequentAction
}

// RICActionType defines the type of RIC action
type RICActionType int

const (
	RICActionTypeReport RICActionType = iota
	RICActionTypeInsert
	RICActionTypePolicy
)

// RICSubsequentAction defines subsequent action parameters
type RICSubsequentAction struct {
	RICSubsequentActionType RICSubsequentActionType
	RICTimeToWait          RICTimeToWait
}

// RICSubsequentActionType defines subsequent action types
type RICSubsequentActionType int

const (
	RICSubsequentActionTypeContinue RICSubsequentActionType = iota
	RICSubsequentActionTypeWait
)

// RICTimeToWait defines time to wait values
type RICTimeToWait int

const (
	RICTimeToWaitZero RICTimeToWait = iota
	RICTimeToWaitW1ms
	RICTimeToWaitW2ms
	RICTimeToWaitW5ms
	RICTimeToWaitW10ms
	RICTimeToWaitW20ms
	RICTimeToWaitW30ms
	RICTimeToWaitW40ms
	RICTimeToWaitW50ms
	RICTimeToWaitW100ms
	RICTimeToWaitW200ms
	RICTimeToWaitW500ms
	RICTimeToWaitW1s
	RICTimeToWaitW2s
	RICTimeToWaitW5s
	RICTimeToWaitW10s
	RICTimeToWaitW20s
	RICTimeToWaitW60s
)

// E2InterfaceHandler represents a handler for E2 interface messages
type E2InterfaceHandler interface {
	HandleMessage(nodeID string, message *E2APMessage) error
}

// E2InterfaceConfig contains configuration for E2 interface
type E2InterfaceConfig struct {
	ListenAddress string
	ListenPort    int
	MaxNodes      int
	HeartbeatInterval time.Duration
	ConnectionTimeout time.Duration
}