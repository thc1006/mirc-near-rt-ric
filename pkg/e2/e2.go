package e2

import (
	"context"
	"fmt"
	"github.com/hctsai1006/near-rt-ric/pkg/e2/asn1"
	"github.com/hctsai1006/near-rt-ric/pkg/e2/e2common"
	"github.com/sirupsen/logrus"
	"time"
)

// Config holds the configuration for the E2 interface.
// It defines the connection parameters and operational settings required for the SCTP listener and node management.
type Config = e2common.E2InterfaceConfig

// E2InterfaceImpl is the main handler for the O-RAN E2 interface.
// It orchestrates the SCTP, node management, and E2AP processing components to ensure compliant E2 procedures.
type E2InterfaceImpl struct {
	config        *Config
	sctpManager   *e2common.SCTPManager
	nodeManager   *e2common.E2NodeManager
	e2apProcessor *asn1.E2APProcessor
	logger        *logrus.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	running       bool
}

// NewE2Interface creates a new E2Interface with O-RAN compliance.
// It initializes all the necessary components, including the SCTP manager, node manager, and E2AP processor.
func NewE2Interface(config *Config) E2Interface {
	ctx, cancel := context.WithCancel(context.Background())

	logger := logrus.WithField("component", "e2-interface").Logger

	// Create SCTP manager
	sctpConfig := &e2common.SCTPConfig{
		ListenAddress:     config.ListenAddress,
		ListenPort:        config.ListenPort,
		ConnectTimeout:    config.ConnectionTimeout,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      10 * time.Second,
		KeepaliveInterval: config.HeartbeatInterval,
	}
	sctpManager := e2common.NewSCTPManager(sctpConfig)

	// Create node manager
	nodeManager := e2common.NewE2NodeManager(config)

	// Create E2AP processor
	e2apProcessor := asn1.NewE2APProcessor(sctpManager, nodeManager)

	// Create main interface
	e2Interface := &E2InterfaceImpl{
		config:        config,
		sctpManager:   sctpManager,
		nodeManager:   nodeManager,
		e2apProcessor: e2apProcessor,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
		running:       false,
	}

	// Set up message handling
	sctpManager.SetMessageHandler(e2Interface.handleIncomingMessage)

	return e2Interface
}

// Start starts the E2 interface and begins listening for connections.
func (e2 *E2InterfaceImpl) Start() error {
	if e2.running {
		return fmt.Errorf("E2 interface is already running")
	}

	e2.logger.WithFields(logrus.Fields{
		"listen_address": e2.config.ListenAddress,
		"listen_port":    e2.config.ListenPort,
		"max_nodes":      e2.config.MaxNodes,
	}).Info("Starting O-RAN E2 interface")

	// Start SCTP listener
	if err := e2.sctpManager.Listen(); err != nil {
		return fmt.Errorf("failed to start SCTP listener: %w", err)
	}

	// Start health monitoring
	go e2.healthMonitor()

	e2.running = true
	e2.logger.Info("O-RAN E2 interface started successfully")

	return nil
}

// Stop stops the E2 interface and cleans up resources.
func (e2 *E2InterfaceImpl) Stop() error {
	if !e2.running {
		return nil
	}

	e2.logger.Info("Stopping O-RAN E2 interface")

	// Cancel context to stop all goroutines
	e2.cancel()

	// Close SCTP manager
	if err := e2.sctpManager.Close(); err != nil {
		e2.logger.WithError(err).Error("Error closing SCTP manager")
	}

	// Cleanup node manager
	e2.nodeManager.Cleanup()

	// Cleanup E2AP processor
	e2.e2apProcessor.Cleanup()

	e2.running = false
	e2.logger.Info("O-RAN E2 interface stopped successfully")

	return nil
}

// handleIncomingMessage processes messages received from the SCTP manager.
// It decodes the E2AP PDU and routes it to the appropriate handler based on the procedure code.
func (e2 *E2InterfaceImpl) handleIncomingMessage(nodeID string, data []byte) error {
	start := time.Now()

	e2.logger.WithFields(logrus.Fields{
		"node_id":   nodeID,
		"data_size": len(data),
	}).Debug("Received message from E2 node")

	// Update node heartbeat
	if err := e2.nodeManager.UpdateHeartbeat(nodeID); err != nil {
		e2.logger.WithError(err).WithField("node_id", nodeID).Warn("Failed to update node heartbeat")
	}

	// Decode message to determine procedure code
	pdu, err := e2.e2apProcessor.ASN1Encoder.DecodeE2AP_PDU(data)
	if err != nil {
		e2.logger.WithError(err).WithField("node_id", nodeID).Error("Failed to decode E2AP PDU")
		return err
	}

	// Route message based on procedure code
	if pdu.InitiatingMessage != nil {
		switch pdu.InitiatingMessage.ProcedureCode {
		case asn1.E2SetupRequestID:
			err = e2.e2apProcessor.ProcessE2SetupRequest(nodeID, data)
		case asn1.RICSubscriptionRequestID:
			err = e2.e2apProcessor.ProcessSubscriptionRequest(nodeID, data)
		case asn1.RICIndicationID:
			err = e2.e2apProcessor.ProcessIndication(nodeID, data)
		default:
			e2.logger.WithFields(logrus.Fields{
				"node_id":        nodeID,
				"procedure_code": pdu.InitiatingMessage.ProcedureCode,
			}).Warn("Unknown E2AP procedure code received")
			err = fmt.Errorf("unknown procedure code: %d", pdu.InitiatingMessage.ProcedureCode)
		}
	} else {
		e2.logger.WithField("node_id", nodeID).Warn("Received non-initiating message")
		// Handle successful/unsuccessful outcomes here
	}

	e2.logger.WithFields(logrus.Fields{
		"node_id":  nodeID,
		"duration": time.Since(start),
		"success":  err == nil,
	}).Debug("Message processing completed")

	return err
}

