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

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/kubernetes/dashboard/src/app/backend/errors"
)

// FLCoordinator implements the FederatedLearningManager interface
// and serves as the central coordinator for federated learning in the O-RAN Near-RT RIC
type FLCoordinator struct {
	// Core dependencies
	logger     *slog.Logger
	kubeClient kubernetes.Interface
	role       CoordinatorRole

	// Storage and persistence
	redisClient *redis.Client
	clientStore ClientStore
	modelStore  ModelStore

	// Security and privacy
	cryptoEngine  CryptoEngine
	privacyEngine PrivacyEngine
	trustManager  TrustManager

	// Training coordination
	trainingOrchestrator TrainingOrchestrator
	aggregationEngine    AggregationEngine

	// Communication
	grpcServer        *grpc.Server
	clientConnections map[string]*grpc.ClientConn
	connMutex         sync.RWMutex

	// Configuration and state
	configManager *ConfigManager
	activeJobs    map[string]*ActiveTrainingJob
	jobsMutex     sync.RWMutex
	jobQueue      *JobQueue

	// Monitoring and observability
	metricsCollector MetricsCollector
	alertManager     AlertManager

	// Resource management
	resourceManager  ResourceManager
	schedulingEngine SchedulingEngine

	// Shutdown coordination
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// CoordinatorConfig defines configuration for the FL coordinator
type CoordinatorConfig struct {
	// Network configuration
	ListenAddress       string        `yaml:"listen_address"`
	TLSEnabled          bool          `yaml:"tls_enabled"`
	CertificatePath     string        `yaml:"certificate_path"`
	PrivateKeyPath      string        `yaml:"private_key_path"`
	Role                CoordinatorRole `yaml:"role"`
	MasterEndpoint      string        `yaml:"master_endpoint"`

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
	AggregationWindow   time.Duration `yaml:"aggregation_window"`
	MinParticipationRatio float64       `yaml:"min_participation_ratio"`
	StalenessThreshold  time.Duration `yaml:"staleness_threshold"`

	// RRM-specific configuration
	E2LatencyThreshold  float64       `yaml:"e2_latency_threshold_ms"`
	MinSpectrumEfficiency float64     `yaml:"min_spectrum_efficiency"`
}

// ConfigManager manages the dynamic configuration for the FLCoordinator
type ConfigManager struct {
	logger        *slog.Logger
	kubeClient    kubernetes.Interface
	namespace     string
	configMapName string

	config      *CoordinatorConfig
	configMutex sync.RWMutex

	updateChan chan<- *CoordinatorConfig
	stopCh     chan struct{}
}

// NewConfigManager creates a new configuration manager.
func NewConfigManager(logger *slog.Logger, kubeClient kubernetes.Interface, namespace, configMapName string, updateChan chan<- *CoordinatorConfig) (*ConfigManager, error) {
	cm := &ConfigManager{
		logger:        logger,
		kubeClient:    kubeClient,
		namespace:     namespace,
		configMapName: configMapName,
		updateChan:    updateChan,
		stopCh:        make(chan struct{}),
	}

	initialConfig, err := cm.loadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load initial config: %w", err)
	}
	cm.config = initialConfig

	go cm.watchConfigChanges()

	return cm, nil
}

// GetConfig returns the current configuration safely.
func (cm *ConfigManager) GetConfig() *CoordinatorConfig {
	cm.configMutex.RLock()
	defer cm.configMutex.RUnlock()
	return cm.config
}

// loadConfig loads the configuration from the Kubernetes ConfigMap.
func (cm *ConfigManager) loadConfig() (*CoordinatorConfig, error) {
	cm.logger.Info("Loading configuration from ConfigMap", "namespace", cm.namespace, "name", cm.configMapName)
	configMap, err := cm.kubeClient.CoreV1().ConfigMaps(cm.namespace).Get(context.TODO(), cm.configMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ConfigMap %s/%s: %w", cm.namespace, cm.configMapName, err)
	}

	configData, ok := configMap.Data["fl-coordinator.yaml"]
	if !ok {
		return nil, fmt.Errorf("ConfigMap %s/%s does not contain key 'fl-coordinator.yaml'", cm.namespace, cm.configMapName)
	}

	var config CoordinatorConfig
	if err := yaml.Unmarshal([]byte(configData), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	cm.logger.Info("Successfully loaded configuration")
	return &config, nil
}

// watchConfigChanges sets up a watcher to monitor the ConfigMap for changes.
func (cm *ConfigManager) watchConfigChanges() {
	factory := informers.NewSharedInformerFactoryWithOptions(cm.kubeClient, 0, informers.WithNamespace(cm.namespace))
	informer := factory.Core().V1().ConfigMaps().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			newCM, ok := newObj.(*corev1.ConfigMap)
			if !ok {
				cm.logger.Error("Failed to cast new object to ConfigMap")
				return
			}
			if newCM.Name == cm.configMapName {
				cm.logger.Info("ConfigMap updated, reloading configuration")
				newConfig, err := cm.loadConfig()
				if err != nil {
					cm.logger.Error("Failed to reload configuration", "error", err)
					return
				}

				cm.configMutex.Lock()
				cm.config = newConfig
				cm.configMutex.Unlock()

				// Notify the coordinator of the update
				cm.updateChan <- newConfig
			}
		},
	})

	factory.Start(cm.stopCh)
	<-cm.stopCh
}

