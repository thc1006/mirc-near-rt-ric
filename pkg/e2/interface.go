package e2

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hctsai1006/near-rt-ric/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// E2Interface implements the O-RAN E2 interface with full SCTP and ASN.1 support
type E2Interface struct {
	config          *config.E2Config
	sctpServer      *SCTPServer
	nodeManager     *NodeManager
	asn1Codec       *ASN1Codec
	subscriptionMgr *SubscriptionManager
	logger          *logrus.Logger
	
	// Metrics
	metrics         *E2Metrics
	
	// Message processing
	messageQueue    chan *E2Message
	workerPool      *WorkerPool
	
	// Lifecycle management
	ctx             context.Context
	cancel          context.CancelFunc
	running         bool
	mutex           sync.RWMutex
	wg              sync.WaitGroup
	startTime       time.Time
}

// E2Status represents the current status of the E2 interface
type E2Status struct {
	Running        bool          `json:"running"`
	ListenAddress  string        `json:"listen_address"`
	Port           int           `json:"port"`
	Connections    int           `json:"connections"`
	MaxConnections int           `json:"max_connections"`
	Nodes          int           `json:"nodes"`
	Subscriptions  int           `json:"subscriptions"`
	LastActivity   time.Time     `json:"last_activity"`
	Uptime         time.Duration `json:"uptime"`
}

// E2Message represents an E2AP message
type E2Message struct {
	ConnectionID string
	NodeID       string
	Data         []byte
	MessageType  E2MessageType
	Timestamp    time.Time
}

// E2MessageType represents the type of E2AP message
type E2MessageType int

const (
	E2SetupRequestMsg E2MessageType = iota
	E2SetupResponseMsg
	E2SetupFailureMsg
	RICSubscriptionRequestMsg
	RICSubscriptionResponseMsg
	RICSubscriptionFailureMsg
	RICIndicationMsg
	RICControlRequestMsg
	RICControlAckMsg
	RICControlFailureMsg
)

// NewE2Interface creates a new E2Interface with full O-RAN compliance
func NewE2Interface(cfg *config.E2Config, logger *logrus.Logger) (*E2Interface, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create logger with component field
	compLogger := logger.WithField("component", "e2-interface")
	
	// Create SCTP server
	sctpServer, err := NewSCTPServer(&SCTPConfig{
		ListenAddress:     cfg.ListenAddress,
		Port:              cfg.Port,
		MaxConnections:    cfg.MaxConnections,
		ConnectionTimeout: cfg.ConnectionTimeout,
		HeartbeatInterval: cfg.HeartbeatInterval,
		BufferSize:        cfg.BufferSize,
		SCTP:              cfg.SCTP,
	}, compLogger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create SCTP server: %w", err)
	}
	
	// Create ASN.1 codec with O-RAN E2AP specifications
	asn1Codec := NewASN1Codec(cfg.ASN1, compLogger)
	
	// Create node manager for E2 node lifecycle
	nodeManager := NewNodeManager(cfg.MaxConnections, compLogger)
	
	// Create subscription manager for RIC subscriptions
	subscriptionMgr := NewSubscriptionManager(compLogger)
	
	// Initialize metrics
	metrics := NewE2Metrics()
	
	// Create message queue with buffer
	messageQueue := make(chan *E2Message, 1000)
	
	// Create worker pool for message processing
	workerPool := NewWorkerPool(cfg.WorkerPoolSize, compLogger)
	
	// Create main interface
	e2Interface := &E2Interface{
		config:          cfg,
		sctpServer:      sctpServer,
		nodeManager:     nodeManager,
		asn1Codec:       asn1Codec,
		subscriptionMgr: subscriptionMgr,
		logger:          compLogger,
		metrics:         metrics,
		messageQueue:    messageQueue,
		workerPool:      workerPool,
		ctx:             ctx,
		cancel:          cancel,
		running:         false,
		startTime:       time.Now(),
	}
	
	// Set up message handlers
	sctpServer.SetMessageHandler(e2Interface.handleIncomingMessage)
	sctpServer.SetConnectionHandler(e2Interface.handleConnectionEvent)
	
	return e2Interface, nil
}

