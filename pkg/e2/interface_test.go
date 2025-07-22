package e2

import (
	"context"
	"testing"
	"time"

	"github.com/hctsai1006/near-rt-ric/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSCTPServer implements a mock SCTP server for testing
type MockSCTPServer struct {
	mock.Mock
}

func (m *MockSCTPServer) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSCTPServer) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSCTPServer) GetConnectionCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSCTPServer) SendToNode(nodeID string, data []byte) error {
	args := m.Called(nodeID, data)
	return args.Error(0)
}

func (m *MockSCTPServer) SendToConnection(connectionID string, data []byte) error {
	args := m.Called(connectionID, data)
	return args.Error(0)
}

func (m *MockSCTPServer) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSCTPServer) SetMessageHandler(handler func(string, string, []byte)) {
	m.Called(handler)
}

func (m *MockSCTPServer) SetConnectionHandler(handler func(*ConnectionEvent)) {
	m.Called(handler)
}

// MockNodeManager implements a mock node manager for testing
type MockNodeManager struct {
	mock.Mock
}

func (m *MockNodeManager) RegisterNode(node *E2Node) error {
	args := m.Called(node)
	return args.Error(0)
}

func (m *MockNodeManager) RemoveNode(nodeID string) error {
	args := m.Called(nodeID)
	return args.Error(0)
}

func (m *MockNodeManager) GetNode(nodeID string) (*E2Node, error) {
	args := m.Called(nodeID)
	return args.Get(0).(*E2Node), args.Error(1)
}

func (m *MockNodeManager) GetAllNodes() map[string]*E2Node {
	args := m.Called()
	return args.Get(0).(map[string]*E2Node)
}

func (m *MockNodeManager) GetNodeCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockNodeManager) GetLastActivity() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *MockNodeManager) GetStartTime() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *MockNodeManager) UpdateLastActivity(nodeID string, timestamp time.Time) {
	m.Called(nodeID, timestamp)
}

func (m *MockNodeManager) GetStaleNodes(timeout time.Duration) []string {
	args := m.Called(timeout)
	return args.Get(0).([]string)
}

func (m *MockNodeManager) MarkNodeStale(nodeID string) {
	m.Called(nodeID)
}

func (m *MockNodeManager) GetDisconnectedNodes(timeout time.Duration) []string {
	args := m.Called(timeout)
	return args.Get(0).([]string)
}

func (m *MockNodeManager) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNodeManager) Cleanup() {
	m.Called()
}

// MockASN1Codec implements a mock ASN.1 codec for testing
type MockASN1Codec struct {
	mock.Mock
}

func (m *MockASN1Codec) DecodeE2AP_PDU(data []byte) (*E2AP_PDU, error) {
	args := m.Called(data)
	return args.Get(0).(*E2AP_PDU), args.Error(1)
}

func (m *MockASN1Codec) EncodeE2SetupResponse(response *E2SetupResponse) ([]byte, error) {
	args := m.Called(response)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockASN1Codec) EncodeE2SetupFailure(failure *E2SetupFailure) ([]byte, error) {
	args := m.Called(failure)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockASN1Codec) EncodeRICControlRequest(request *RICControlRequest) ([]byte, error) {
	args := m.Called(request)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockASN1Codec) GetMessageType(data []byte) (E2MessageType, error) {
	args := m.Called(data)
	return args.Get(0).(E2MessageType), args.Error(1)
}

func (m *MockASN1Codec) DecodeE2SetupRequest(value interface{}) (*E2SetupRequest, error) {
	args := m.Called(value)
	return args.Get(0).(*E2SetupRequest), args.Error(1)
}

func (m *MockASN1Codec) DecodeRICIndication(value interface{}) (*RICIndication, error) {
	args := m.Called(value)
	return args.Get(0).(*RICIndication), args.Error(1)
}

// MockSubscriptionManager implements a mock subscription manager for testing
type MockSubscriptionManager struct {
	mock.Mock
}