// Stop stops the config watcher.
func (cm *ConfigManager) Stop() {
	close(cm.stopCh)
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
	stalenessScores  map[string]float64
	updateTimestamps map[string]time.Time

	// Communication channels
	modelUpdates      chan ModelUpdate
	statusUpdates     chan TrainingStatusUpdate
	errorChan         chan error
	aggregationTrigger chan struct{}
	roundCompleteChan  chan int64 // New channel to signal round completion
	collectedUpdates  map[string]ModelUpdate
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
	namespace string,
	configMapName string,
	redisClient *redis.Client,
) (*FLCoordinator, error) {
	ctx, cancel := context.WithCancel(context.Background())
	configUpdateChan := make(chan *CoordinatorConfig)

	configManager, err := NewConfigManager(logger, kubeClient, namespace, configMapName, configUpdateChan)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	initialConfig := configManager.GetConfig()

	coordinator := &FLCoordinator{
		logger:            logger,
		kubeClient:        kubeClient,
		role:              initialConfig.Role,
		redisClient:       redisClient,
		configManager:     configManager,
		clientConnections: make(map[string]*grpc.ClientConn),
		activeJobs:        make(map[string]*ActiveTrainingJob),
		jobQueue:          NewJobQueue(initialConfig.MaxConcurrentJobs),
		ctx:               ctx,
		cancel:            cancel,
		cryptoEngine:      &mockCryptoEngine{logger: logger},
		privacyEngine:     &mockPrivacyEngine{logger: logger},
		trustManager:      &mockTrustManager{logger: logger},
		trainingOrchestrator: &mockTrainingOrchestrator{logger: logger},
		aggregationEngine:    &mockAggregationEngine{logger: logger},
		metricsCollector:     &mockMetricsCollector{logger: logger},
		alertManager:         &mockAlertManager{logger: logger},
		resourceManager:      &mockResourceManager{logger: logger},
		schedulingEngine:     &mockSchedulingEngine{logger: logger},
	}

	if err := coordinator.initializeStorage(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}
	if err := coordinator.initializeSecurity(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize security: %w", err)
	}
	if err := coordinator.initializeTraining(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize training: %w", err)
	}
	if err := coordinator.initializeMonitoring(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize monitoring: %w", err)
	}

	coordinator.startBackgroundServices()

	coordinator.wg.Add(1)
	go func() {
		defer coordinator.wg.Done()
		coordinator.handleConfigUpdates(configUpdateChan)
	}()

	return coordinator, nil
}

func (c *FLCoordinator) handleConfigUpdates(updateChan <-chan *CoordinatorConfig) {
	for {
		select {
		case newConfig := <-updateChan:
			c.logger.Info("Applying new configuration")
			c.applyNewConfig(newConfig)
		case <-c.ctx.Done():
			c.logger.Info("Configuration update handler stopped")
			return
		}
	}
}

func (c *FLCoordinator) applyNewConfig(config *CoordinatorConfig) {
	c.jobQueue.SetCapacity(config.MaxConcurrentJobs)
	c.logger.Info("Configuration updated successfully")
}


// initializeStorage initializes the storage backends.
func (c *FLCoordinator) initializeStorage() error {
	c.logger.Info("Initializing Redis-based storage for FL coordinator")
	c.clientStore = NewRedisClientStore(c.redisClient)
	c.modelStore = NewRedisModelStore(c.redisClient)
	return nil
}


// initializeSecurity initializes security components.
func (c *FLCoordinator) initializeSecurity() error {
	c.logger.Info("Initializing mock security components")
	return nil
}

// initializeTraining initializes training-related components.
func (c *FLCoordinator) initializeTraining() error {
	c.logger.Info("Initializing mock training components")
	return nil
}

// initializeMonitoring initializes monitoring and observability components.
func (c *FLCoordinator) initializeMonitoring() error {
	c.logger.Info("Initializing mock monitoring components")
	return nil
}

// startBackgroundServices starts any necessary background goroutines.
func (c *FLCoordinator) startBackgroundServices() {
	c.logger.Info("Starting FL coordinator background services")
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(c.config.CoordinatorConfig.HeartbeatInterval * 2)
		defer ticker.Stop()
		for {
			select {
			case <-c.ctx.Done():
				c.logger.Info("Background client cleanup service stopped")
				return
			case <-ticker.C:
				c.cleanupInactiveClients()
			}
		}
	}()

	// Start job queue processing
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.jobQueue.Start(c.ctx)
	}()
}

// cleanupInactiveClients removes clients that haven't sent heartbeats for a long time.
func (c *FLCoordinator) cleanupInactiveClients() {
	c.logger.Debug("Running inactive client cleanup")
	clients, err := c.clientStore.List(c.ctx)
	if err != nil {
		c.logger.Error("Failed to list clients for cleanup", slog.String("error", err.Error()))
		return
	}

	for _, client := range clients {
		if time.Since(client.LastHeartbeat) > c.config.MaxInactiveTime {
			c.logger.Info("Unregistering inactive client", slog.String("client_id", client.ID))
			if err := c.UnregisterClient(c.ctx, client.ID); err != nil {
				c.logger.Error("Failed to unregister inactive client",
					slog.String("client_id", client.ID), slog.String("error", err.Error()))
			}
		}
	}
}

