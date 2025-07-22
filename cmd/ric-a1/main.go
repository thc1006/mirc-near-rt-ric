package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hctsai1006/near-rt-ric/pkg/a1"
	"github.com/hctsai1006/near-rt-ric/pkg/a1/db"
	"github.com/hctsai1006/near-rt-ric/pkg/a1/enrichment"
	"github.com/hctsai1006/near-rt-ric/pkg/a1/mlmodel"
	"github.com/hctsai1006/near-rt-ric/pkg/a1/policy"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Dummy config and logging packages for compilation
type ServerConfig struct {
	Address string
	Port    int
	Timeout int
	TLS     TLSConfig
}
type TLSConfig struct {
	Enabled  bool
	CertPath string
	KeyPath  string
}
type DatabaseConfig struct {
	ConnectionString string
}
type A1Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	LogLevel string
}

func NewLogger(level string) *logrus.Logger {
	logger := logrus.New()
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})
	return logger
}

func main() {
	// Load configuration
	// TODO: Implement a proper configuration loading mechanism
	cfg := &A1Config{
		Server: ServerConfig{
			Address: "0.0.0.0",
			Port:    10020,
			Timeout: 30,
			TLS: TLSConfig{
				Enabled:  false,
				CertPath: "certs/server.crt",
				KeyPath:  "certs/server.key",
			},
		},
		Database: DatabaseConfig{
			// Example connection string. Replace with your actual connection details.
			// It's recommended to use environment variables for sensitive data.
			ConnectionString: "host=localhost port=5432 user=ricuser password=ricpassword dbname=ricdb sslmode=disable",
		},
		LogLevel: "info",
	}

	// Initialize logger
	logger := NewLogger(cfg.LogLevel)
	logger.Info("starting RIC A1 service")

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection
	dbClient, err := db.NewDB(cfg.Database.ConnectionString)
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}
	defer dbClient.Close()

	// Initialize managers
	policyMgr := policy.NewPostgresPolicyManager(dbClient)
	if err := policyMgr.CreatePoliciesTable(ctx); err != nil {
		logger.Fatalf("failed to create policies table: %v", err)
	}

	mlModelMgr := mlmodel.NewPostgresMLModelManager(dbClient)
	if err := mlModelMgr.CreateMLModelsTable(ctx); err != nil {
		logger.Fatalf("failed to create ml_models table: %v", err)
	}
	enrichMgr := enrichment.NewPostgresEnrichmentManager(dbClient)
	if err := enrichMgr.CreateEnrichmentInfoTable(ctx); err != nil {
		logger.Fatalf("failed to create enrichment_info table: %v", err)
	}

	// Initialize A1 server
	a1Server := a1.NewA1Server(policyMgr, mlModelMgr, enrichMgr)

	// Create HTTP server
	listenAddr := fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)
	router := http.NewServeMux()

	// Policy Management Endpoints
	router.HandleFunc("/a1-p/v2/policies", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			a1Server.CreatePolicy(w, r)
		case http.MethodGet:
			a1Server.GetAllPolicies(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	router.HandleFunc("/a1-p/v2/policies/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/a1-p/v2/policies/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			a1Server.GetPolicyById(w, r, id)
		case http.MethodPut:
			a1Server.UpdatePolicyById(w, r, id)
		case http.MethodDelete:
			a1Server.DeletePolicyById(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	router.HandleFunc("/a1-p/v2/policytypes", a1Server.GetPolicyTypes)

	// ML Model Management Endpoints
	router.HandleFunc("/a1-ml/v1/models", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			a1Server.DeployMLModel(w, r)
		case http.MethodGet:
			a1Server.GetAllMLModels(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	router.HandleFunc("/a1-ml/v1/models/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/a1-ml/v1/models/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			a1Server.GetMLModelById(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	router.HandleFunc("/a1-ml/v1/models/{id}/rollback", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/a1-ml/v1/models/")
		id = strings.TrimSuffix(id, "/rollback")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodPost:
			a1Server.RollbackMLModel(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Enrichment Information Endpoints
	router.HandleFunc("/a1-ei/v1/info", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			a1Server.UpdateEnrichmentInfo(w, r)
		case http.MethodGet:
			a1Server.GetAllEnrichmentInfo(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	router.HandleFunc("/a1-ei/v1/info/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/a1-ei/v1/info/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			a1Server.GetEnrichmentInfoById(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      a1Server.LoggingMiddleware(a1Server.AuthMiddleware(router)), // Apply middleware
		ReadTimeout:  time.Duration(cfg.Server.Timeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.Timeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.Timeout) * time.Second,
	}

	// Start the A1 server
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

	// Stop the A1 server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("error while stopping A1 server: %v", err)
	}

	logger.Info("A1 service stopped")
}