// Start starts the E2 interface and begins listening for connections
func (e2 *E2Interface) Start(ctx context.Context) error {
	e2.mutex.Lock()
	defer e2.mutex.Unlock()
	
	if e2.running {
		return fmt.Errorf("E2 interface is already running")
	}
	
	e2.logger.WithFields(logrus.Fields{
		"listen_address":  e2.config.ListenAddress,
		"port":            e2.config.Port,
		"max_connections": e2.config.MaxConnections,
	}).Info("Starting O-RAN E2 interface")
	
	// Register Prometheus metrics
	if err := e2.metrics.Register(); err != nil {
		e2.logger.WithError(err).Warn("Failed to register E2 metrics")
	}
	
	// Start SCTP server
	if err := e2.sctpServer.Start(e2.ctx); err != nil {
		return fmt.Errorf("failed to start SCTP server: %w", err)
	}
	
	// Start worker pool
	if err := e2.workerPool.Start(e2.ctx); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}
	
	// Start background workers
	g, gctx := errgroup.WithContext(e2.ctx)
	
	g.Go(func() error {
		return e2.healthMonitor(gctx)
	})
	
	g.Go(func() error {
		return e2.messageProcessor(gctx)
	})
	
	g.Go(func() error {
		return e2.subscriptionWorker(gctx)
	})
	
	g.Go(func() error {
		return e2.nodeMaintenanceWorker(gctx)
	})
	
	// Start the error group in the background
	e2.wg.Add(1)
	go func() {
		defer e2.wg.Done()
		if err := g.Wait(); err != nil {
			e2.logger.WithError(err).Error("Background worker error")
		}
	}()
	
	e2.running = true
	e2.startTime = time.Now()
	e2.logger.Info("O-RAN E2 interface started successfully")
	
	// Record startup metric
	e2.metrics.InterfaceStatus.WithLabelValues("e2").Set(1)
	
	return nil
}

// Stop gracefully stops the E2 interface and cleans up resources
func (e2 *E2Interface) Stop(ctx context.Context) error {
	e2.mutex.Lock()
	defer e2.mutex.Unlock()
	
	if !e2.running {
		return nil
	}
	
	e2.logger.Info("Stopping O-RAN E2 interface")
	
	// Cancel internal context to stop all goroutines
	e2.cancel()
	
	// Stop worker pool
	if err := e2.workerPool.Stop(ctx); err != nil {
		e2.logger.WithError(err).Error("Error stopping worker pool")
	}
	
	// Stop SCTP server with timeout
	stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	if err := e2.sctpServer.Stop(stopCtx); err != nil {
		e2.logger.WithError(err).Error("Error stopping SCTP server")
	}
	
	// Wait for background workers to finish
	done := make(chan struct{})
	go func() {
		e2.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		e2.logger.Debug("All background workers stopped")
	case <-ctx.Done():
		e2.logger.Warn("Shutdown timeout reached, forcing exit")
	}
	
	// Close message queue
	close(e2.messageQueue)
	
	// Cleanup components
	e2.subscriptionMgr.Cleanup()
	e2.nodeManager.Cleanup()
	
	// Unregister metrics
	e2.metrics.Unregister()
	
	e2.running = false
	e2.logger.Info("O-RAN E2 interface stopped successfully")
	
	// Record shutdown metric
	e2.metrics.InterfaceStatus.WithLabelValues("e2").Set(0)
	
	return nil
}

// GetStatus returns the current status of the E2 interface
func (e2 *E2Interface) GetStatus() *E2Status {
	e2.mutex.RLock()
	defer e2.mutex.RUnlock()
	
	return &E2Status{
		Running:        e2.running,
		ListenAddress:  e2.config.ListenAddress,
		Port:           e2.config.Port,
		Connections:    e2.sctpServer.GetConnectionCount(),
		MaxConnections: e2.config.MaxConnections,
		Nodes:          e2.nodeManager.GetNodeCount(),
		Subscriptions:  e2.subscriptionMgr.GetSubscriptionCount(),
		LastActivity:   e2.nodeManager.GetLastActivity(),
		Uptime:         time.Since(e2.startTime),
	}
}