// validateClientRegistration performs basic validation on client registration data.
func (c *FLCoordinator) validateClientRegistration(client *FLClient) error {
	if client.XAppName == "" {
		return fmt.Errorf("xapp_name cannot be empty")
	}
	if client.Endpoint == "" {
		return fmt.Errorf("endpoint cannot be empty")
	}
	if len(client.RRMTasks) == 0 {
		return fmt.Errorf("at least one RRM task is required")
	}
	// Further validation based on config.CoordinatorConfig
	return nil
}

// calculateInitialTrustScore calculates an initial trust score for a new client.
func (c *FLCoordinator) calculateInitialTrustScore(client *FLClient) float64 {
	// Placeholder: In a real system, this would involve more complex logic
	// based on client history, security features, and network conditions.
	return 0.75 // Default initial trust score
}

// validateNetworkSlices validates if the provided network slices are valid and accessible.
func (c *FLCoordinator) validateNetworkSlices(ctx context.Context, slices []NetworkSliceInfo) error {
	// Placeholder: In a real system, this would query a network slice manager
	// or A1 interface for valid slice IDs and permissions.
	if len(slices) > 0 {
		c.logger.Debug("Network slice validation (placeholder)", slog.Any("slices", slices))
	}
	return nil
}

// matchesSelector checks if a client matches the given selector criteria.
func (c *FLCoordinator) matchesSelector(client *FLClient, selector ClientSelector) bool {
	if selector.MinTrustScore > 0 && client.TrustScore < selector.MinTrustScore {
		return false
	}
	if selector.GeographicZone != "" && client.Metadata["geographic_zone"] != selector.GeographicZone {
		return false
	}
	for _, task := range selector.MatchRRMTasks {
		found := false
		for _, clientTask := range client.RRMTasks {
			if clientTask == task {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	// Add logic for MatchLabels, MinComputeCapacity, MaxLatency etc.
	return true
}

// selectTopClients selects the top N clients based on trust score.
func (c *FLCoordinator) selectTopClients(clients []*FLClient, count int) []*FLClient {
	// Simple sort by trust score (descending)
	// In a real system, this would be more sophisticated (e.g., considering diversity)
	sort.Slice(clients, func(i, j int) bool {
		return clients[i].TrustScore > clients[j].TrustScore
	})
	if len(clients) > count {
		return clients[:count]
	}
	return clients
}

// validateTrainingJobSpec validates the training job specification.
func (c *FLCoordinator) validateTrainingJobSpec(jobSpec TrainingJobSpec) error {
	if jobSpec.ModelID == "" {
		return fmt.Errorf("model ID cannot be empty")
	}
	if jobSpec.RRMTask == "" {
		return fmt.Errorf("RRM task cannot be empty")
	}
	if jobSpec.TrainingConfig.MinParticipants <= 0 {
		return fmt.Errorf("min_participants must be greater than 0")
	}
	// Add more comprehensive validation based on config
	return nil
}

// getOrCreateGlobalModel retrieves an existing global model or creates a new one.
func (c *FLCoordinator) getOrCreateGlobalModel(ctx context.Context, modelID string, rrmTask RRMTaskType) (*GlobalModel, error) {
	model, err := c.modelStore.GetGlobalModel(ctx, modelID)
	if err == nil {
		c.logger.Info("Found existing global model", slog.String("model_id", modelID))
		return model, nil
	}

	// If not found, create a new placeholder model
	c.logger.Info("Creating new placeholder global model", slog.String("model_id", modelID))
	newModel := &GlobalModel{
		ID:           modelID,
		Name:         fmt.Sprintf("model-%s", modelID),
		Version:      "v1.0.0",
		RRMTask:      rrmTask,
		Format:       ModelFormatTensorFlow, // Default
		Architecture: "CNN",                 // Default
		Parameters:   []byte("initial_model_parameters"),
		ParametersHash: hex.EncodeToString(sha256.New().Sum([]byte("initial_model_parameters"))),
		TrainingConfig: TrainingConfiguration{
			BatchSize:       32,
			LocalEpochs:     5,
			LearningRate:    0.01,
			MinParticipants: 2,
			MaxParticipants: 10,
			TimeoutSeconds:  300,
		},
		AggregationAlg: AggregationFedAvg,
		PrivacyMech:    PrivacyDifferential,
		PrivacyParams: PrivacyParameters{
			Epsilon: 1.0,
			Delta:   1e-5,
		},
		MaxRounds:      10,
		TargetAccuracy: 0.9,
		Status:         TrainingStatusInitializing,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := c.modelStore.CreateGlobalModel(ctx, newModel); err != nil {
		return nil, fmt.Errorf("failed to create new global model: %w", err)
	}
	return newModel, nil
}

// selectClientsForTraining selects clients based on job requirements and availability.
func (c *FLCoordinator) selectClientsForTraining(ctx context.Context, selector ClientSelector, rrmTask RRMTaskType) ([]*FLClient, error) {
	allClients, err := c.clientStore.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list clients for selection: %w", err)
	}

	var eligibleClients []*FLClient
	for _, client := range allClients {
		// Check if client is idle and matches RRM task
		if client.Status == FLClientStatusIdle {
			for _, task := range client.RRMTasks {
				if task == rrmTask {
					eligibleClients = append(eligibleClients, client)
					break
				}
			}
		}
	}

	// Apply selector filters
	var selectedClients []*FLClient
	for _, client := range eligibleClients {
		if c.matchesSelector(client, selector) {
			selectedClients = append(selectedClients, client)
		}
	}

	// Apply MaxClients limit
	if selector.MaxClients > 0 && len(selectedClients) > selector.MaxClients {
		selectedClients = c.selectTopClients(selectedClients, selector.MaxClients)
	}

	c.logger.Info("Clients selected for training",
		slog.Int("selected_count", len(selectedClients)),
		slog.Int("eligible_count", len(eligibleClients)))

	return selectedClients, nil
}

func (c *FLCoordinator) orchestrateTraining(ctx context.Context, activeJob *ActiveTrainingJob) {
	defer c.wg.Done()
	jobID := activeJob.Job.Name
	c.logger.Info("Starting training orchestration for job", slog.String("job_id", jobID))

	activeJob.Status = TrainingStatusRunning
	activeJob.Job.Status.Phase = TrainingStatusRunning
	activeJob.Job.Status.LastUpdate = metav1.Now()

	if c.role == MasterCoordinatorRole {
		c.orchestrateMaster(ctx, activeJob)
	} else {
		c.orchestrateRegional(ctx, activeJob)
	}
}

func (c *FLCoordinator) orchestrateMaster(ctx context.Context, activeJob *ActiveTrainingJob) {
	// 1. Discover regional coordinators
	regionalCoordinators, err := c.discoverRegionalCoordinators(ctx)
	if err != nil {
		c.logger.Error("Failed to discover regional coordinators", slog.String("error", err.Error()))
		return
	}

	// 2. Distribute training job to regional coordinators
	var wg sync.WaitGroup
	for _, regionalCoordinator := range regionalCoordinators {
		wg.Add(1)
		go func(regionalCoordinator *RegionalCoordinatorClient) {
			defer wg.Done()
			// Connect to regional coordinator
			conn, err := grpc.Dial(regionalCoordinator.Endpoint, grpc.WithInsecure())
			if err != nil {
				c.logger.Error("Failed to connect to regional coordinator", slog.String("endpoint", regionalCoordinator.Endpoint), slog.String("error", err.Error()))
				return
			}
			defer conn.Close()
			client := NewFederatedLearningClient(conn)

			// Start training on regional coordinator
			stream, err := client.HandleTraining(ctx)
			if err != nil {
				c.logger.Error("Failed to start training on regional coordinator", slog.String("endpoint", regionalCoordinator.Endpoint), slog.String("error", err.Error()))
				return
			}

			// Send training job to regional coordinator
			if err := stream.Send(&TrainingRequest{Request: &TrainingRequest_Registration{Registration: &ClientRegistration{
				XappName: activeJob.Job.Spec.RRMTask,
			}}}); err != nil {
				c.logger.Error("Failed to send training job to regional coordinator", slog.String("endpoint", regionalCoordinator.Endpoint), slog.String("error", err.Error()))
				return
			}

			// Wait for response from regional coordinator
			for {
				resp, err := stream.Recv()
				if err != nil {
					c.logger.Error("Failed to receive response from regional coordinator", slog.String("endpoint", regionalCoordinator.Endpoint), slog.String("error", err.Error()))
					return
				}
				// Process response
				switch resp.Response.(type) {
				case *TrainingResponse_Result:
					// Handle result
				}
			}
		}(regionalCoordinator)
	}
	wg.Wait()

	// 3. Aggregate results from regional coordinators
	// TODO: Implement aggregation logic
}

func (c *FLCoordinator) registerWithMaster(ctx context.Context) error {
	if c.role != RegionalCoordinatorRole {
		return nil
	}

	// Connect to master coordinator
	conn, err := grpc.Dial(c.config.CoordinatorConfig.MasterEndpoint, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to master coordinator: %w", err)
	}
	defer conn.Close()
	client := NewFederatedLearningClient(conn)

	// Register with master coordinator
	stream, err := client.HandleTraining(ctx)
	if err != nil {
		return fmt.Errorf("failed to register with master coordinator: %w", err)
	}

	if err := stream.Send(&TrainingRequest{Request: &TrainingRequest_Registration{Registration: &ClientRegistration{
		Endpoint: c.config.CoordinatorConfig.ListenAddress,
	}}}); err != nil {
		return fmt.Errorf("failed to send registration to master coordinator: %w", err)
	}

	return nil
}

func (c *FLCoordinator) orchestrateRegional(ctx context.Context, activeJob *ActiveTrainingJob) {
	defer c.wg.Done()
	jobID := activeJob.Job.Name
	c.logger.Info("Starting asynchronous training orchestration for job", slog.String("job_id", jobID))

	// Start the first round
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.startNewRound(ctx, activeJob, 1)
	}()

	for {
		select {
		case round := <-activeJob.roundCompleteChan:
			c.logger.Info("Round completed successfully", slog.String("job_id", jobID), slog.Int64("round", round))
			if activeJob.CurrentModel.ModelMetrics.Accuracy >= activeJob.Job.Spec.TrainingConfig.TargetAccuracy {
				c.logger.Info("Model has converged", slog.String("job_id", jobID), slog.Float64("accuracy", activeJob.CurrentModel.ModelMetrics.Accuracy))
				c.finishTrainingJob(activeJob, TrainingStatusCompleted, "Model converged successfully")
				return
			}

			if activeJob.CompletedRounds >= activeJob.Job.Spec.TrainingConfig.MaxRounds {
				c.logger.Info("Maximum rounds reached", slog.String("job_id", jobID))
				c.finishTrainingJob(activeJob, TrainingStatusCompleted, "Maximum training rounds reached")
				return
			}

			// Start the next round
			nextRound := activeJob.CompletedRounds + 1
			c.wg.Add(1)
			go func() {
				defer c.wg.Done()
				c.startNewRound(ctx, activeJob, nextRound)
			}()

		case err := <-activeJob.errorChan:
			c.logger.Error("A training round failed", slog.String("job_id", jobID), slog.String("error", err.Error()))
			activeJob.FailedRounds++
			// For simplicity, we fail the job on any round error. A more advanced implementation could have retry logic.
			c.finishTrainingJob(activeJob, TrainingStatusFailed, fmt.Sprintf("Training round failed: %v", err))
			return

		case <-ctx.Done():
			c.logger.Info("Training orchestration cancelled for job", slog.String("job_id", jobID))
			c.finishTrainingJob(activeJob, TrainingStatusPaused, "Training was cancelled")
			return
		}
	}
}

func (c *FLCoordinator) startNewRound(ctx context.Context, activeJob *ActiveTrainingJob, round int64) {
	c.logger.Info("Starting new training round", slog.String("job_id", activeJob.Job.Name), slog.Int64("round", round))
	activeJob.roundMutex.Lock()
	activeJob.RoundStartTime = time.Now()
	activeJob.collectedUpdates = make(map[string]ModelUpdate)
	activeJob.updateTimestamps = make(map[string]time.Time)
	activeJob.roundMutex.Unlock()

	// Distribute model to clients (mock implementation)
	for _, client := range activeJob.Participants {
		go func(client *FLClient) {
			c.metricsCollector.RecordClientStatusChange(client.ID, FLClientStatusIdle, FLClientStatusTraining)
			time.Sleep(time.Duration(1+rand.Intn(5)) * time.Second) // Simulate variable training time
			localMetrics := ModelMetrics{Accuracy: 0.8 + rand.Float64()*0.1}
			update := ModelUpdate{
				ClientID:         client.ID,
				ModelID:          activeJob.CurrentModel.ID,
				Round:            round,
				Parameters:       []byte(fmt.Sprintf("client_%s_model_update_round_%d", client.ID, round)),
				ParametersHash:   hex.EncodeToString(sha256.New().Sum([]byte(fmt.Sprintf("client_%s_model_update_round_%d", client.ID, round)))),
				DataSamplesCount: 1000,
				LocalMetrics:     localMetrics,
				SubmissionTime:   time.Now(),
			}
			activeJob.modelUpdates <- update
			c.metricsCollector.RecordClientStatusChange(client.ID, FLClientStatusTraining, FLClientStatusIdle)
		}(client)
	}

	// Start a goroutine to collect and process updates asynchronously for this round
	c.wg.Add(1)
	go c.collectAndAggregateAsync(ctx, activeJob, round)
}

func (c *FLCoordinator) collectAndAggregateAsync(ctx context.Context, activeJob *ActiveTrainingJob, round int64) {
	defer c.wg.Done()
	jobID := activeJob.Job.Name
	config := c.configManager.GetConfig()
	timeout := time.NewTimer(config.AggregationWindow)
	defer timeout.Stop()

	minParticipants := int(float64(len(activeJob.Participants)) * config.MinParticipationRatio)

	for {
		select {
		case update := <-activeJob.modelUpdates:
			activeJob.roundMutex.Lock()
			activeJob.collectedUpdates[update.ClientID] = update
			collectedCount := len(activeJob.collectedUpdates)
			activeJob.roundMutex.Unlock()

			c.logger.Debug("Collected model update",
				slog.String("job_id", jobID), slog.Int64("round", round),
				slog.String("client_id", update.ClientID), slog.Int("collected_count", collectedCount))

			if collectedCount >= minParticipants {
				c.logger.Info("Minimum participation met, triggering aggregation", slog.String("job_id", jobID), slog.Int64("round", round))
				c.triggerAggregation(ctx, activeJob, round)
				return // Stop collecting for this round
			}

		case <-timeout.C:
			c.logger.Warn("Aggregation window timed out, triggering aggregation",
				slog.String("job_id", jobID), slog.Int64("round", round), slog.Int("collected", len(activeJob.collectedUpdates)))
			c.triggerAggregation(ctx, activeJob, round)
			return // Stop collecting for this round

		case <-ctx.Done():
			c.logger.Info("Context cancelled during update collection", slog.String("job_id", jobID))
			return
		}
	}
}

func (c *FLCoordinator) triggerAggregation(ctx context.Context, activeJob *ActiveTrainingJob, round int64) {
	activeJob.roundMutex.RLock()
	updatesToAggregate := make([]ModelUpdate, 0, len(activeJob.collectedUpdates))
	for _, update := range activeJob.collectedUpdates {
		updatesToAggregate = append(updatesToAggregate, update)
	}
	activeJob.roundMutex.RUnlock()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		if err := c.processCollectedUpdates(ctx, activeJob, round, updatesToAggregate); err != nil {
			activeJob.errorChan <- err
		} else {
			activeJob.roundCompleteChan <- round
		}
	}()
}

func (c *FLCoordinator) processCollectedUpdates(ctx context.Context, activeJob *ActiveTrainingJob, round int64, updates []ModelUpdate) error {
	config := c.configManager.GetConfig()
	minParticipants := int(float64(len(activeJob.Participants)) * config.MinParticipationRatio)

	if len(updates) < minParticipants {
		return fmt.Errorf("insufficient model updates for aggregation in round %d (%d collected, %d required)",
			round, len(updates), minParticipants)
	}

	// Calculate staleness scores without filtering
	stalenessScores := make(map[string]float64)
	for _, update := range updates {
		staleness := time.Since(update.SubmissionTime)
		// De-emphasize stale updates by reducing their weight. A score <= 0 will effectively discard the update during weighted aggregation.
		stalenessScores[update.ClientID] = 1.0 - (float64(staleness) / float64(config.StalenessThreshold))
	}

	c.logger.Info("Aggregating models", slog.String("job_id", activeJob.Job.Name), slog.Int64("round", round))
	activeJob.Status = TrainingStatusAggregating
	activeJob.Job.Status.Phase = TrainingStatusAggregating
	activeJob.Job.Status.LastUpdate = metav1.Now()

	newGlobalModel, err := c.aggregationEngine.Aggregate(updates, activeJob.CurrentModel.AggregationAlg, stalenessScores)
	if err != nil {
		return fmt.Errorf("model aggregation failed in round %d: %w", round, err)
	}

	activeJob.roundMutex.Lock()
	activeJob.CurrentModel = newGlobalModel
	activeJob.CurrentModel.CurrentRound = round
	activeJob.CurrentModel.UpdatedAt = time.Now()
	activeJob.LastAggregation = time.Now()
	activeJob.CompletedRounds = round
	activeJob.CurrentModel.Status = TrainingStatusRunning
	activeJob.roundMutex.Unlock()

	if err := c.modelStore.UpdateGlobalModel(ctx, activeJob.CurrentModel); err != nil {
		c.logger.Error("Failed to update global model in store",
			slog.String("model_id", activeJob.CurrentModel.ID), slog.String("error", err.Error()))
		// Non-fatal error, continue with the new model in memory
	}

	roundMetrics := RoundMetrics{
		Round:             round,
		Timestamp:         time.Now(),
		ParticipatingClients: len(updates),
		ModelMetrics:      newGlobalModel.ModelMetrics,
		AggregationTimeMs: float64(time.Since(activeJob.LastAggregation).Milliseconds()),
	}
	c.metricsCollector.RecordRoundMetrics(roundMetrics)

	c.logger.Info("Aggregation successful",
		slog.String("job_id", activeJob.Job.Name), slog.Int64("round", round),
		slog.Float64("accuracy", newGlobalModel.ModelMetrics.Accuracy))

	return nil
}

func (c *FLCoordinator) finishTrainingJob(activeJob *ActiveTrainingJob, status TrainingStatus, message string) {
	activeJob.Status = status
	activeJob.Job.Status.Phase = status
	activeJob.Job.Status.Message = message
	activeJob.Job.Status.LastUpdate = metav1.Now()
	if activeJob.CurrentModel != nil {
		activeJob.CurrentModel.Status = status
	}
	c.metricsCollector.RecordTrainingJobCompletion(activeJob.Job)
	c.logger.Info("Training orchestration finished for job", slog.String("job_id", activeJob.Job.Name), slog.String("final_status", string(status)))

	c.jobsMutex.Lock()
	delete(c.activeJobs, activeJob.Job.Name)
	c.jobsMutex.Unlock()
}




// --- Mock Implementations for Interfaces ---

type mockCryptoEngine struct {
	logger *slog.Logger
	keys   map[string]ed25519.PrivateKey
	mutex  sync.RWMutex
}

func (m *mockCryptoEngine) StorePrivateKey(clientID string, privateKey ed25519.PrivateKey) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.keys == nil {
		m.keys = make(map[string]ed25519.PrivateKey)
	}
	m.keys[clientID] = privateKey
	m.logger.Debug("Mock CryptoEngine: Stored private key", slog.String("client_id", clientID))
	return nil
}

