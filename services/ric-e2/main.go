package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hctsai1006/near-rt-ric/pkg/e2"
	"github.com/hctsai1006/near-rt-ric/pkg/e2/e2common"
	"github.com/sirupsen/logrus"
)

const (
	version = "1.0.0"
	appName = "O-RAN Near-RT RIC"
)

var (
	listenAddr  = flag.String("listen-addr", "0.0.0.0", "Listen address for E2 interface")
	listenPort  = flag.Int("listen-port", 36421, "Listen port for E2 interface")
	maxNodes    = flag.Int("max-nodes", 100, "Maximum number of E2 nodes")
	logLevel    = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
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
		"version":     version,
		"listen_addr": *listenAddr,
		"listen_port": *listenPort,
		"max_nodes":   *maxNodes,
	}).Info("Starting O-RAN Near-RT RIC")

	// Create E2 interface configuration
	config := &e2common.E2InterfaceConfig{
		ListenAddress:     *listenAddr,
		ListenPort:        *listenPort,
		MaxNodes:          *maxNodes,
		HeartbeatInterval: 30 * time.Second,
		ConnectionTimeout: 60 * time.Second,
	}

	// Create and start E2 interface
	e2Interface := e2.NewE2Interface(config)

	if err := e2Interface.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start E2 interface")
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("O-RAN Near-RT RIC started successfully")

	// Start status reporting goroutine
	go statusReporter(e2Interface, logger)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.WithField("signal", sig.String()).Info("Received shutdown signal")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	shutdownChan := make(chan error, 1)
	go func() {
		shutdownChan <- e2Interface.Stop()
	}()

	select {
	case err := <-shutdownChan:
		if err != nil {
			logger.WithError(err).Error("Error during shutdown")
			os.Exit(1)
		}
		logger.Info("O-RAN Near-RT RIC shutdown completed successfully")
	case <-shutdownCtx.Done():
		logger.Error("Shutdown timeout exceeded")
		os.Exit(1)
	}
}

// statusReporter periodically reports the status of the RIC
func statusReporter(e2Interface e2.E2Interface, logger *logrus.Logger) {
	ticker := time.NewTicker(5 * time.Minute) // Report every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status := e2Interface.GetStatus()
			nodes := e2Interface.GetOperationalNodes()

			logger.WithFields(logrus.Fields{
				"operational_nodes":  len(nodes),
				"total_connections":  status["connection_statistics"].(map[string]interface{})["total_connections"],
				"active_connections": status["connection_statistics"].(map[string]interface{})["active_connections"],
			}).Info("O-RAN Near-RT RIC status report")

			// Log details about each operational node
			for _, node := range nodes {
				logger.WithFields(logrus.Fields{
					"node_id":        node.ID,
					"node_type":      node.Type,
					"address":        node.Address,
					"functions":      len(node.FunctionList),
					"setup_complete": node.SetupComplete,
					"last_heartbeat": node.LastHeartbeat.Format(time.RFC3339),
				}).Debug("Operational E2 node details")
			}
		}
	}
}