// HealthCheck performs a comprehensive health check of the E2 interface
func (e2 *E2Interface) HealthCheck() error {
	e2.mutex.RLock()
	defer e2.mutex.RUnlock()
	
	if !e2.running {
		return fmt.Errorf("E2 interface is not running")
	}
	
	// Check SCTP server health
	if err := e2.sctpServer.HealthCheck(); err != nil {
		return fmt.Errorf("SCTP server health check failed: %w", err)
	}
	
	// Check node manager health
	if err := e2.nodeManager.HealthCheck(); err != nil {
		return fmt.Errorf("node manager health check failed: %w", err)
	}
	
	// Check subscription manager health
	if err := e2.subscriptionMgr.HealthCheck(); err != nil {
		return fmt.Errorf("subscription manager health check failed: %w", err)
	}
	
	// Check worker pool health
	if err := e2.workerPool.HealthCheck(); err != nil {
		return fmt.Errorf("worker pool health check failed: %w", err)
	}
	
	return nil
}

// GetNodes returns all registered E2 nodes
func (e2 *E2Interface) GetNodes() map[string]*E2Node {
	return e2.nodeManager.GetAllNodes()
}

// GetNode returns a specific E2 node by ID
func (e2 *E2Interface) GetNode(nodeID string) (*E2Node, error) {
	return e2.nodeManager.GetNode(nodeID)
}

// SendRICControlRequest sends a RIC control request to a specific E2 node
func (e2 *E2Interface) SendRICControlRequest(nodeID string, request *RICControlRequest) error {
	// Encode the control request
	data, err := e2.asn1Codec.EncodeRICControlRequest(request)
	if err != nil {
		e2.metrics.MessagesTotal.WithLabelValues("control_request", "encode_error").Inc()
		return fmt.Errorf("failed to encode RIC control request: %w", err)
	}
	
	// Send to node
	if err := e2.sctpServer.SendToNode(nodeID, data); err != nil {
		e2.metrics.MessagesTotal.WithLabelValues("control_request", "send_error").Inc()
		return fmt.Errorf("failed to send RIC control request: %w", err)
	}
	
	e2.metrics.MessagesTotal.WithLabelValues("control_request", "sent").Inc()
	e2.logger.WithFields(logrus.Fields{
		"node_id": nodeID,
		"request_id": request.RICRequestID,
	}).Debug("RIC control request sent")
	
	return nil
}

// CreateSubscription creates a new RIC subscription
func (e2 *E2Interface) CreateSubscription(nodeID string, subscription *RICSubscription) error {
	// Add subscription to manager
	if err := e2.subscriptionMgr.CreateSubscription(nodeID, subscription); err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	
	// Create subscription request message
	request := &RICSubscriptionRequest{
		RICRequestID:           subscription.RequestID,
		RANFunctionID:          subscription.RANFunctionID,
		RICSubscriptionDetails: subscription.SubscriptionDetails,
	}
	
	// Encode and send
	data, err := e2.asn1Codec.EncodeRICSubscriptionRequest(request)
	if err != nil {
		e2.subscriptionMgr.DeleteSubscription(subscription.RequestID)
		return fmt.Errorf("failed to encode subscription request: %w", err)
	}
	
	if err := e2.sctpServer.SendToNode(nodeID, data); err != nil {
		e2.subscriptionMgr.DeleteSubscription(subscription.RequestID)
		return fmt.Errorf("failed to send subscription request: %w", err)
	}
	
	e2.metrics.SubscriptionsTotal.WithLabelValues("created").Inc()
	return nil
}

// DeleteSubscription deletes an existing RIC subscription
func (e2 *E2Interface) DeleteSubscription(subscriptionID string) error {
	// Get subscription details
	subscription, err := e2.subscriptionMgr.GetSubscription(subscriptionID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}
	
	// Create subscription delete request
	deleteReq := &RICSubscriptionDeleteRequest{
		RICRequestID:  subscription.RequestID,
		RANFunctionID: subscription.RANFunctionID,
	}
	
	// Encode and send
	data, err := e2.asn1Codec.EncodeRICSubscriptionDeleteRequest(deleteReq)
	if err != nil {
		return fmt.Errorf("failed to encode subscription delete request: %w", err)
	}
	
	if err := e2.sctpServer.SendToNode(subscription.NodeID, data); err != nil {
		return fmt.Errorf("failed to send subscription delete request: %w", err)
	}
	
	// Remove from manager
	if err := e2.subscriptionMgr.DeleteSubscription(subscriptionID); err != nil {
		e2.logger.WithError(err).Warn("Failed to remove subscription from manager")
	}
	
	e2.metrics.SubscriptionsTotal.WithLabelValues("deleted").Inc()
	return nil
}