func (m *mockCryptoEngine) DeleteKeys(clientID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.keys, clientID)
	m.logger.Debug("Mock CryptoEngine: Deleted keys", slog.String("client_id", clientID))
	return nil
}

type mockPrivacyEngine struct {
	logger *slog.Logger
}

func (m *mockPrivacyEngine) ApplyDifferentialPrivacy(data []byte, epsilon float64) ([]byte, error) {
	m.logger.Debug("Mock PrivacyEngine: Applying differential privacy (no-op)", slog.Float64("epsilon", epsilon))
	return data, nil // No-op for mock
}

func (m *mockPrivacyEngine) ValidatePrivacyBudget(clientID string, requestedBudget float64) error {
	m.logger.Debug("Mock PrivacyEngine: Validating privacy budget (always true)",
		slog.String("client_id", clientID), slog.Float64("requested_budget", requestedBudget))
	return nil // Always valid for mock
}

type mockTrustManager struct {
	logger *slog.Logger
	scores map[string]float64
	mutex  sync.RWMutex
}

func (m *mockTrustManager) RegisterClient(ctx context.Context, client *FLClient) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.scores == nil {
		m.scores = make(map[string]float64)
	}
	m.scores[client.ID] = client.TrustScore
	m.logger.Debug("Mock TrustManager: Registered client", slog.String("client_id", client.ID))
	return nil
}