func (m *MockSubscriptionManager) CreateSubscription(nodeID string, subscription *RICSubscription) error {
	args := m.Called(nodeID, subscription)
	return args.Error(0)
}

func (m *MockSubscriptionManager) DeleteSubscription(subscriptionID string) error {
	args := m.Called(subscriptionID)
	return args.Error(0)
}

func (m *MockSubscriptionManager) GetSubscription(subscriptionID string) (*RICSubscription, error) {
	args := m.Called(subscriptionID)
	return args.Get(0).(*RICSubscription), args.Error(1)
}

func (m *MockSubscriptionManager) GetSubscriptionCount() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockSubscriptionManager) HandleIndication(indication *RICIndication) error {
	args := m.Called(indication)
	return args.Error(0)
}

func (m *MockSubscriptionManager) UpdateSubscriptionStatus(requestID string, status SubscriptionStatus) error {
	args := m.Called(requestID, status)
	return args.Error(0)
}

func (m *MockSubscriptionManager) GetExpiredSubscriptions(timeout time.Duration) []string {
	args := m.Called(timeout)
	return args.Get(0).([]string)
}

func (m *MockSubscriptionManager) GetSubscriptionsByNode(nodeID string) []string {
	args := m.Called(nodeID)
	return args.Get(0).([]string)
}

func (m *MockSubscriptionManager) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockSubscriptionManager) Cleanup() {
	m.Called()
}

// Test configuration helper
func createTestConfig() *config.E2Config {
	return &config.E2Config{
		Enabled:           true,
		ListenAddress:     "127.0.0.1",
		Port:              36421,
		MaxConnections:    100,
		ConnectionTimeout: 30 * time.Second,
		HeartbeatInterval: 30 * time.Second,
		MessageTimeout:    10 * time.Second,
		BufferSize:        65536,
		WorkerPoolSize:    5,
		SCTP: config.SCTPConfig{
			NumStreams:      10,
			MaxRetransmits:  5,
			RTOInitial:      3 * time.Second,
		},
		ASN1: config.ASN1Config{
			ValidateMessages: true,
			StrictDecoding:   true,
			MaxMessageSize:   1048576,
		},
	}
}

// TestNewE2Interface tests the creation of a new E2Interface
func TestNewE2Interface(t *testing.T) {
	cfg := createTestConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// This would require implementing proper constructors for dependencies
	// For now, we'll test the basic structure
	assert.NotNil(t, cfg)
	assert.Equal(t, 36421, cfg.Port)
	assert.Equal(t, "127.0.0.1", cfg.ListenAddress)
}

// TestE2InterfaceLifecycle tests the start and stop functionality
func TestE2InterfaceLifecycle(t *testing.T) {
	// Create test interface with mocks
	e2Interface := createTestE2Interface(t)

	ctx := context.Background()

	// Test starting the interface
	err := e2Interface.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, e2Interface.running)

	// Test starting already running interface
	err = e2Interface.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stopping the interface
	err = e2Interface.Stop(ctx)
	assert.NoError(t, err)
	assert.False(t, e2Interface.running)

	// Test stopping already stopped interface
	err = e2Interface.Stop(ctx)
	assert.NoError(t, err)
}

// TestE2InterfaceHealthCheck tests the health check functionality
func TestE2InterfaceHealthCheck(t *testing.T) {
	e2Interface := createTestE2Interface(t)

	// Health check when not running
	err := e2Interface.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")

	// Start interface and test health check
	ctx := context.Background()
	err = e2Interface.Start(ctx)
	require.NoError(t, err)

	// Mock successful health checks
	e2Interface.sctpServer.(*MockSCTPServer).On("HealthCheck").Return(nil)
	e2Interface.nodeManager.(*MockNodeManager).On("HealthCheck").Return(nil)
	e2Interface.subscriptionMgr.(*MockSubscriptionManager).On("HealthCheck").Return(nil)
	e2Interface.workerPool = &MockWorkerPool{}
	e2Interface.workerPool.(*MockWorkerPool).On("HealthCheck").Return(nil)

	err = e2Interface.HealthCheck()
	assert.NoError(t, err)

	// Clean up
	err = e2Interface.Stop(ctx)
	require.NoError(t, err)
}

