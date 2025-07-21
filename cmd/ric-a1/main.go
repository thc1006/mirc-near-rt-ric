package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hctsai1006/near-rt-ric/pkg/a1"
	"github.com/sirupsen/logrus"
)

const (
	version = "1.0.0"
	appName = "O-RAN A1 Interface"
)

var (
	listenAddr = flag.String("listen-addr", "0.0.0.0", "Listen address for A1 interface")
	listenPort = flag.Int("listen-port", 10020, "Listen port for A1 interface") 
	tlsEnabled = flag.Bool("tls-enabled", true, "Enable TLS")
	tlsCert    = flag.String("tls-cert", "/certs/tls.crt", "TLS certificate file")
	tlsKey     = flag.String("tls-key", "/certs/tls.key", "TLS private key file")
	authEnabled = flag.Bool("auth-enabled", true, "Enable authentication")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
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
		"tls_enabled": *tlsEnabled,
		"auth_enabled": *authEnabled,
	}).Info("Starting O-RAN A1 Interface")

	// Create A1 interface configuration
	config := &a1.A1InterfaceConfig{
		ListenAddress:         *listenAddr,
		ListenPort:            *listenPort,
		TLSEnabled:            *tlsEnabled,
		TLSCertPath:           *tlsCert,
		TLSKeyPath:            *tlsKey,
		AuthenticationEnabled: *authEnabled,
		RequestTimeout:        30 * time.Second,
		MaxRequestSize:        1024 * 1024, // 1MB
		RateLimitEnabled:      true,
		RateLimitPerMinute:    1000,
		NotificationEnabled:   true,
		DatabaseURL:           "", // Using in-memory for now
		LogLevel:              *logLevel,
	}

	// Create repository (in-memory for now)
	repository := a1.NewInMemoryA1Repository()

	// Create and start A1 interface
	a1Interface := a1.NewA1Interface(config, repository)

	if err := a1Interface.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start A1 interface")
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("O-RAN A1 Interface started successfully")

	// Start status reporting goroutine
	go statusReporter(a1Interface, repository, logger)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.WithField("signal", sig.String()).Info("Received shutdown signal")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	shutdownChan := make(chan error, 1)
	go func() {
		shutdownChan <- a1Interface.Stop()
	}()

	select {
	case err := <-shutdownChan:
		if err != nil {
			logger.WithError(err).Error("Error during shutdown")
			os.Exit(1)
		}
		logger.Info("O-RAN A1 Interface shutdown completed successfully")
	case <-shutdownCtx.Done():
		logger.Error("Shutdown timeout exceeded")
		os.Exit(1)
	}
}

// statusReporter periodically reports the status of the A1 interface
func statusReporter(a1Interface *a1.A1Interface, repository a1.A1Repository, logger *logrus.Logger) {
	ticker := time.NewTicker(5 * time.Minute) // Report every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status := a1Interface.GetStatus()
			
			// Get repository statistics if supported
			var repoStats map[string]interface{}
			if repo, ok := repository.(*a1.InMemoryA1Repository); ok {
				repoStats = repo.GetStatistics()
			}

			logger.WithFields(logrus.Fields{
				"status":                status.Status,
				"policy_types":          status.PolicyTypes,
				"total_policies":        status.Policies,
				"active_policies":       status.ActivePolicies,
				"repository_statistics": repoStats,
			}).Info("O-RAN A1 Interface status report")

			// Log policy types and their usage
			if policyTypes, err := repository.ListPolicyTypes(); err == nil {
				for _, policyType := range policyTypes {
					policies, _ := repository.ListPoliciesByType(policyType.PolicyTypeID)
					logger.WithFields(logrus.Fields{
						"policy_type_id": policyType.PolicyTypeID,
						"name":           policyType.Name,
						"policy_count":   len(policies),
					}).Debug("Policy type details")
				}
			}
		}
	}
}