func (m *mockTrustManager) UnregisterClient(ctx context.Context, clientID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.scores, clientID)
	m.logger.Debug("Mock TrustManager: Unregistered client", slog.String("client_id", clientID))
	return nil
}

func (m *mockTrustManager) UpdateTrustScore(ctx context.Context, clientID string, score float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.scores[clientID] = score
	m.logger.Debug("Mock TrustManager: Updated trust score", slog.String("client_id", clientID), slog.Float64("score", score))
	return nil
}

func (m *mockTrustManager) GetTrustScore(ctx context.Context, clientID string) (float64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	score, ok := m.scores[clientID]
	if !ok {
		return 0, fmt.Errorf("trust score for client %s not found", clientID)
	}
	m.logger.Debug("Mock TrustManager: Got trust score", slog.String("client_id", clientID), slog.Float64("score", score))
	return score, nil
}

type mockTrainingOrchestrator struct {
	logger *slog.Logger
}

func (m *mockTrainingOrchestrator) StartRound(ctx context.Context, job *ActiveTrainingJob) error {
	m.logger.Debug("Mock TrainingOrchestrator: Starting round (no-op)", slog.String("job_id", job.Job.Name))
	return nil
}

func (m *mockTrainingOrchestrator) AggregateRound(ctx context.Context, updates []ModelUpdate) (*GlobalModel, error) {
	m.logger.Debug("Mock TrainingOrchestrator: Aggregating round (no-op)")
	return &GlobalModel{}, nil
}

