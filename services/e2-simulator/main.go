package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hctsai1006/near-rt-ric/pkg/e2/asn1"
	"github.com/hctsai1006/near-rt-ric/pkg/e2/e2common"
	"github.com/sirupsen/logrus"
)

const (
	version = "1.0.0"
	appName = "O-RAN E2 Node Simulator"
)

var (
	ricAddr     = flag.String("ric-addr", "127.0.0.1", "RIC address to connect to")
	ricPort     = flag.Int("ric-port", 36421, "RIC port to connect to")
	nodeID      = flag.String("node-id", "gnb_001", "E2 node identifier")
	nodeType    = flag.String("node-type", "gnb", "E2 node type (gnb, enb, cu, du)")
	plmnID      = flag.String("plmn-id", "310410", "PLMN identifier")
	logLevel    = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	interval    = flag.Duration("report-interval", 30*time.Second, "Reporting interval for indications")
	showVersion = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	// Show version if requested
	if *showVersion {
		fmt.Printf("%s version %s\n", appName, version)
		os.Exit(0)
	}

	// Configure logging
	logger := logrus.New()
	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logger.WithError(err).Fatal("Invalid log level")
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	logger.WithFields(logrus.Fields{
		"version":   version,
		"ric_addr":  *ricAddr,
		"ric_port":  *ricPort,
		"node_id":   *nodeID,
		"node_type": *nodeType,
		"plmn_id":   *plmnID,
	}).Info("Starting O-RAN E2 Node Simulator")

	// Convert node type string to enum
	var nodeTypeEnum e2common.E2NodeType
	switch *nodeType {
	case "gnb":
		nodeTypeEnum = e2common.E2NodeTypeGNB
	case "enb":
		nodeTypeEnum = e2common.E2NodeTypeENB
	case "cu":
		nodeTypeEnum = e2common.E2NodeTypeCU
	case "du":
		nodeTypeEnum = e2common.E2NodeTypeDU
	default:
		logger.WithField("node_type", *nodeType).Fatal("Invalid node type")
	}

	// Create E2 simulator
	simulator := NewE2Simulator(&E2SimulatorConfig{
		NodeID:         *nodeID,
		NodeType:       nodeTypeEnum,
		PLMNIdentity:   []byte(*plmnID),
		RICAddress:     *ricAddr,
		RICPort:        *ricPort,
		ReportInterval: *interval,
		Logger:         logger,
	})

	// Start simulator
	if err := simulator.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start E2 simulator")
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("O-RAN E2 Node Simulator started successfully")

	// Wait for shutdown signal
	sig := <-sigChan
	logger.WithField("signal", sig.String()).Info("Received shutdown signal")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	shutdownChan := make(chan error, 1)
	go func() {
		shutdownChan <- simulator.Stop()
	}()

	select {
	case err := <-shutdownChan:
		if err != nil {
			logger.WithError(err).Error("Error during shutdown")
			os.Exit(1)
		}
		logger.Info("O-RAN E2 Node Simulator shutdown completed successfully")
	case <-shutdownCtx.Done():
		logger.Error("Shutdown timeout exceeded")
		os.Exit(1)
	}
}

// E2SimulatorConfig holds configuration for the E2 simulator
type E2SimulatorConfig struct {
	NodeID         string
	NodeType       e2common.E2NodeType
	PLMNIdentity   []byte
	RICAddress     string
	RICPort        int
	ReportInterval time.Duration
	Logger         *logrus.Logger
}

// E2Simulator simulates an E2 node
type E2Simulator struct {
	config   *E2SimulatorConfig
	sctpConn *e2common.SCTPConnection
	encoder  *asn1.ASN1Encoder
	logger   *logrus.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	running  bool
}

// NewE2Simulator creates a new E2 simulator
func NewE2Simulator(config *E2SimulatorConfig) *E2Simulator {
	ctx, cancel := context.WithCancel(context.Background())

	return &E2Simulator{
		config:  config,
		encoder: asn1.NewASN1Encoder(),
		logger:  config.Logger,
		ctx:     ctx,
		cancel:  cancel,
		running: false,
	}
}