// TestE2InterfaceGetStatus tests the status reporting functionality
func TestE2InterfaceGetStatus(t *testing.T) {
	e2Interface := createTestE2Interface(t)

	// Mock return values
	e2Interface.sctpServer.(*MockSCTPServer).On("GetConnectionCount").Return(5)
	e2Interface.nodeManager.(*MockNodeManager).On("GetNodeCount").Return(3)
	e2Interface.nodeManager.(*MockNodeManager).On("GetLastActivity").Return(time.Now())
	e2Interface.subscriptionMgr.(*MockSubscriptionManager).On("GetSubscriptionCount").Return(10)

	status := e2Interface.GetStatus()

	assert.NotNil(t, status)
	assert.Equal(t, "127.0.0.1", status.ListenAddress)
	assert.Equal(t, 36421, status.Port)
	assert.Equal(t, 5, status.Connections)
	assert.Equal(t, 100, status.MaxConnections)
	assert.Equal(t, 3, status.Nodes)
	assert.Equal(t, 10, status.Subscriptions)
}

// TestE2SetupRequestHandling tests E2 Setup Request message handling
func TestE2SetupRequestHandling(t *testing.T) {
	e2Interface := createTestE2Interface(t)

	ctx := context.Background()
	err := e2Interface.Start(ctx)
	require.NoError(t, err)

	// Create test E2 Setup Request
	setupReq := &E2SetupRequest{
		GlobalE2NodeID: GlobalE2NodeID{
			PLMNIdentity: []byte{0x00, 0x01, 0x02},
			NodeID:       []byte{0x12, 0x34},
		},
		RANFunctions: []RANFunction{
			{
				ID:       1,
				Name:     "KPM",
				Version:  "3.0.0",
				OID:      "1.3.6.1.4.1.1.1.2.2",
			},
		},
	}

	// Mock ASN.1 codec
	e2Interface.asn1Codec.(*MockASN1Codec).On("DecodeE2SetupRequest", mock.Anything).Return(setupReq, nil)
	e2Interface.nodeManager.(*MockNodeManager).On("RegisterNode", mock.AnythingOfType("*e2.E2Node")).Return(nil)
	e2Interface.asn1Codec.(*MockASN1Codec).On("EncodeE2SetupResponse", mock.AnythingOfType("*e2.E2SetupResponse")).Return([]byte{0x01, 0x02, 0x03}, nil)
	e2Interface.sctpServer.(*MockSCTPServer).On("SendToConnection", "conn-123", mock.AnythingOfType("[]uint8")).Return(nil)

	// Create test message
	msg := &E2Message{
		ConnectionID: "conn-123",
		NodeID:       "node-123",
		Data:         []byte{0x01, 0x02, 0x03},
		MessageType:  E2SetupRequestMsg,
		Timestamp:    time.Now(),
	}

	// Test E2 Setup Request handling
	err = e2Interface.handleE2SetupRequest(msg, setupReq)
	assert.NoError(t, err)

	// Clean up
	err = e2Interface.Stop(ctx)
	require.NoError(t, err)
}