func (m *mockTrainingOrchestrator) ValidateParticipants(ctx context.Context, participants []*FLClient) error {
	m.logger.Debug("Mock TrainingOrchestrator: Validating participants (always true)")
	return nil
}

type mockAggregationEngine struct {
	logger *slog.Logger
}

func (m *mockAggregationEngine) Aggregate(updates []ModelUpdate, algorithm AggregationAlgorithm, stalenessScores map[string]float64) (*GlobalModel, error) {
	m.logger.Debug("Mock AggregationEngine: Aggregating models", slog.String("algorithm", string(algorithm)), slog.Any("staleness_scores", stalenessScores))
	// Simulate aggregation: simple average of accuracies, weighted by staleness
	var totalWeightedAccuracy float64
	var totalWeight float64
	for _, update := range updates {
		weight := 1.0
		if score, ok := stalenessScores[update.ClientID]; ok {
			weight = score
		}
		totalWeightedAccuracy += update.LocalMetrics.Accuracy * weight
		totalWeight += weight
	}
	avgAccuracy := 0.0
	if totalWeight > 0 {
		avgAccuracy = totalWeightedAccuracy / totalWeight
	}

	return &GlobalModel{
		ModelMetrics: ModelMetrics{
			Accuracy: avgAccuracy,
			Loss:     0.1, // Mock loss
		},
	}, nil
}

