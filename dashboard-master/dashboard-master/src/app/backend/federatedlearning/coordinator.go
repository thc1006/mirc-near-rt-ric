// Copyright 2024 The O-RAN Near-RT RIC Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package federatedlearning

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubernetes/dashboard/src/app/backend/errors"
)

// FLCoordinator implements the FederatedLearningManager interface
// and serves as the central coordinator for federated learning in the O-RAN Near-RT RIC
type FLCoordinator struct {
	// Core dependencies
	logger          *slog.Logger
	kubeClient      kubernetes.Interface
	
	// Storage and persistence
	clientStore     ClientStore
	modelStore      ModelStore
	metricsStore    MetricsStore
	
	// Security and privacy
	cryptoEngine    CryptoEngine
	privacyEngine   PrivacyEngine
	trustManager    TrustManager
	
	// Training coordination
	trainingOrchestrator TrainingOrchestrator
	aggregationEngine    AggregationEngine
	
	// Communication
	grpcServer      *grpc.Server
	clientConnections map[string]*grpc.ClientConn
	connMutex       sync.RWMutex
	
	// Configuration and state
	config          CoordinatorConfig
	activeJobs      map[string]*ActiveTrainingJob
	jobsMutex       sync.RWMutex
	
	// Monitoring and observability
	metricsCollector MetricsCollector
	alertManager     AlertManager
	
	// Resource management
	resourceManager  ResourceManager
	schedulingEngine SchedulingEngine
	
	// Shutdown coordination
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// CoordinatorConfig defines configuration for the FL coordinator
type CoordinatorConfig struct {
	// Network configuration
	ListenAddress       string        `yaml:"listen_address"`
	TLSEnabled          bool          `yaml:"tls_enabled"`
	CertificatePath     string        `yaml:"certificate_path"`
	PrivateKeyPath      string        `yaml:"private_key_path"`
	
	// Training configuration
	MaxConcurrentJobs   int           `yaml:"max_concurrent_jobs"`
	DefaultTimeout      time.Duration `yaml:"default_timeout"`
	HeartbeatInterval   time.Duration `yaml:"heartbeat_interval"`
	ClientTimeout       time.Duration `yaml:"client_timeout"`
	
	// Security configuration
	RequireMutualTLS    bool          `yaml:"require_mutual_tls"`
	TrustScoreThreshold float64       `yaml:"trust_score_threshold"`
	MaxFailedAttempts   int           `yaml:"max_failed_attempts"`
	
	// Privacy configuration
	DefaultPrivacyBudget float64      `yaml:"default_privacy_budget"`
	EnforcePrivacy      bool          `yaml:"enforce_privacy"`
	
	// Resource limits
	MaxModelSize        int64         `yaml:"max_model_size_bytes"`
	MaxClientsPerJob    int           `yaml:"max_clients_per_job"`
	
	// RRM-specific configuration
	E2LatencyThreshold  float64       `yaml:"e2_latency_threshold_ms"`
	MinSpectrumEfficiency float64     `yaml:"min_spectrum_efficiency"`
}

// ActiveTrainingJob represents an active federated learning training job
type ActiveTrainingJob struct {
	Job              *TrainingJob
	Participants     map[string]*FLClient
	CurrentModel     *GlobalModel
	RoundStartTime   time.Time
	LastAggregation  time.Time
	CompletedRounds  int64
	FailedRounds     int64
	Status           TrainingStatus
	ErrorCount       int
	
	// Synchronization
	roundMutex       sync.RWMutex
	participantMutex sync.RWMutex
	
	// Communication channels
	modelUpdates     chan ModelUpdate
	statusUpdates    chan TrainingStatusUpdate
	errorChan        chan error
}

// TrainingStatusUpdate represents a status update for a training job
type TrainingStatusUpdate struct {
	JobID     string
	ClientID  string
	Status    FLClientStatus
	Message   string
	Timestamp time.Time
}

// NewFLCoordinator creates a new federated learning coordinator
func NewFLCoordinator(
	logger *slog.Logger,
	kubeClient kubernetes.Interface,
	config CoordinatorConfig,
) (*FLCoordinator, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	coordinator := &FLCoordinator{
		logger:            logger,
		kubeClient:        kubeClient,
		config:            config,
		clientConnections: make(map[string]*grpc.ClientConn),
		activeJobs:        make(map[string]*ActiveTrainingJob),
		ctx:               ctx,
		cancel:            cancel,
	}
	
	// Initialize storage backends
	if err := coordinator.initializeStorage(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	
	// Initialize security components
	if err := coordinator.initializeSecurity(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize security: %w", err)
	}
	
	// Initialize training components
	if err := coordinator.initializeTraining(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize training: %w", err)
	}
	
	// Initialize monitoring
	if err := coordinator.initializeMonitoring(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize monitoring: %w", err)
	}
	
	// Start background services
	coordinator.startBackgroundServices()
	
	return coordinator, nil
}

// RegisterClient registers a new federated learning client (xApp)
func (c *FLCoordinator) RegisterClient(ctx context.Context, client *FLClient) error {
	c.logger.Info("Registering new FL client",
		slog.String("client_id", client.ID),
		slog.String("xapp_name", client.XAppName),
		slog.Any("rrm_tasks", client.RRMTasks))
	
	// Validate client registration
	if err := c.validateClientRegistration(client); err != nil {
		return errors.NewErrorBuilder(errors.ErrInvalidInput).
			WithDetail("validation_error", err.Error()).
			WithCaller().
			Build()
	}
	
	// Generate unique client ID if not provided
	if client.ID == "" {
		client.ID = uuid.New().String()
	}
	
	// Set registration metadata
	client.RegisteredAt = time.Now()
	client.UpdatedAt = time.Now()
	client.Status = FLClientStatusRegistered
	client.TrustScore = c.calculateInitialTrustScore(client)
	
	// Generate cryptographic keys if not provided
	if client.PublicKey == nil {
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return errors.NewInternal(fmt.Sprintf("failed to generate keys: %v", err))
		}
		client.PublicKey = publicKey
		// Store private key securely (implementation specific)
		if err := c.cryptoEngine.StorePrivateKey(client.ID, privateKey); err != nil {
			return errors.NewInternal(fmt.Sprintf("failed to store private key: %v", err))
		}
	}
	
	// Verify network slice associations
	if err := c.validateNetworkSlices(ctx, client.NetworkSlices); err != nil {
		return errors.NewErrorBuilder(errors.ErrInvalidInput).
			WithDetail("network_slice_error", err.Error()).
			WithCaller().
			Build()
	}
	
	// Store client registration
	if err := c.clientStore.Store(ctx, client); err != nil {
		return errors.NewInternal(fmt.Sprintf("failed to store client: %v", err))
	}
	
	// Update trust manager
	if err := c.trustManager.RegisterClient(ctx, client); err != nil {
		c.logger.Warn("Failed to register client with trust manager",
			slog.String("client_id", client.ID),
			slog.String("error", err.Error()))
	}
	
	// Emit registration event
	c.metricsCollector.RecordClientRegistration(client)
	
	c.logger.Info("Successfully registered FL client",
		slog.String("client_id", client.ID),
		slog.Float64("trust_score", client.TrustScore))
	
	return nil
}

// UnregisterClient removes a federated learning client
func (c *FLCoordinator) UnregisterClient(ctx context.Context, clientID string) error {
	c.logger.Info("Unregistering FL client", slog.String("client_id", clientID))
	
	// Get client information
	client, err := c.clientStore.Get(ctx, clientID)
	if err != nil {
		return errors.NewNotFound("client", clientID)
	}
	
	// Check if client is participating in active training
	c.jobsMutex.RLock()
	activeParticipation := false
	for jobID, job := range c.activeJobs {
		job.participantMutex.RLock()
		if _, exists := job.Participants[clientID]; exists {
			activeParticipation = true
			c.logger.Warn("Client participating in active training",
				slog.String("client_id", clientID),
				slog.String("job_id", jobID))
		}
		job.participantMutex.RUnlock()
	}
	c.jobsMutex.RUnlock()
	
	if activeParticipation {
		return errors.NewErrorBuilder(errors.ErrResourceConflict).
			WithDetail("reason", "client is participating in active training").
			WithCaller().
			Build()
	}
	
	// Close any existing connections
	c.connMutex.Lock()
	if conn, exists := c.clientConnections[clientID]; exists {
		conn.Close()
		delete(c.clientConnections, clientID)
	}
	c.connMutex.Unlock()
	
	// Remove from trust manager
	if err := c.trustManager.UnregisterClient(ctx, clientID); err != nil {
		c.logger.Warn("Failed to unregister client from trust manager",
			slog.String("client_id", clientID),
			slog.String("error", err.Error()))
	}
	
	// Remove client data
	if err := c.clientStore.Delete(ctx, clientID); err != nil {
		return errors.NewInternal(fmt.Sprintf("failed to delete client: %v", err))
	}
	
	// Clean up cryptographic material
	if err := c.cryptoEngine.DeleteKeys(clientID); err != nil {
		c.logger.Warn("Failed to delete cryptographic keys",
			slog.String("client_id", clientID),
			slog.String("error", err.Error()))
	}
	
	// Emit unregistration event
	c.metricsCollector.RecordClientUnregistration(client)
	
	c.logger.Info("Successfully unregistered FL client", slog.String("client_id", clientID))
	return nil
}

// GetClient retrieves a registered client by ID
func (c *FLCoordinator) GetClient(ctx context.Context, clientID string) (*FLClient, error) {
	client, err := c.clientStore.Get(ctx, clientID)
	if err != nil {
		return nil, errors.NewNotFound("client", clientID)
	}
	
	// Update last access time
	client.UpdatedAt = time.Now()
	if err := c.clientStore.Store(ctx, client); err != nil {
		c.logger.Warn("Failed to update client access time",
			slog.String("client_id", clientID),
			slog.String("error", err.Error()))
	}
	
	return client, nil
}

// ListClients returns a list of registered clients matching the selector
func (c *FLCoordinator) ListClients(ctx context.Context, selector ClientSelector) ([]*FLClient, error) {
	allClients, err := c.clientStore.List(ctx)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to list clients: %v", err))
	}
	
	var filteredClients []*FLClient
	for _, client := range allClients {
		if c.matchesSelector(client, selector) {
			filteredClients = append(filteredClients, client)
		}
	}
	
	// Apply limits
	if selector.MaxClients > 0 && len(filteredClients) > selector.MaxClients {
		// Sort by trust score and take top clients
		filteredClients = c.selectTopClients(filteredClients, selector.MaxClients)
	}
	
	return filteredClients, nil
}

