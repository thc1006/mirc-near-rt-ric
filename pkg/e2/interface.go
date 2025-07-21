package e2

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// E2Interface is the main handler for the O-RAN E2 interface
type E2Interface struct {
	config       *E2InterfaceConfig
	sctpManager  *SCTPManager
	nodeManager  *E2NodeManager
	e2apProcessor *E2APProcessor
	logger       *logrus.Logger
	ctx          context.Context
	cancel       context.CancelFunc
	running      bool
}

// NewE2Interface creates a new E2Interface with O-RAN compliance
func NewE2Interface(config *E2InterfaceConfig) *E2Interface {
	ctx, cancel := context.WithCancel(context.Background())
	
	logger := logrus.WithField("component", "e2-interface").Logger
	
	// Create SCTP manager
	sctpConfig := &SCTPConfig{
		ListenAddress:     config.ListenAddress,
		ListenPort:        config.ListenPort,
		ConnectTimeout:    config.ConnectionTimeout,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      10 * time.Second,
		KeepaliveInterval: config.HeartbeatInterval,
	}
	sctpManager := NewSCTPManager(sctpConfig)
	
	// Create node manager
	nodeManager := NewE2NodeManager(config)
	
	// Create E2AP processor
	e2apProcessor := NewE2APProcessor(sctpManager, nodeManager)
	
	// Create main interface
	e2Interface := &E2Interface{
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

// Start starts the E2 interface and begins listening for connections
func (e2 *E2Interface) Start() error {
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

// Stop stops the E2 interface and cleans up resources
func (e2 *E2Interface) Stop() error {
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

// GetStatus returns the current status of the E2 interface
func (e2 *E2Interface) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"running":              e2.running,
		"listen_address":       e2.config.ListenAddress,
		"listen_port":          e2.config.ListenPort,
		"connection_statistics": e2.sctpManager.GetConnectionStatistics(),
		"node_statistics":      e2.nodeManager.GetNodeStatistics(),
		"connection_status":    e2.sctpManager.GetConnectionStatus(),
	}
	
	return status
}

// GetNodes returns all registered E2 nodes
func (e2 *E2Interface) GetNodes() map[string]*E2Node {
	return e2.nodeManager.GetAllNodes()
}

// GetNode returns a specific E2 node by ID
func (e2 *E2Interface) GetNode(nodeID string) (*E2Node, error) {
	return e2.nodeManager.GetNode(nodeID)
}

// GetOperationalNodes returns all nodes in operational status
func (e2 *E2Interface) GetOperationalNodes() []*E2Node {
	return e2.nodeManager.GetOperationalNodes()
}

// SendToNode sends data to a specific E2 node
func (e2 *E2Interface) SendToNode(nodeID string, data []byte) error {
	return e2.sctpManager.SendToNode(nodeID, data)
}

// BroadcastToAllNodes sends data to all connected E2 nodes
func (e2 *E2Interface) BroadcastToAllNodes(data []byte) []error {
	return e2.sctpManager.BroadcastToAllNodes(data)
}

// DisconnectNode forcibly disconnects a specific E2 node
func (e2 *E2Interface) DisconnectNode(nodeID string) error {
	// Disconnect from SCTP
	if err := e2.sctpManager.DisconnectNode(nodeID); err != nil {
		e2.logger.WithError(err).WithField("node_id", nodeID).Error("Failed to disconnect SCTP connection")
	}
	
	// Remove from node manager
	if err := e2.nodeManager.RemoveNode(nodeID); err != nil {
		e2.logger.WithError(err).WithField("node_id", nodeID).Error("Failed to remove node from manager")
		return err
	}
	
	return nil
}

// HandleE2SetupRequest processes an E2 setup request from a node
func (e2 *E2Interface) HandleE2SetupRequest(nodeID string, requestData []byte) error {
	return e2.e2apProcessor.ProcessE2SetupRequest(nodeID, requestData)
}

// HandleSubscriptionRequest processes a RIC subscription request
func (e2 *E2Interface) HandleSubscriptionRequest(nodeID string, requestData []byte) error {
	return e2.e2apProcessor.ProcessSubscriptionRequest(nodeID, requestData)
}

// HandleIndication processes a RIC indication from an E2 node
func (e2 *E2Interface) HandleIndication(nodeID string, indicationData []byte) error {
	return e2.e2apProcessor.ProcessIndication(nodeID, indicationData)
}

// Private methods

func (e2 *E2Interface) handleIncomingMessage(nodeID string, data []byte) error {
	start := time.Now()
	
	e2.logger.WithFields(logrus.Fields{
		"node_id": nodeID,
		"data_size": len(data),
	}).Debug("Received message from E2 node")
	
	// Update node heartbeat
	if err := e2.nodeManager.UpdateHeartbeat(nodeID); err != nil {
		e2.logger.WithError(err).WithField("node_id", nodeID).Warn("Failed to update node heartbeat")
	}
	
	// Decode message to determine procedure code
	pdu, err := e2.e2apProcessor.asn1Encoder.DecodeE2AP_PDU(data)
	if err != nil {
		e2.logger.WithError(err).WithField("node_id", nodeID).Error("Failed to decode E2AP PDU")
		return err
	}
	
	// Route message based on procedure code
	if pdu.InitiatingMessage != nil {
		switch pdu.InitiatingMessage.ProcedureCode {
		case E2SetupRequestID:
			err = e2.HandleE2SetupRequest(nodeID, data)
		case RICSubscriptionRequestID:
			err = e2.HandleSubscriptionRequest(nodeID, data)
		case RICIndicationID:
			err = e2.HandleIndication(nodeID, data)
		default:
			e2.logger.WithFields(logrus.Fields{
				"node_id": nodeID,
				"procedure_code": pdu.InitiatingMessage.ProcedureCode,
			}).Warn("Unknown E2AP procedure code received")
			err = fmt.Errorf("unknown procedure code: %d", pdu.InitiatingMessage.ProcedureCode)
		}
	} else {
		e2.logger.WithField("node_id", nodeID).Warn("Received non-initiating message")
		// Handle successful/unsuccessful outcomes here
	}
	
	e2.logger.WithFields(logrus.Fields{
		"node_id": nodeID,
		"duration": time.Since(start),
		"success": err == nil,
	}).Debug("Message processing completed")
	
	return err
}

func (e2 *E2Interface) healthMonitor() {
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

func (e2 *E2Interface) performHealthCheck() {
	// Get current statistics
	nodeStats := e2.nodeManager.GetNodeStatistics()
	connStats := e2.sctpManager.GetConnectionStatistics()
	
	e2.logger.WithFields(logrus.Fields{
		"total_nodes": nodeStats["total_nodes"],
		"operational_nodes": nodeStats["operational_nodes"],
		"active_connections": connStats["active_connections"],
		"capacity_usage": nodeStats["capacity_usage"],
	}).Info("E2 Interface health check")
	
	// Check for any nodes in error state
	allNodes := e2.nodeManager.GetAllNodes()
	errorNodes := 0
	for _, node := range allNodes {
		if node.ConnectionStatus == StatusError {
			errorNodes++
		}
	}
	
	if errorNodes > 0 {
		e2.logger.WithField("error_nodes", errorNodes).Warn("Some E2 nodes are in error state")
	}
}

// DefaultE2InterfaceConfig returns a default configuration for the E2 interface
func DefaultE2InterfaceConfig() *E2InterfaceConfig {
	return &E2InterfaceConfig{
		ListenAddress:     "0.0.0.0",
		ListenPort:        E2SCTPPort,
		MaxNodes:          100,
		HeartbeatInterval: 30 * time.Second,
		ConnectionTimeout: 60 * time.Second,
	}
}