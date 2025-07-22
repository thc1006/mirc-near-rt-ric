package e2

import (
	"github.com/hctsai1006/near-rt-ric/pkg/e2/asn1"
	"github.com/hctsai1006/near-rt-ric/pkg/e2/e2common"
)

// E2Interface defines the interface for the O-RAN E2 interface.
// It provides methods for managing E2 connections, handling E2AP procedures,
// and interacting with E2 nodes.
type E2Interface interface {
	Start() error
	Stop() error
	HandleE2Setup(request *asn1.E2SetupRequest) (*asn1.E2SetupResponse, error)
	Subscribe(request *asn1.RICSubscription) (*asn1.RICSubscriptionResponse, error)
	ProcessIndication(indication *asn1.RICIndicationIEs) error
	GetStatus() map[string]interface{}
	GetOperationalNodes() []*e2common.E2Node
}