func (m *mockAggregationEngine) ValidateUpdate(update ModelUpdate) error {
	m.logger.Debug("Mock AggregationEngine: Validating update (always true)", slog.String("client_id", update.ClientID))
	return nil
}

func (m *mockAggregationEngine) ApplyPrivacyMechanism(model *GlobalModel, mechanism PrivacyMechanism) error {
	m.logger.Debug("Mock AggregationEngine: Applying privacy mechanism (no-op)", slog.String("mechanism", string(mechanism)))
	return nil
}

type mockMetricsCollector struct {
	logger *slog.Logger
}

func (m *mockMetricsCollector) RecordClientRegistration(client *FLClient) {
	m.logger.Debug("Mock MetricsCollector: Client registered", slog.String("client_id", client.ID))
}

func (m *mockMetricsCollector) RecordClientUnregistration(client *FLClient) {
	m.logger.Debug("Mock MetricsCollector: Client unregistered", slog.String("client_id", client.ID))
}

func (m *mockMetricsCollector) RecordClientStatusChange(clientID string, oldStatus, newStatus FLClientStatus) {
	m.logger.Debug("Mock MetricsCollector: Client status change",
		slog.String("client_id", clientID), slog.String("old", string(oldStatus)), slog.String("new", string(newStatus)))
}

func (m *mockMetricsCollector) RecordTrainingJobStart(job *TrainingJob) {
	m.logger.Debug("Mock MetricsCollector: Training job started", slog.String("job_id", job.Name))
}

func (m *mockMetricsCollector) RecordRoundMetrics(metrics RoundMetrics) {
	m.logger.Debug("Mock MetricsCollector: Recorded round metrics", slog.Int64("round", metrics.Round))
}

func (m *mockMetricsCollector) RecordTrainingJobCompletion(job *TrainingJob) {
	m.logger.Debug("Mock MetricsCollector: Training job completed", slog.String("job_id", job.Name))
}

type mockAlertManager struct {
	logger *slog.Logger
}

func (m *mockAlertManager) SendAlert(alertType, message string, severity string) error {
	m.logger.Warn("Mock AlertManager: Sending alert",
		slog.String("type", alertType), slog.String("severity", severity), slog.String("message", message))
	return nil
}

