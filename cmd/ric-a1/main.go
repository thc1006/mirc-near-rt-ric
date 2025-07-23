package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/hctsai1006/near-rt-ric/pkg/a1"
	"github.com/hctsai1006/near-rt-ric/pkg/config"
	"github.com/hctsai1006/near-rt-ric/pkg/logging"
	"github.com/hctsai1006/near-rt-ric/pkg/common/monitoring"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("A1")
	if err != nil {
		logrus.Fatalf("failed to load A1 configuration: %v", err)
	}

	// Initialize logger
	logger := logging.NewLogger(cfg.LogLevel)
	logger.Info("starting RIC A1 service")

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Address,
	})

	// Initialize services
	authService, err := a1.NewAuthService(cfg, logger, redisClient)
	if err != nil {
		logger.Fatalf("failed to create auth service: %v", err)
	}

	policyManager := a1.NewPolicyManager(logger)
	metricsCollector := monitoring.NewMetricsCollector()

	// Initialize API handlers
	apiHandlers := a1.NewAPIHandlers(policyManager, authService, logger, metricsCollector)

	// Create router and set up routes
	router := mux.NewRouter()
	apiHandlers.SetupRoutes(router)

	// Create HTTP server
	listenAddr := fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)
	server := &http.Server{
		Addr:         listenAddr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.Timeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.Timeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.Timeout) * time.Second,
	}

	// Start the server
	go func() {
		if cfg.Server.TLS.Enabled {
			logger.Infof("Starting A1 server with TLS on %s", listenAddr)
			if err := server.ListenAndServeTLS(cfg.Server.TLS.CertPath, cfg.Server.TLS.KeyPath); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("A1 server failed to listen with TLS: %v", err)
			}
		} else {
			logger.Infof("Starting A1 server without TLS on %s", listenAddr)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("A1 server failed to listen: %v", err)
			}
		}
	}()

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("shutting down A1 service")

	// Stop the server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("error while stopping A1 server: %v", err)
	}

	logger.Info("A1 service stopped")
}