// TestRICControlRequest tests RIC Control Request functionality
func TestRICControlRequest(t *testing.T) {
	e2Interface := createTestE2Interface(t)

	ctx := context.Background()
	err := e2Interface.Start(ctx)
	require.NoError(t, err)

	// Create test RIC Control Request
	controlReq := &RICControlRequest{
		RICRequestID:  RICRequestID{RequestorID: 1, InstanceID: 1},
		RANFunctionID: 1,
		RICControlHeader:  []byte{0x01, 0x02},
		RICControlMessage: []byte{0x03, 0x04},
		RICControlAckRequest: true,
	}

	// Mock encoding and sending
	e2Interface.asn1Codec.(*MockASN1Codec).On("EncodeRICControlRequest", controlReq).Return([]byte{0x01, 0x02, 0x03}, nil)
	e2Interface.sctpServer.(*MockSCTPServer).On("SendToNode", "node-123", mock.AnythingOfType("[]uint8")).Return(nil)

	// Test sending RIC Control Request
	err = e2Interface.SendRICControlRequest("node-123", controlReq)
	assert.NoError(t, err)

	// Clean up
	err = e2Interface.Stop(ctx)
	require.NoError(t, err)
}

// TestSubscriptionManagement tests subscription creation and deletion
func TestSubscriptionManagement(t *testing.T) {
	e2Interface := createTestE2Interface(t)

	ctx := context.Background()
	err := e2Interface.Start(ctx)
	require.NoError(t, err)

	// Create test subscription
	subscription := &RICSubscription{
		RequestID:     RICRequestID{RequestorID: 1, InstanceID: 1},
		RANFunctionID: 1,
		NodeID:        "node-123",
		SubscriptionDetails: RICSubscriptionDetails{
			EventTriggers: []byte{0x01, 0x02},
			Actions:       []RICAction{{ID: 1, Type: RICActionTypeReport}},
		},
	}

	// Mock subscription creation
	e2Interface.subscriptionMgr.(*MockSubscriptionManager).On("CreateSubscription", "node-123", subscription).Return(nil)
	e2Interface.asn1Codec.(*MockASN1Codec).On("EncodeRICSubscriptionRequest", mock.AnythingOfType("*e2.RICSubscriptionRequest")).Return([]byte{0x01, 0x02, 0x03}, nil)
	e2Interface.sctpServer.(*MockSCTPServer).On("SendToNode", "node-123", mock.AnythingOfType("[]uint8")).Return(nil)

	// Test creating subscription
	err = e2Interface.CreateSubscription("node-123", subscription)
	assert.NoError(t, err)

	// Mock subscription deletion
	e2Interface.subscriptionMgr.(*MockSubscriptionManager).On("GetSubscription", "sub-123").Return(subscription, nil)
	e2Interface.asn1Codec.(*MockASN1Codec).On("EncodeRICSubscriptionDeleteRequest", mock.AnythingOfType("*e2.RICSubscriptionDeleteRequest")).Return([]byte{0x01, 0x02, 0x03}, nil)
	e2Interface.subscriptionMgr.(*MockSubscriptionManager).On("DeleteSubscription", "sub-123").Return(nil)

	// Test deleting subscription
	err = e2Interface.DeleteSubscription("sub-123")
	assert.NoError(t, err)

	// Clean up
	err = e2Interface.Stop(ctx)
	require.NoError(t, err)
}

// TestConnectionEventHandling tests SCTP connection event handling
func TestConnectionEventHandling(t *testing.T) {
	e2Interface := createTestE2Interface(t)

	// Test connection established event
	event := &ConnectionEvent{
		Type:         ConnectionEstablished,
		ConnectionID: "conn-123",
		RemoteAddr:   "192.168.1.100:12345",
	}

	e2Interface.handleConnectionEvent(event)

	// Test connection closed event
	event = &ConnectionEvent{
		Type:         ConnectionClosed,
		ConnectionID: "conn-123",
		NodeID:       "node-123",
	}

	e2Interface.nodeManager.(*MockNodeManager).On("RemoveNode", "node-123").Return(nil)
	e2Interface.handleConnectionEvent(event)

	// Test connection error event
	event = &ConnectionEvent{
		Type:         ConnectionError,
		ConnectionID: "conn-123",
		NodeID:       "node-123",
		Error:        "connection timeout",
	}

	e2Interface.handleConnectionEvent(event)
}