// handleIncomingMessage processes incoming E2AP messages
func (e2 *E2Interface) handleIncomingMessage(connectionID, nodeID string, data []byte) {
	start := time.Now()
	
	// Create message object
	msg := &E2Message{
		ConnectionID: connectionID,
		NodeID:       nodeID,
		Data:         data,
		Timestamp:    start,
	}
	
	// Try to determine message type for routing
	if messageType, err := e2.asn1Codec.GetMessageType(data); err == nil {
		msg.MessageType = messageType
	}
	
	// Queue message for processing
	select {
	case e2.messageQueue <- msg:
		e2.metrics.MessagesTotal.WithLabelValues("received", "queued").Inc()
	default:
		e2.logger.WithField("node_id", nodeID).Error("Message queue is full, dropping message")
		e2.metrics.MessagesTotal.WithLabelValues("received", "dropped").Inc()
	}
	
	// Update node last activity
	e2.nodeManager.UpdateLastActivity(nodeID, start)
	
	// Record processing time metric
	e2.metrics.MessageProcessingDuration.WithLabelValues("received").Observe(time.Since(start).Seconds())
}

// handleConnectionEvent handles SCTP connection events
func (e2 *E2Interface) handleConnectionEvent(event *ConnectionEvent) {
	switch event.Type {
	case ConnectionEstablished:
		e2.logger.WithFields(logrus.Fields{
			"connection_id": event.ConnectionID,
			"remote_addr":   event.RemoteAddr,
		}).Info("New E2 connection established")
		
		e2.metrics.ConnectionsTotal.WithLabelValues("established").Inc()
		
	case ConnectionClosed:
		e2.logger.WithFields(logrus.Fields{
			"connection_id": event.ConnectionID,
			"node_id":       event.NodeID,
		}).Info("E2 connection closed")
		
		// Clean up node if registered
		if event.NodeID != "" {
			e2.nodeManager.RemoveNode(event.NodeID)
		}
		
		e2.metrics.ConnectionsTotal.WithLabelValues("closed").Inc()
		
	case ConnectionError:
		e2.logger.WithFields(logrus.Fields{
			"connection_id": event.ConnectionID,
			"node_id":       event.NodeID,
			"error":         event.Error,
		}).Error("E2 connection error")
		
		e2.metrics.ConnectionsTotal.WithLabelValues("error").Inc()
	}
}

// messageProcessor processes messages from the queue
func (e2 *E2Interface) messageProcessor(ctx context.Context) error {
	e2.logger.Debug("Starting E2 message processor")
	
	for {
		select {
		case <-ctx.Done():
			e2.logger.Debug("Message processor shutting down")
			return ctx.Err()
			
		case msg := <-e2.messageQueue:
			if msg == nil {
				continue
			}
			
			// Submit to worker pool for processing
			task := &ProcessingTask{
				Message: msg,
				Handler: e2.processE2APMessage,
			}
			
			if err := e2.workerPool.Submit(task); err != nil {
				e2.logger.WithError(err).Error("Failed to submit message to worker pool")
				e2.metrics.MessagesTotal.WithLabelValues("processing", "worker_error").Inc()
			}
		}
	}
}

// processE2APMessage processes a single E2AP message
func (e2 *E2Interface) processE2APMessage(msg *E2Message) error {
	start := time.Now()
	defer func() {
		e2.metrics.MessageProcessingDuration.WithLabelValues("processed").Observe(time.Since(start).Seconds())
	}()
	
	// Decode E2AP message
	pdu, err := e2.asn1Codec.DecodeE2AP_PDU(msg.Data)
	if err != nil {
		e2.logger.WithError(err).WithField("node_id", msg.NodeID).Error("Failed to decode E2AP PDU")
		e2.metrics.MessagesTotal.WithLabelValues("processing", "decode_error").Inc()
		return err
	}
	
	// Route message based on type
	switch {
	case pdu.InitiatingMessage != nil:
		return e2.handleInitiatingMessage(msg, pdu.InitiatingMessage)
	case pdu.SuccessfulOutcome != nil:
		return e2.handleSuccessfulOutcome(msg, pdu.SuccessfulOutcome)
	case pdu.UnsuccessfulOutcome != nil:
		return e2.handleUnsuccessfulOutcome(msg, pdu.UnsuccessfulOutcome)
	default:
		e2.logger.WithField("node_id", msg.NodeID).Warn("Unknown E2AP PDU type")
		e2.metrics.MessagesTotal.WithLabelValues("processing", "unknown_type").Inc()
		return fmt.Errorf("unknown E2AP PDU type")
	}
}

