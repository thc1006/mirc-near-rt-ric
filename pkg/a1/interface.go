package a1

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/hctsai1006/near-rt-ric/internal/config"
	"github.com/hctsai1006/near-rt-ric/pkg/common/monitoring"
	"github.com/sirupsen/logrus"
)

// A1Interface represents the main A1 interface implementation
// Complies with O-RAN.WG2.A1-v06.00 specification
type A1Interface struct {
	config *config.A1Config
	logger *logrus.Logger
	metrics *monitoring.MetricsCollector

	// Core components
	policyManager     *PolicyManager
	mlModelManager    *MLModelManager
	enrichmentManager *EnrichmentManager
	authService       *AuthService
	database          *Database
	httpServer        *http.Server
	router            *mux.Router
	redisClient       *redis.Client

	// Control and synchronization
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running bool
	mutex   sync.RWMutex

	// Event handlers
	eventHandlers []A1InterfaceEventHandler

	// Background tasks
	cleanupTicker *time.Ticker
}

// A1InterfaceEventHandler defines interface for A1 interface events
type A1InterfaceEventHandler interface {
	OnA1InterfaceStarted()
	OnA1InterfaceStopped()
	OnPolicyTypeCreated(policyType *PolicyType)
	OnPolicyTypeDeleted(policyTypeID PolicyTypeID)
	OnPolicyInstanceCreated(policy *PolicyInstance)
	OnPolicyInstanceDeleted(policyID PolicyID)
	OnError(err error)
}

// NewA1Interface creates a new A1 interface instance
func NewA1Interface(cfg *config.A1Config, logger *logrus.Logger, metrics *monitoring.MetricsCollector, redisClient *redis.Client, dbConfig *config.DatabaseConfig) (*A1Interface, error) {
	ctx, cancel := context.WithCancel(context.Background())

	a1 := &A1Interface{
		config:      cfg,
		logger:      logger.WithField("component", "a1-interface"),
		metrics:     metrics,
		redisClient: redisClient,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize database
	var err error
	a1.database, err = NewDatabase(dbConfig, logger)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Run database migrations
	if err := a1.database.Migrate(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	// Initialize authentication service
	a1.authService, err = NewAuthService(cfg, logger, redisClient)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create auth service: %w", err)
	}

	// Initialize policy manager
	a1.policyManager = NewPolicyManager(cfg, logger, metrics, a1.database.db)

	// Initialize ML model manager
	a1.mlModelManager = NewMLModelManager(cfg, logger, a1.database.db)

	// Initialize enrichment manager
	a1.enrichmentManager = NewEnrichmentManager(cfg, logger, a1.database.db)

	// Initialize API handlers
	a1.apiHandlers = NewAPIHandlers(a1.policyManager, a1.mlModelManager, a1.enrichmentManager, a1.authService, logger, metrics)

	// Set up event handlers
	a1.policyManager.AddEventHandler(a1)

	// Initialize HTTP router
	a1.setupRoutes()

	// Initialize HTTP server
	a1.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.ListenAddress, cfg.HTTP.Port),
		Handler:      a1.router,
		ReadTimeout:  time.Duration(cfg.HTTP.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.HTTP.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.HTTP.IdleTimeout) * time.Second,
	}

	a1.logger.Info("A1 interface initialized successfully")
	return a1, nil
}

// setupRoutes sets up HTTP routes
func (a1 *A1Interface) setupRoutes() {
	a1.router = mux.NewRouter()

	// Set up CORS middleware if enabled
	if a1.config.HTTP.CORS.Enabled {
		a1.router.Use(a1.corsMiddleware)
	}

	// Set up common middleware
	a1.router.Use(a1.loggingMiddleware)
	a1.router.Use(a1.metricsMiddleware)

	// Set up API routes
	a1.apiHandlers.SetupRoutes(a1.router)

	a1.logger.Info("A1 HTTP routes configured")
}

// Start starts the A1 interface
func (a1 *A1Interface) Start(ctx context.Context) error {
	a1.mutex.Lock()
	defer a1.mutex.Unlock()

	if a1.running {
		return fmt.Errorf("A1 interface is already running")
	}

	a1.logger.WithFields(logrus.Fields{
		"http_address": fmt.Sprintf("%s:%d", a1.config.HTTP.ListenAddress, a1.config.HTTP.Port),
		"tls_enabled":  a1.config.HTTP.TLS.Enabled,
	}).Info("Starting O-RAN A1 interface")

	// Start policy manager
	if err := a1.policyManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start policy manager: %w", err)
	}

	// Start background cleanup tasks
	a1.startBackgroundTasks()

	// Start HTTP server
	a1.wg.Add(1)
	go func() {
		defer a1.wg.Done()
		
		var err error
		if a1.config.HTTP.TLS.Enabled {
			err = a1.httpServer.ListenAndServeTLS(a1.config.HTTP.TLS.CertFile, a1.config.HTTP.TLS.KeyFile)
		} else {
			err = a1.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			a1.logger.WithError(err).Error("HTTP server error")
		}
	}()

	a1.running = true
	a1.logger.Info("O-RAN A1 interface started successfully")

	// Send event
	for _, handler := range a1.eventHandlers {
		go handler.OnA1InterfaceStarted()
	}

	return nil
}