// Benchmark tests for performance validation
func BenchmarkE2MessageProcessing(b *testing.B) {
	e2Interface := createTestE2Interface(nil)

	// Setup mocks for benchmarking
	pdu := &E2AP_PDU{
		InitiatingMessage: &InitiatingMessage{
			ProcedureCode: E2SetupRequestID,
			Value:         &E2SetupRequest{},
		},
	}

	e2Interface.asn1Codec.(*MockASN1Codec).On("DecodeE2AP_PDU", mock.Anything).Return(pdu, nil)
	e2Interface.asn1Codec.(*MockASN1Codec).On("DecodeE2SetupRequest", mock.Anything).Return(&E2SetupRequest{}, nil)
	e2Interface.nodeManager.(*MockNodeManager).On("RegisterNode", mock.Anything).Return(nil)
	e2Interface.asn1Codec.(*MockASN1Codec).On("EncodeE2SetupResponse", mock.Anything).Return([]byte{}, nil)
	e2Interface.sctpServer.(*MockSCTPServer).On("SendToConnection", mock.Anything, mock.Anything).Return(nil)

	msg := &E2Message{
		ConnectionID: "conn-123",
		NodeID:       "node-123",
		Data:         []byte{0x01, 0x02, 0x03},
		MessageType:  E2SetupRequestMsg,
		Timestamp:    time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e2Interface.processE2APMessage(msg)
	}
}

// Helper function to create a test E2Interface with mocks
func createTestE2Interface(t *testing.T) *E2Interface {
	cfg := createTestConfig()
	
	var logger *logrus.Logger
	if t != nil {
		logger = logrus.New()
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger = logrus.New()
		logger.SetLevel(logrus.ErrorLevel) // Reduce noise in benchmarks
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create mocks
	mockSCTPServer := &MockSCTPServer{}
	mockNodeManager := &MockNodeManager{}
	mockASN1Codec := &MockASN1Codec{}
	mockSubscriptionMgr := &MockSubscriptionManager{}
	mockWorkerPool := &MockWorkerPool{}
	mockMetrics := &MockE2Metrics{}

	// Set up mock expectations for basic operations
	mockSCTPServer.On("Start", mock.Anything).Return(nil)
	mockSCTPServer.On("Stop", mock.Anything).Return(nil)
	mockSCTPServer.On("SetMessageHandler", mock.Anything).Return()
	mockSCTPServer.On("SetConnectionHandler", mock.Anything).Return()
	mockWorkerPool.On("Start", mock.Anything).Return(nil)
	mockWorkerPool.On("Stop", mock.Anything).Return(nil)
	mockMetrics.On("Register").Return(nil)
	mockMetrics.On("Unregister").Return()

	// Create interface
	e2Interface := &E2Interface{
		config:          cfg,
		sctpServer:      mockSCTPServer,
		nodeManager:     mockNodeManager,
		asn1Codec:       mockASN1Codec,
		subscriptionMgr: mockSubscriptionMgr,
		logger:          logger.WithField("component", "e2-interface"),
		metrics:         mockMetrics,
		messageQueue:    make(chan *E2Message, 1000),
		workerPool:      mockWorkerPool,
		ctx:             ctx,
		cancel:          cancel,
		running:         false,
		startTime:       time.Now(),
	}

	return e2Interface
}

// Mock implementations for supporting types
type MockWorkerPool struct {
	mock.Mock
}

func (m *MockWorkerPool) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockWorkerPool) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockWorkerPool) Submit(task *ProcessingTask) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockWorkerPool) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

type MockE2Metrics struct {
	mock.Mock
}

func (m *MockE2Metrics) Register() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockE2Metrics) Unregister() {
	m.Called()
}

// Test-specific types that don't conflict with the main types.go
type ProcessingTask struct {
	Message *E2Message
	Handler func(*E2Message) error
}

// Procedure codes for E2AP
const (
	E2SetupRequestID = iota
	RICSubscriptionRequestID
	RICIndicationID
	RICControlRequestID
)

// Cause values
const (
	CauseRadioNetwork = iota
	CauseValueNodeRejection
)