// handleInitiatingMessage handles initiating messages (from E2 nodes)
func (e2 *E2Interface) handleInitiatingMessage(msg *E2Message, initiating *InitiatingMessage) error {
	switch initiating.ProcedureCode {
	case E2SetupRequestID:
		return e2.handleE2SetupRequest(msg, initiating.Value)
	case RICIndicationID:
		return e2.handleRICIndication(msg, initiating.Value)
	default:
		e2.logger.WithFields(logrus.Fields{
			"node_id":        msg.NodeID,
			"procedure_code": initiating.ProcedureCode,
		}).Warn("Unknown initiating message procedure code")
		e2.metrics.MessagesTotal.WithLabelValues("processing", "unknown_procedure").Inc()
		return fmt.Errorf("unknown procedure code: %d", initiating.ProcedureCode)
	}
}

// handleE2SetupRequest handles E2 Setup Request messages
func (e2 *E2Interface) handleE2SetupRequest(msg *E2Message, value interface{}) error {
	e2.logger.WithField("node_id", msg.NodeID).Debug("Processing E2 Setup Request")
	
	// Decode setup request
	setupReq, err := e2.asn1Codec.DecodeE2SetupRequest(value)
	if err != nil {
		return fmt.Errorf("failed to decode E2 setup request: %w", err)
	}
	
	// Register the E2 node
	node := &E2Node{
		NodeID:            msg.NodeID,
		ConnectionID:      msg.ConnectionID,
		GlobalE2NodeID:    setupReq.GlobalE2NodeID,
		RANFunctions:      setupReq.RANFunctions,
		Status:            NodeStatusConnected,
		ConnectedAt:       time.Now(),
		LastHeartbeat:     time.Now(),
		SupportedVersions: []string{"1.0.0"}, // O-RAN E2AP version
	}
	
	if err := e2.nodeManager.RegisterNode(node); err != nil {
		// Send setup failure
		failure := &E2SetupFailure{
			Cause: CauseRadioNetwork,
			CauseValue: CauseValueNodeRejection,
		}
		
		return e2.sendE2SetupFailure(msg.ConnectionID, failure)
	}
	
	// Send setup response
	response := &E2SetupResponse{
		GlobalRICID: GlobalRICID{
			PLMNIdentity: []byte{0x00, 0x01, 0x02}, // Example PLMN
			RICIdentity:  []byte{0x12, 0x34, 0x56, 0x78},
		},
		RANFunctionsAccepted: setupReq.RANFunctions, // Accept all functions for now
	}
	
	e2.metrics.MessagesTotal.WithLabelValues("e2_setup", "processed").Inc()
	return e2.sendE2SetupResponse(msg.ConnectionID, response)
}

// handleRICIndication handles RIC Indication messages
func (e2 *E2Interface) handleRICIndication(msg *E2Message, value interface{}) error {
	e2.logger.WithField("node_id", msg.NodeID).Debug("Processing RIC Indication")
	
	// Decode indication
	indication, err := e2.asn1Codec.DecodeRICIndication(value)
	if err != nil {
		return fmt.Errorf("failed to decode RIC indication: %w", err)
	}
	
	// Update subscription with indication data
	if err := e2.subscriptionMgr.HandleIndication(indication); err != nil {
		e2.logger.WithError(err).Warn("Failed to handle indication in subscription manager")
	}
	
	// Forward indication to interested xApps
	// This would be implemented based on subscription details
	
	e2.metrics.MessagesTotal.WithLabelValues("ric_indication", "processed").Inc()
	return nil
}

// sendE2SetupResponse sends an E2 Setup Response message
func (e2 *E2Interface) sendE2SetupResponse(connectionID string, response *E2SetupResponse) error {
	data, err := e2.asn1Codec.EncodeE2SetupResponse(response)
	if err != nil {
		return fmt.Errorf("failed to encode E2 setup response: %w", err)
	}
	
	if err := e2.sctpServer.SendToConnection(connectionID, data); err != nil {
		return fmt.Errorf("failed to send E2 setup response: %w", err)
	}
	
	e2.metrics.MessagesTotal.WithLabelValues("e2_setup_response", "sent").Inc()
	return nil
}

