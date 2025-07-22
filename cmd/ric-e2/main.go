package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hctsai1006/near-rt-ric/pkg/config"
	"github.com/hctsai1006/near-rt-ric/pkg/e2"
	"github.com/hctsai1006/near-rt-ric/pkg/logging"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("E2")
	if err != nil {
		logrus.Fatalf("failed to load E2 configuration: %v", err)
	}

	// Initialize logger
	logger := logging.NewLogger(cfg.LogLevel)
	logger.Info("starting RIC E2 service")

	// Initialize E2 interface
		e2Interface := e2.NewE2Interface(&e2.Config{
		ListenAddress:     cfg.Server.Address,
		ListenPort:        cfg.Server.Port,
		MaxNodes:          100, // Default value, can be from config
		HeartbeatInterval: 30 * time.Second, // Default value, can be from config
		ConnectionTimeout: 60 * time.Second, // Default value, can be from config
	})

	// Start the E2 interface
	if err := e2Interface.Start(); err != nil {
		logger.Fatalf("failed to start E2 interface: %v", err)
	}

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down E2 service")

	// Stop the E2 interface
	if err := e2Interface.Stop(); err != nil {
		logger.Errorf("error while stopping E2 interface: %v", err)
	}

	logger.Info("E2 service stopped")
}