package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/hctsai1006/near-rt-ric/pkg/config"
	"github.com/hctsai1006/near-rt-ric/pkg/logging"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("Control")
	if err != nil {
		logrus.Fatalf("failed to load Control configuration: %v", err)
	}

	// Initialize logger
	logger := logging.NewLogger(cfg.LogLevel)
	logger.Info("starting RIC Control service")

	// Create a context that can be cancelled
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: Initialize control plane components (xApp manager, etc.)

	logger.Info("Control service started")

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down Control service")

	// TODO: Stop the control plane components

	logger.Info("Control service stopped")
}