// sendE2SetupFailure sends an E2 Setup Failure message
func (e2 *E2Interface) sendE2SetupFailure(connectionID string, failure *E2SetupFailure) error {
	data, err := e2.asn1Codec.EncodeE2SetupFailure(failure)
	if err != nil {
		return fmt.Errorf("failed to encode E2 setup failure: %w", err)
	}
	
	if err := e2.sctpServer.SendToConnection(connectionID, data); err != nil {
		return fmt.Errorf("failed to send E2 setup failure: %w", err)
	}
	
	e2.metrics.MessagesTotal.WithLabelValues("e2_setup_failure", "sent").Inc()
	return nil
}

// healthMonitor monitors the health of the E2 interface
func (e2 *E2Interface) healthMonitor(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	e2.logger.Debug("Starting E2 health monitor")
	
	for {
		select {
		case <-ctx.Done():
			e2.logger.Debug("Health monitor shutting down")
			return ctx.Err()
			
		case <-ticker.C:
			e2.performHealthCheck()
		}
	}
}

// performHealthCheck performs periodic health checks
func (e2 *E2Interface) performHealthCheck() {
	// Get current statistics
	status := e2.GetStatus()
	
	e2.logger.WithFields(logrus.Fields{
		"connections":       status.Connections,
		"nodes":            status.Nodes,
		"subscriptions":    status.Subscriptions,
		"uptime":           status.Uptime,
	}).Debug("E2 interface health check")
	
	// Update metrics
	e2.metrics.ActiveConnections.Set(float64(status.Connections))
	e2.metrics.ActiveNodes.Set(float64(status.Nodes))
	e2.metrics.ActiveSubscriptions.Set(float64(status.Subscriptions))
	
	// Check for stale connections
	staleNodes := e2.nodeManager.GetStaleNodes(5 * time.Minute)
	if len(staleNodes) > 0 {
		e2.logger.WithField("stale_nodes", len(staleNodes)).Warn("Found stale E2 nodes")
		
		for _, nodeID := range staleNodes {
			e2.nodeManager.MarkNodeStale(nodeID)
		}
	}
}

// subscriptionWorker manages subscription lifecycle
func (e2 *E2Interface) subscriptionWorker(ctx context.Context) error {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	e2.logger.Debug("Starting subscription worker")
	
	for {
		select {
		case <-ctx.Done():
			e2.logger.Debug("Subscription worker shutting down")
			return ctx.Err()
			
		case <-ticker.C:
			e2.cleanupExpiredSubscriptions()
		}
	}
}

// cleanupExpiredSubscriptions removes expired subscriptions
func (e2 *E2Interface) cleanupExpiredSubscriptions() {
	expired := e2.subscriptionMgr.GetExpiredSubscriptions(24 * time.Hour)
	
	for _, subscriptionID := range expired {
		e2.logger.WithField("subscription_id", subscriptionID).Info("Cleaning up expired subscription")
		
		if err := e2.subscriptionMgr.DeleteSubscription(subscriptionID); err != nil {
			e2.logger.WithError(err).Error("Failed to delete expired subscription")
		}
	}
	
	if len(expired) > 0 {
		e2.metrics.SubscriptionsTotal.WithLabelValues("expired").Add(float64(len(expired)))
	}
}

// nodeMaintenanceWorker performs node maintenance tasks
func (e2 *E2Interface) nodeMaintenanceWorker(ctx context.Context) error {
	ticker := time.NewTicker(120 * time.Second)
	defer ticker.Stop()
	
	e2.logger.Debug("Starting node maintenance worker")
	
	for {
		select {
		case <-ctx.Done():
			e2.logger.Debug("Node maintenance worker shutting down")
			return ctx.Err()
			
		case <-ticker.C:
			e2.performNodeMaintenance()
		}
	}
}

// performNodeMaintenance performs node maintenance tasks
func (e2 *E2Interface) performNodeMaintenance() {
	// Clean up disconnected nodes
	disconnected := e2.nodeManager.GetDisconnectedNodes(10 * time.Minute)
	
	for _, nodeID := range disconnected {
		e2.logger.WithField("node_id", nodeID).Info("Removing disconnected node")
		
		// Clean up associated subscriptions
		subscriptions := e2.subscriptionMgr.GetSubscriptionsByNode(nodeID)
		for _, subID := range subscriptions {
			e2.subscriptionMgr.DeleteSubscription(subID)
		}
		
		// Remove node
		e2.nodeManager.RemoveNode(nodeID)
	}
}

