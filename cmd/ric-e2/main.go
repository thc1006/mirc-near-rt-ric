package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hctsai1006/near-rt-ric/pkg/config"
	"github.com/hctsai1006/near-rt-ric/pkg/e2/asn1"
	"github.com/hctsai1006/near-rt-ric/pkg/e2/node"
	"github.com/hctsai1006/near-rt-ric/pkg/e2/subscription"
	"github.com/hctsai1006/near-rt-ric/pkg/e2/transport"
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

	// Create a new E2 codec, subscription manager, handler, and server
	codec := asn1.NewCodec(logger)
	sm := subscription.NewManager(logger)
	handler := node.NewHandler(logger, codec, sm)
	server := transport.NewServer(logger, handler)

	// Start the server in a goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)
		if err := server.Start(context.Background(), addr); err != nil {
			logger.Fatalf("failed to start E2 server: %v", err)
		}
	}()

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down E2 service")

	// Stop the server
	server.Stop()

	logger.Info("E2 service stopped")
}