// UpdateClientStatus updates the status of a registered client
func (c *FLCoordinator) UpdateClientStatus(ctx context.Context, clientID string, status FLClientStatus) error {
	client, err := c.clientStore.Get(ctx, clientID)
	if err != nil {
		return errors.NewNotFound("client", clientID)
	}
	
	oldStatus := client.Status
	client.Status = status
	client.UpdatedAt = time.Now()
	
	if status == FLClientStatusIdle || status == FLClientStatusTraining {
		client.LastHeartbeat = time.Now()
	}
	
	if err := c.clientStore.Store(ctx, client); err != nil {
		return errors.NewInternal(fmt.Sprintf("failed to update client status: %v", err))
	}
	
	// Emit status change event
	c.metricsCollector.RecordClientStatusChange(clientID, oldStatus, status)
	
	c.logger.Debug("Updated client status",
		slog.String("client_id", clientID),
		slog.String("old_status", string(oldStatus)),
		slog.String("new_status", string(status)))
	
	return nil
}

// StartTraining initiates a new federated learning training job
func (c *FLCoordinator) StartTraining(ctx context.Context, jobSpec TrainingJobSpec) (*TrainingJob, error) {
	c.logger.Info("Starting new FL training job",
		slog.String("model_id", jobSpec.ModelID),
		slog.String("rrm_task", string(jobSpec.RRMTask)))
	
	// Validate job specification
	if err := c.validateTrainingJobSpec(jobSpec); err != nil {
		return nil, errors.NewErrorBuilder(errors.ErrInvalidInput).
			WithDetail("validation_error", err.Error()).
			WithCaller().
			Build()
	}
	
	// Check resource availability
	if err := c.resourceManager.CheckAvailability(ctx, jobSpec.Resources); err != nil {
		return nil, errors.NewErrorBuilder(errors.ErrServiceUnavailable).
			WithDetail("resource_error", err.Error()).
			WithCaller().
			Build()
	}
	
	// Create training job
	job := &TrainingJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "federatedlearning.o-ran.org/v1",
			Kind:       "TrainingJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("fl-job-%s", uuid.New().String()[:8]),
			Namespace: "federated-learning",
			Labels: map[string]string{
				"app":      "federated-learning",
				"rrm-task": string(jobSpec.RRMTask),
				"model-id": jobSpec.ModelID,
			},
		},
		Spec: jobSpec,
		Status: TrainingJobStatus{
			Phase:      TrainingStatusInitializing,
			LastUpdate: metav1.Now(),
		},
	}
	
	// Get or create global model
	globalModel, err := c.getOrCreateGlobalModel(ctx, jobSpec.ModelID, jobSpec.RRMTask)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to get global model: %v", err))
	}
	
	// Select participating clients
	selectedClients, err := c.selectClientsForTraining(ctx, jobSpec.ClientSelector, jobSpec.RRMTask)
	if err != nil {
		return nil, errors.NewInternal(fmt.Sprintf("failed to select clients: %v", err))
	}
	
	if len(selectedClients) < jobSpec.TrainingConfig.MinParticipants {
		return nil, errors.NewErrorBuilder(errors.ErrServiceUnavailable).
			WithDetail("reason", "insufficient participants").
			WithDetail("required", jobSpec.TrainingConfig.MinParticipants).
			WithDetail("available", len(selectedClients)).
			WithCaller().
			Build()
	}
	
	// Create active training job
	activeJob := &ActiveTrainingJob{
		Job:              job,
		Participants:     make(map[string]*FLClient),
		CurrentModel:     globalModel,
		RoundStartTime:   time.Now(),
		LastAggregation:  time.Now(),
		Status:           TrainingStatusInitializing,
		modelUpdates:     make(chan ModelUpdate, len(selectedClients)*2),
		statusUpdates:    make(chan TrainingStatusUpdate, len(selectedClients)*2),
		errorChan:        make(chan error, len(selectedClients)),
	}
	
	// Add participants
	for _, client := range selectedClients {
		activeJob.Participants[client.ID] = client
	}
	
	// Store active job
	c.jobsMutex.Lock()
	c.activeJobs[job.Name] = activeJob
	c.jobsMutex.Unlock()
	
	// Start training orchestration
	c.wg.Add(1)
	go c.orchestrateTraining(ctx, activeJob)
	
	// Update job status
	job.Status.Phase = TrainingStatusWaiting
	job.Status.ParticipatingClients = make([]string, 0, len(selectedClients))
	for clientID := range activeJob.Participants {
		job.Status.ParticipatingClients = append(job.Status.ParticipatingClients, clientID)
	}
	job.Status.LastUpdate = metav1.Now()
	
	// Emit training start event
	c.metricsCollector.RecordTrainingJobStart(job)
	
	c.logger.Info("Successfully started FL training job",
		slog.String("job_id", job.Name),
		slog.Int("participants", len(selectedClients)),
		slog.String("model_id", jobSpec.ModelID))
	
	return job, nil
}