// handleSuccessfulOutcome handles successful outcome messages
func (e2 *E2Interface) handleSuccessfulOutcome(msg *E2Message, outcome *SuccessfulOutcome) error {
	switch outcome.ProcedureCode {
	case RICSubscriptionRequestID:
		return e2.handleRICSubscriptionResponse(msg, outcome.Value)
	case RICControlRequestID:
		return e2.handleRICControlAck(msg, outcome.Value)
	default:
		e2.logger.WithFields(logrus.Fields{
			"node_id":        msg.NodeID,
			"procedure_code": outcome.ProcedureCode,
		}).Warn("Unknown successful outcome procedure code")
		return nil
	}
}

// handleUnsuccessfulOutcome handles unsuccessful outcome messages
func (e2 *E2Interface) handleUnsuccessfulOutcome(msg *E2Message, outcome *UnsuccessfulOutcome) error {
	switch outcome.ProcedureCode {
	case RICSubscriptionRequestID:
		return e2.handleRICSubscriptionFailure(msg, outcome.Value)
	case RICControlRequestID:
		return e2.handleRICControlFailure(msg, outcome.Value)
	default:
		e2.logger.WithFields(logrus.Fields{
			"node_id":        msg.NodeID,
			"procedure_code": outcome.ProcedureCode,
		}).Warn("Unknown unsuccessful outcome procedure code")
		return nil
	}
}

// handleRICSubscriptionResponse handles RIC Subscription Response messages
func (e2 *E2Interface) handleRICSubscriptionResponse(msg *E2Message, value interface{}) error {
	e2.logger.WithField("node_id", msg.NodeID).Debug("Processing RIC Subscription Response")
	
	response, err := e2.asn1Codec.DecodeRICSubscriptionResponse(value)
	if err != nil {
		return fmt.Errorf("failed to decode RIC subscription response: %w", err)
	}
	
	// Update subscription status
	if err := e2.subscriptionMgr.UpdateSubscriptionStatus(response.RICRequestID, SubscriptionActive); err != nil {
		e2.logger.WithError(err).Warn("Failed to update subscription status")
	}
	
	e2.metrics.MessagesTotal.WithLabelValues("subscription_response", "processed").Inc()
	return nil
}

// handleRICSubscriptionFailure handles RIC Subscription Failure messages
func (e2 *E2Interface) handleRICSubscriptionFailure(msg *E2Message, value interface{}) error {
	e2.logger.WithField("node_id", msg.NodeID).Warn("Processing RIC Subscription Failure")
	
	failure, err := e2.asn1Codec.DecodeRICSubscriptionFailure(value)
	if err != nil {
		return fmt.Errorf("failed to decode RIC subscription failure: %w", err)
	}
	
	// Mark subscription as failed
	if err := e2.subscriptionMgr.UpdateSubscriptionStatus(failure.RICRequestID, SubscriptionFailed); err != nil {
		e2.logger.WithError(err).Warn("Failed to update subscription status")
	}
	
	e2.metrics.MessagesTotal.WithLabelValues("subscription_failure", "processed").Inc()
	return nil
}

// handleRICControlAck handles RIC Control Acknowledge messages
func (e2 *E2Interface) handleRICControlAck(msg *E2Message, value interface{}) error {
	e2.logger.WithField("node_id", msg.NodeID).Debug("Processing RIC Control Ack")
	
	ack, err := e2.asn1Codec.DecodeRICControlAck(value)
	if err != nil {
		return fmt.Errorf("failed to decode RIC control ack: %w", err)
	}
	
	// Handle control acknowledgment
	// This would typically involve notifying the requesting xApp
	
	e2.metrics.MessagesTotal.WithLabelValues("control_ack", "processed").Inc()
	return nil
}

// handleRICControlFailure handles RIC Control Failure messages
func (e2 *E2Interface) handleRICControlFailure(msg *E2Message, value interface{}) error {
	e2.logger.WithField("node_id", msg.NodeID).Warn("Processing RIC Control Failure")
	
	failure, err := e2.asn1Codec.DecodeRICControlFailure(value)
	if err != nil {
		return fmt.Errorf("failed to decode RIC control failure: %w", err)
	}
	
	// Handle control failure
	// This would typically involve notifying the requesting xApp
	
	e2.metrics.MessagesTotal.WithLabelValues("control_failure", "processed").Inc()
	return nil
}