func (m *mockAlertManager) ConfigureThresholds(thresholds map[string]float64) error {
	m.logger.Debug("Mock AlertManager: Configuring thresholds (no-op)", slog.Any("thresholds", thresholds))
	return nil
}

type mockResourceManager struct {
	logger *slog.Logger
}

func (m *mockResourceManager) CheckAvailability(ctx context.Context, requirements ResourceRequirements) error {
	m.logger.Debug("Mock ResourceManager: Checking resource availability (always available)")
	return nil // Always available for mock
}

func (m *mockResourceManager) AllocateResources(ctx context.Context, jobID string, requirements ResourceRequirements) error {
	m.logger.Debug("Mock ResourceManager: Allocating resources (no-op)", slog.String("job_id", jobID))
	return nil
}

func (m *mockResourceManager) ReleaseResources(ctx context.Context, jobID string) error {
	m.logger.Debug("Mock ResourceManager: Releasing resources (no-op)", slog.String("job_id", jobID))
	return nil
}

type mockSchedulingEngine struct {
	logger *slog.Logger
}

func (m *mockSchedulingEngine) ScheduleJob(ctx context.Context, job *TrainingJob) error {
	m.logger.Debug("Mock SchedulingEngine: Scheduling job (no-op)", slog.String("job_id", job.Name))
	return nil
}

func (m *mockSchedulingEngine) CancelJob(ctx context.Context, jobID string) error {
	m.logger.Debug("Mock SchedulingEngine: Cancelling job (no-op)", slog.String("job_id", jobID))
	return nil
}

func (m *mockSchedulingEngine) GetScheduledJobs(ctx context.Context) ([]*TrainingJob, error) {
	m.logger.Debug("Mock SchedulingEngine: Getting scheduled jobs (empty list)")
	return []*TrainingJob{}, nil
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

	if err := c.validateTrainingJobSpec(jobSpec); err != nil {
		return nil, errors.NewErrorBuilder(errors.ErrInvalidInput).
			WithDetail("validation_error", err.Error()).
			WithCaller().
			Build()
	}

	if err := c.resourceManager.CheckAvailability(ctx, jobSpec.Resources); err != nil {
		return nil, errors.NewErrorBuilder(errors.ErrServiceUnavailable).
			WithDetail("resource_error", err.Error()).
			WithCaller().
			Build()
	}

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

	c.jobQueue.Submit(job, func() {
		c.executeTrainingJob(context.Background(), job)
	})

	c.logger.Info("Successfully queued FL training job",
		slog.String("job_id", job.Name),
		slog.String("model_id", jobSpec.ModelID))

	return job, nil
}

func (c *FLCoordinator) executeTrainingJob(ctx context.Context, job *TrainingJob) {
	globalModel, err := c.getOrCreateGlobalModel(ctx, job.Spec.ModelID, job.Spec.RRMTask)
	if err != nil {
		c.logger.Error("Failed to get global model", slog.String("error", err.Error()))
		return
	}

	selectedClients, err := c.selectClientsForTraining(ctx, job.Spec.ClientSelector, job.Spec.RRMTask)
	if err != nil {
		c.logger.Error("Failed to select clients", slog.String("error", err.Error()))
		return
	}

	if len(selectedClients) < job.Spec.TrainingConfig.MinParticipants {
		c.logger.Error("Insufficient participants",
			slog.Int("required", job.Spec.TrainingConfig.MinParticipants),
			slog.Int("available", len(selectedClients)))
		return
	}

	activeJob := &ActiveTrainingJob{
		Job:              job,
		Participants:     make(map[string]*FLClient),
		CurrentModel:     globalModel,
		RoundStartTime:   time.Now(),
		LastAggregation:  time.Now(),
		Status:           TrainingStatusInitializing,
		stalenessScores:  make(map[string]float64),
		updateTimestamps: make(map[string]time.Time),
		modelUpdates:     make(chan ModelUpdate, len(selectedClients)*2),
		statusUpdates:    make(chan TrainingStatusUpdate, len(selectedClients)*2),
		errorChan:        make(chan error, len(selectedClients)),
		aggregationTrigger: make(chan struct{}, 1),
		roundCompleteChan: make(chan int64, 1),
		collectedUpdates: make(map[string]ModelUpdate),
	}

	for _, client := range selectedClients {
		activeJob.Participants[client.ID] = client
	}

	c.jobsMutex.Lock()
	c.activeJobs[job.Name] = activeJob
	c.jobsMutex.Unlock()

	c.wg.Add(1)
	go c.orchestrateTraining(ctx, activeJob)

	job.Status.Phase = TrainingStatusWaiting
	job.Status.ParticipatingClients = make([]string, 0, len(selectedClients))
	for clientID := range activeJob.Participants {
		job.Status.ParticipatingClients = append(job.Status.ParticipatingClients, clientID)
	}
	job.Status.LastUpdate = metav1.Now()

	c.metricsCollector.RecordTrainingJobStart(job)
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
		close(job.aggregationTrigger)
		job.collectedUpdates = nil
		job.updateTimestamps = nil
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

func (c *FLCoordinator) HandleTraining(stream FederatedLearning_HandleTrainingServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		switch req.Request.(type) {
		case *TrainingRequest_Registration:
			// Handle registration
		case *TrainingRequest_Update:
			// Handle model update
		case *TrainingRequest_Heartbeat:
			// Handle heartbeat
		}
	}
}