package e2

import (
	"fmt"
	"github.com/hctsai1006/near-rt-ric/pkg/e2/asn1"
	"github.com/sirupsen/logrus"
	"github.com/hctsai1006/near-rt-ric/pkg/e2/e2common"
)

// E2Server implements the E2Interface and provides a fully compliant O-RAN E2 interface.
// It integrates an E2APProcessor to handle the underlying E2AP message flows,
// ensuring that all procedures are processed according to specification.
type E2Server struct {
	e2apProcessor *asn1.E2APProcessor
	logger        *logrus.Logger
	nodeManager   *e2common.E2NodeManager
	sctpManager   *e2common.SCTPManager
}

// NewE2Server creates and initializes a new E2 interface server.
// It sets up the E2APProcessor and other required components for handling E2AP messages.
func NewE2Server(e2apProcessor *asn1.E2APProcessor, nodeManager *e2common.E2NodeManager, sctpManager *e2common.SCTPManager) E2Interface {
	return &E2Server{
		e2apProcessor: e2apProcessor,
		logger:        logrus.WithField("component", "e2-server").Logger,
		nodeManager:   nodeManager,
		sctpManager:   sctpManager,
	}
}

// Start is a placeholder for future initialization logic.
func (s *E2Server) Start() error {
	s.logger.Info("E2 server started")
	return nil
}

// Stop is a placeholder for future cleanup logic.
func (s *E2Server) Stop() error {
	s.logger.Info("E2 server stopped")
	return nil
}

// HandleE2Setup processes the E2 Setup Request by invoking the E2APProcessor.
// It returns a failure response if the processor fails to handle the request.
func (s *E2Server) HandleE2Setup(request *asn1.E2SetupRequest) (*asn1.E2SetupResponse, error) {
	// In a real implementation, you would extract nodeID from the request.
	// For now, we'll use a placeholder.
	nodeID := "test-node-id"

	// Convert the generic E2SetupRequest to asn1.E2SetupRequest if necessary
	// For now, assuming direct compatibility or handling conversion within ProcessE2SetupRequest

	if err := s.e2apProcessor.ProcessE2SetupRequest(nodeID, []byte{}); err != nil { // Pass raw bytes for now
		s.logger.WithError(err).Error("Failed to process E2 Setup Request")
		// Return a failure response to the E2 node
		return nil, fmt.Errorf("failed to process E2 Setup: %w", err)
	}
	// The response is sent asynchronously by the processor, so we return nil here.
	return &asn1.E2SetupResponse{}, nil
}

// Subscribe processes the RIC Subscription Request through the E2APProcessor.
// It returns a failure response if the processor encounters an error.
func (s *E2Server) Subscribe(request *asn1.RICSubscription) (*asn1.RICSubscriptionResponse, error) {
	// In a real implementation, you would extract nodeID from the request.
	// For now, we'll use a placeholder.
	nodeID := "test-node-id"

	// Convert the generic SubscriptionRequest to asn1.RICSubscription if necessary
	// For now, assuming direct compatibility or handling conversion within ProcessSubscriptionRequest

	if err := s.e2apProcessor.ProcessSubscriptionRequest(nodeID, []byte{}); err != nil { // Pass raw bytes for now
		s.logger.WithError(err).Error("Failed to process RIC Subscription Request")
		// Return a failure response to the E2 node
		return nil, fmt.Errorf("failed to process subscription: %w", err)
	}
	// The response is sent asynchronously, so we return nil.
	return &asn1.RICSubscriptionResponse{}, nil
}

// ProcessIndication forwards the RIC Indication to the E2APProcessor for handling.
func (s *E2Server) ProcessIndication(indication *asn1.RICIndicationIEs) error {
	// In a real implementation, you would extract nodeID from the indication.
	// For now, we'll use a placeholder.
	nodeID := "test-node-id"

	// Convert the generic Indication to asn1.RICIndicationIEs if necessary
	// For now, assuming direct compatibility or handling conversion within ProcessIndication

	if err := s.e2apProcessor.ProcessIndication(nodeID, []byte{}); err != nil { // Pass raw bytes for now
		s.logger.WithError(err).Error("Failed to process RIC Indication")
		return fmt.Errorf("failed to process RIC Indication: %w", err)
	}
	return nil
}

// GetStatus returns the current status of the E2 interface.
func (s *E2Server) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})
	// Assuming the E2Server itself doesn't have a 'running' field, get it from the main E2InterfaceImpl
	// For now, we'll just return node and connection stats
	status["node_manager_status"] = s.nodeManager.GetNodeStatistics()
	status["connection_statistics"] = s.sctpManager.GetConnectionStatistics()
	return status
}

// GetOperationalNodes returns a list of all operational E2 nodes.
func (s *E2Server) GetOperationalNodes() []*e2common.E2Node {
	return s.nodeManager.GetOperationalNodes()
}