// Start starts the E2 simulator
func (sim *E2Simulator) Start() error {
	if sim.running {
		return fmt.Errorf("E2 simulator is already running")
	}

	sim.logger.Info("Connecting to Near-RT RIC")

	// Create SCTP manager for outgoing connection
	sctpConfig := &e2common.SCTPConfig{
		ListenAddress:     "",
		ListenPort:        0,
		ConnectTimeout:    30 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      10 * time.Second,
		KeepaliveInterval: 30 * time.Second,
	}

	sctpManager := e2common.NewSCTPManager(sctpConfig)

	// Connect to RIC
	if err := sctpManager.Connect(sim.config.NodeID, sim.config.RICAddress, sim.config.RICPort); err != nil {
		return fmt.Errorf("failed to connect to RIC: %w", err)
	}

	// Send E2 Setup Request
	if err := sim.sendE2SetupRequest(sctpManager); err != nil {
		return fmt.Errorf("failed to send E2 setup request: %w", err)
	}

	// Start periodic reporting
	go sim.periodicReporting(sctpManager)

	sim.running = true
	sim.logger.Info("E2 simulator started successfully")

	return nil
}

// Stop stops the E2 simulator
func (sim *E2Simulator) Stop() error {
	if !sim.running {
		return nil
	}

	sim.logger.Info("Stopping E2 simulator")
	sim.cancel()
	sim.running = false
	sim.logger.Info("E2 simulator stopped successfully")

	return nil
}

// sendE2SetupRequest sends an E2 setup request to the RIC
func (sim *E2Simulator) sendE2SetupRequest(sctpManager *e2common.SCTPManager) error {
	setupReq := &asn1.E2SetupRequest{
		TransactionID: 1,
		GlobalE2NodeID: e2common.GlobalE2NodeID{
			PLMNIdentity: sim.config.PLMNIdentity,
			E2NodeType:   sim.config.NodeType,
			NodeIdentity: []byte(sim.config.NodeID),
		},
		RANFunctions: []e2common.RANFunction{
			{
				RANFunctionID:         1,
				RANFunctionDefinition: []byte("E2SM-KPM-v01.00"),
				RANFunctionRevision:   1,
			},
		},
	}

	data, err := sim.encoder.EncodeE2SetupRequest(setupReq)
	if err != nil {
		return fmt.Errorf("failed to encode E2 setup request: %w", err)
	}

	if err := sctpManager.SendToNode(sim.config.NodeID, data); err != nil {
		return fmt.Errorf("failed to send E2 setup request: %w", err)
	}

	sim.logger.Info("E2 Setup Request sent successfully")
	return nil
}

// periodicReporting sends periodic RIC indications
func (sim *E2Simulator) periodicReporting(sctpManager *e2common.SCTPManager) {
	ticker := time.NewTicker(sim.config.ReportInterval)
	defer ticker.Stop()

	indicationSN := int64(1)

	for {
		select {
		case <-sim.ctx.Done():
			return
		case <-ticker.C:
			if err := sim.sendRICIndication(sctpManager, indicationSN); err != nil {
				sim.logger.WithError(err).Error("Failed to send RIC indication")
			} else {
				sim.logger.WithField("sn", indicationSN).Debug("RIC indication sent successfully")
			}
			indicationSN++
		}
	}
}

// sendRICIndication sends a RIC indication message
func (sim *E2Simulator) sendRICIndication(sctpManager *e2common.SCTPManager, sn int64) error {
	// Create sample KPM data
	kmpData := fmt.Sprintf(`{
		"timestamp": "%s",
		"node_id": "%s",
		"measurements": [
			{"name": "DRB.RlcSduDelayDl", "value": %d},
			{"name": "DRB.UEThpDl", "value": %d},
			{"name": "RRU.PrbTotDl", "value": %d}
		]
	}`, time.Now().Format(time.RFC3339), sim.config.NodeID, sn%100+50, sn%1000+5000, sn%50+25)

	// This is a simplified indication - in a real implementation,
	// this would be properly encoded according to E2SM-KPM specification
	sim.logger.WithFields(logrus.Fields{
		"sn":       sn,
		"node_id":  sim.config.NodeID,
		"data_len": len(kmpData),
	}).Debug("Sending RIC indication")

	// For now, just log the indication data
	// In a real implementation, this would be properly encoded and sent
	return nil
}