// Helper methods and initialization functions would continue here...
// This includes methods like:
// - initializeStorage()
// - initializeSecurity()
// - initializeTraining()
// - initializeMonitoring()
// - startBackgroundServices()
// - validateClientRegistration()
// - calculateInitialTrustScore()
// - validateNetworkSlices()
// - matchesSelector()
// - selectTopClients()
// - validateTrainingJobSpec()
// - getOrCreateGlobalModel()
// - selectClientsForTraining()
// - orchestrateTraining()

// Shutdown gracefully shuts down the FL coordinator
func (c *FLCoordinator) Shutdown(ctx context.Context) error {
	c.logger.Info("Shutting down FL coordinator")
	
	// Cancel context to stop background services
	c.cancel()
	
	// Stop accepting new training jobs
	c.jobsMutex.Lock()
	for jobID, job := range c.activeJobs {
		c.logger.Info("Stopping active training job", slog.String("job_id", jobID))
		job.Status = TrainingStatusPaused
		close(job.modelUpdates)
		close(job.statusUpdates)
		close(job.errorChan)
	}
	c.jobsMutex.Unlock()
	
	// Close client connections
	c.connMutex.Lock()
	for clientID, conn := range c.clientConnections {
		c.logger.Debug("Closing client connection", slog.String("client_id", clientID))
		conn.Close()
	}
	c.clientConnections = make(map[string]*grpc.ClientConn)
	c.connMutex.Unlock()
	
	// Stop gRPC server
	if c.grpcServer != nil {
		c.grpcServer.GracefulStop()
	}
	
	// Wait for background services to complete
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		c.logger.Info("FL coordinator shutdown completed")
	case <-ctx.Done():
		c.logger.Warn("FL coordinator shutdown timed out")
		return errors.NewTimeout("shutdown", ctx.Err().Error())
	}
	
	return nil
}