// Stop stops the A1 interface gracefully
func (a1 *A1Interface) Stop(ctx context.Context) error {
	a1.mutex.Lock()
	defer a1.mutex.Unlock()

	if !a1.running {
		return nil
	}

	a1.logger.Info("Stopping O-RAN A1 interface")

	// Stop background tasks
	a1.stopBackgroundTasks()

	// Stop HTTP server
	if err := a1.httpServer.Shutdown(ctx); err != nil {
		a1.logger.WithError(err).Error("Error shutting down HTTP server")
	}

	// Stop policy manager
	if err := a1.policyManager.Stop(ctx); err != nil {
		a1.logger.WithError(err).Error("Error stopping policy manager")
	}

	// Cancel context
	a1.cancel()

	// Wait for goroutines to finish
	done := make(chan struct{})
	go func() {
		a1.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		a1.logger.Info("O-RAN A1 interface stopped successfully")
	case <-ctx.Done():
		a1.logger.Warn("A1 interface shutdown timeout")
	}

	a1.running = false

	// Send event
	for _, handler := range a1.eventHandlers {
		go handler.OnA1InterfaceStopped()
	}

	return nil
}

// startBackgroundTasks starts background maintenance tasks
func (a1 *A1Interface) startBackgroundTasks() {
	// Token cleanup task
	a1.cleanupTicker = time.NewTicker(1 * time.Hour)
	a1.wg.Add(1)
	go func() {
		defer a1.wg.Done()
		for {
			select {
			case <-a1.ctx.Done():
				return
			case <-a1.cleanupTicker.C:
				a1.authService.CleanupExpiredTokens()
			}
		}
	}()
}

// stopBackgroundTasks stops background maintenance tasks
func (a1 *A1Interface) stopBackgroundTasks() {
	if a1.cleanupTicker != nil {
		a1.cleanupTicker.Stop()
	}
}

// Middleware functions

// corsMiddleware handles CORS headers
func (a1 *A1Interface) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (a1 *A1Interface) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		a1.logger.WithFields(logrus.Fields{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": wrapped.statusCode,
			"duration":    duration,
			"remote_addr": r.RemoteAddr,
			"user_agent":  r.UserAgent(),
		}).Info("A1 HTTP request processed")
	})
}

// metricsMiddleware records metrics for HTTP requests
func (a1 *A1Interface) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code and size
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// Record metrics
		a1.metrics.RecordA1Request(r.Method, r.URL.Path, wrapped.statusCode, duration, 0, wrapped.size)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.size += size
	return size, err
}

// PolicyEventHandler interface implementation
func (a1 *A1Interface) OnPolicyTypeCreated(policyType *PolicyType) {
	for _, handler := range a1.eventHandlers {
		go handler.OnPolicyTypeCreated(policyType)
	}
}

func (a1 *A1Interface) OnPolicyTypeDeleted(policyTypeID PolicyTypeID) {
	for _, handler := range a1.eventHandlers {
		go handler.OnPolicyTypeDeleted(policyTypeID)
	}
}

func (a1 *A1Interface) OnPolicyInstanceCreated(policy *PolicyInstance) {
	for _, handler := range a1.eventHandlers {
		go handler.OnPolicyInstanceCreated(policy)
	}
}

func (a1 *A1Interface) OnPolicyInstanceUpdated(policy *PolicyInstance) {}

func (a1 *A1Interface) OnPolicyInstanceDeleted(policyID PolicyID) {
	for _, handler := range a1.eventHandlers {
		go handler.OnPolicyInstanceDeleted(policyID)
	}
}

func (a1 *A1Interface) OnPolicyEnforced(policy *PolicyInstance) {}

func (a1 *A1Interface) OnPolicyFailed(policy *PolicyInstance, err error) {
	for _, handler := range a1.eventHandlers {
		go handler.OnError(err)
	}
}

// Public interface methods

// IsRunning returns whether the A1 interface is running
func (a1 *A1Interface) IsRunning() bool {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()
	return a1.running
}

// Database returns the database instance
func (a1 *A1Interface) Database() *Database {
	return a1.database
}

// AddEventHandler adds an event handler
func (a1 *A1Interface) AddEventHandler(handler A1InterfaceEventHandler) {
	a1.eventHandlers = append(a1.eventHandlers, handler)
}

// GetStats returns A1 interface statistics
func (a1 *A1Interface) GetStats() map[string]interface{} {
	stats := a1.policyManager.GetStatistics()
	return map[string]interface{}{
		"policy_stats":      stats,
		"running":           a1.IsRunning(),
		"server_address":    a1.httpServer.Addr,
		"tls_enabled":       a1.config.HTTP.TLS.Enabled,
		"cors_enabled":      a1.config.HTTP.CORS.Enabled,
	}
}

// HealthCheck performs a health check of the A1 interface
func (a1 *A1Interface) HealthCheck() error {
	if !a1.IsRunning() {
		return fmt.Errorf("A1 interface is not running")
	}
	return nil
}

// GetPolicyTypes returns all policy types
func (a1 *A1Interface) GetPolicyTypes() []*PolicyType {
	return a1.policyManager.GetAllPolicyTypes()
}

// GetPolicyType returns a specific policy type
func (a1 *A1Interface) GetPolicyType(policyTypeID PolicyTypeID) (*PolicyType, error) {
	return a1.policyManager.GetPolicyType(policyTypeID)
}

// GetPolicyInstances returns all policy instances
func (a1 *A1Interface) GetPolicyInstances() []*PolicyInstance {
	return a1.policyManager.GetAllPolicyInstances()
}

// GetPolicyInstance returns a specific policy instance
func (a1 *A1Interface) GetPolicyInstance(policyID PolicyID) (*PolicyInstance, error) {
	return a1.policyManager.GetPolicyInstance(policyID)
}

// GetPolicyInstancesByType returns policy instances for a specific type
func (a1 *A1Interface) GetPolicyInstancesByType(policyTypeID PolicyTypeID) []*PolicyInstance {
	return a1.policyManager.GetPolicyInstancesByType(policyTypeID)
}