// HandleE2Setup processes the E2 Setup Request by invoking the E2APProcessor.
// It returns a failure response if the processor fails to handle the request.
func (e2 *E2InterfaceImpl) HandleE2Setup(request *asn1.E2SetupRequest) (*asn1.E2SetupResponse, error) {
	// In a real implementation, you would extract nodeID from the request.
	// For now, we'll use a placeholder.
	nodeID := "test-node-id"

	// Convert the generic E2SetupRequest to asn1.E2SetupRequest if necessary
	// For now, assuming direct compatibility or handling conversion within ProcessE2SetupRequest

	if err := e2.e2apProcessor.ProcessE2SetupRequest(nodeID, []byte{}); err != nil { // Pass raw bytes for now
		e2.logger.WithError(err).Error("Failed to process E2 Setup Request")
		// Return a failure response to the E2 node
		return nil, fmt.Errorf("failed to process E2 Setup: %w", err)
	}
	// The response is sent asynchronously by the processor, so we return nil here.
	return &asn1.E2SetupResponse{}, nil
}

// Subscribe processes the RIC Subscription Request through the E2APProcessor.
// It returns a failure response if the processor encounters an error.
func (e2 *E2InterfaceImpl) Subscribe(request *asn1.RICSubscription) (*asn1.RICSubscriptionResponse, error) {
	// In a real implementation, you would extract nodeID from the request.
	// For now, we'll use a placeholder.
	nodeID := "test-node-id"

	// Convert the generic SubscriptionRequest to asn1.RICSubscription if necessary
	// For now, assuming direct compatibility or handling conversion within ProcessSubscriptionRequest

	if err := e2.e2apProcessor.ProcessSubscriptionRequest(nodeID, []byte{}); err != nil { // Pass raw bytes for now
		e2.logger.WithError(err).Error("Failed to process RIC Subscription Request")
		// Return a failure response to the E2 node
		return nil, fmt.Errorf("failed to process subscription: %w", err)
	}
	// The response is sent asynchronously, so we return nil.
	return &asn1.RICSubscriptionResponse{}, nil
}

// ProcessIndication forwards the RIC Indication to the E2APProcessor for handling.
func (e2 *E2InterfaceImpl) ProcessIndication(indication *asn1.RICIndicationIEs) error {
	// In a real implementation, you would extract nodeID from the indication.
	// For now, we'll use a placeholder.
	nodeID := "test-node-id"

	// Convert the generic Indication to asn1.RICIndicationIEs if necessary
	// For now, assuming direct compatibility or handling conversion within ProcessIndication

	if err := e2.e2apProcessor.ProcessIndication(nodeID, []byte{}); err != nil { // Pass raw bytes for now
		e2.logger.WithError(err).Error("Failed to process RIC Indication")
		return fmt.Errorf("failed to process RIC Indication: %w", err)
	}
	return nil
}

// GetStatus returns the current status of the E2 interface.
func (e2 *E2InterfaceImpl) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})
	status["running"] = e2.running
	status["node_manager_status"] = e2.nodeManager.GetNodeStatistics()
	status["connection_statistics"] = e2.sctpManager.GetConnectionStatistics()
	return status
}

// GetOperationalNodes returns a list of all operational E2 nodes.
func (e2 *E2InterfaceImpl) GetOperationalNodes() []*e2common.E2Node {
	return e2.nodeManager.GetOperationalNodes()
}

// healthMonitor periodically checks the health of the E2 interface and its nodes.
func (e2 *E2InterfaceImpl) healthMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e2.ctx.Done():
			return
		case <-ticker.C:
			e2.performHealthCheck()
		}
	}
}

// performHealthCheck gathers and logs health statistics for the E2 interface.
func (e2 *E2InterfaceImpl) performHealthCheck() {
	// Get current statistics
	nodeStats := e2.nodeManager.GetNodeStatistics()
	connStats := e2.sctpManager.GetConnectionStatistics()

	e2.logger.WithFields(logrus.Fields{
		"total_nodes":        nodeStats["total_nodes"],
		"operational_nodes":  nodeStats["operational_nodes"],
		"active_connections": connStats["active_connections"],
		"capacity_usage":     nodeStats["capacity_usage"],
	}).Info("E2 Interface health check")

	// Check for any nodes in error state
	allNodes := e2.nodeManager.GetAllNodes()
	errorNodes := 0
	for _, node := range allNodes {
		if node.ConnectionStatus == e2common.StatusError {
			errorNodes++
		}
	}

	if errorNodes > 0 {
		e2.logger.WithField("error_nodes", errorNodes).Warn("Some E2 nodes are in error state")
	}
}
