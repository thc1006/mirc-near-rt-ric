package o1

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hctsai1006/near-rt-ric/internal/config"
	"github.com/hctsai1006/near-rt-ric/pkg/common/monitoring"
	"github.com/sirupsen/logrus"
)

// O1Interface represents the main O1 interface implementation
// Complies with O-RAN.WG10.O1-Interface.0-v08.00 specification
type O1Interface struct {
	config  *config.O1Config
	logger  *logrus.Logger
	metrics *monitoring.MetricsCollector

	// Core components
	netconfServer *NetconfServer
	fcapsManager  *FCAPSManager

	// Control and synchronization
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running bool
	mutex   sync.RWMutex

	// Event handlers
	eventHandlers []O1InterfaceEventHandler

	// Statistics
	startTime time.Time
}

// O1InterfaceEventHandler defines interface for O1 interface events
type O1InterfaceEventHandler interface {
	OnO1InterfaceStarted()
	OnO1InterfaceStopped()
	OnNetconfSessionCreated(session *NetconfSession)
	OnNetconfSessionClosed(sessionID uint32)
	OnConfigurationChanged(change *ConfigurationChange)
	OnAlarmRaised(alarm *Alarm)
	OnAlarmCleared(alarm *Alarm)
	OnPerformanceReportGenerated(report *PerformanceReport)
	OnError(err error)
}

// NewO1Interface creates a new O1 interface instance
func NewO1Interface(cfg *config.O1Config, logger *logrus.Logger, metrics *monitoring.MetricsCollector) (*O1Interface, error) {
	ctx, cancel := context.WithCancel(context.Background())

	o1 := &O1Interface{
		config:    cfg,
		logger:    logger.WithField("component", "o1-interface"),
		metrics:   metrics,
		ctx:       ctx,
		cancel:    cancel,
		startTime: time.Now(),
	}

	// Initialize NETCONF server
	var err error
	o1.netconfServer, err = NewNetconfServer(cfg, logger, metrics)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create NETCONF server: %w", err)
	}

	// Initialize FCAPS manager
	o1.fcapsManager = NewFCAPSManager(cfg, logger, metrics)

	// Set up message handlers
	o1.netconfServer.SetMessageHandler(o1)

	// Set up event handlers
	o1.fcapsManager.AddEventHandler(o1)

	o1.logger.Info("O1 interface initialized successfully")
	return o1, nil
}

// Start starts the O1 interface
func (o1 *O1Interface) Start(ctx context.Context) error {
	o1.mutex.Lock()
	defer o1.mutex.Unlock()

	if o1.running {
		return fmt.Errorf("O1 interface is already running")
	}

	o1.logger.WithFields(logrus.Fields{
		"netconf_port":     o1.config.NETCONF.Port,
		"netconf_tls_port": o1.config.NETCONF.TLSPort,
		"tls_enabled":      o1.config.NETCONF.TLS.Enabled,
	}).Info("Starting O-RAN O1 interface")

	// Start FCAPS manager first
	if err := o1.fcapsManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start FCAPS manager: %w", err)
	}

	// Start NETCONF server
	if err := o1.netconfServer.Start(ctx); err != nil {
		o1.fcapsManager.Stop(ctx)
		return fmt.Errorf("failed to start NETCONF server: %w", err)
	}

	// Start background monitoring
	o1.wg.Add(1)
	go o1.monitoringWorker()

	o1.running = true
	o1.logger.Info("O-RAN O1 interface started successfully")

	// Send event
	for _, handler := range o1.eventHandlers {
		go handler.OnO1InterfaceStarted()
	}

	return nil
}

// Stop stops the O1 interface gracefully
func (o1 *O1Interface) Stop(ctx context.Context) error {
	o1.mutex.Lock()
	defer o1.mutex.Unlock()

	if !o1.running {
		return nil
	}

	o1.logger.Info("Stopping O-RAN O1 interface")

	// Stop NETCONF server
	if err := o1.netconfServer.Stop(ctx); err != nil {
		o1.logger.WithError(err).Error("Error stopping NETCONF server")
	}

	// Stop FCAPS manager
	if err := o1.fcapsManager.Stop(ctx); err != nil {
		o1.logger.WithError(err).Error("Error stopping FCAPS manager")
	}

	// Cancel context and wait for workers
	o1.cancel()

	done := make(chan struct{})
	go func() {
		o1.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		o1.logger.Info("O-RAN O1 interface stopped successfully")
	case <-ctx.Done():
		o1.logger.Warn("O1 interface shutdown timeout")
	}

	o1.running = false

	// Send event
	for _, handler := range o1.eventHandlers {
		go handler.OnO1InterfaceStopped()
	}

	return nil
}

// NetconfMessageHandler interface implementation

// HandleGetConfig handles NETCONF get-config operations
func (o1 *O1Interface) HandleGetConfig(sessionID uint32, req *GetConfigRequest) (interface{}, error) {
	o1.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"datastore":  req.Datastore,
		"filter":     req.Filter != "",
	}).Debug("Processing get-config request")

	// Get configuration from FCAPS manager
	config, err := o1.fcapsManager.GetConfig(req.Datastore, req.Filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	// Update metrics
	o1.metrics.O1Metrics.NetconfOperations.WithLabelValues("get-config", "success").Inc()

	return config, nil
}

// HandleEditConfig handles NETCONF edit-config operations
func (o1 *O1Interface) HandleEditConfig(sessionID uint32, req *EditConfigRequest) error {
	o1.logger.WithFields(logrus.Fields{
		"session_id":        sessionID,
		"datastore":         req.Datastore,
		"default_operation": req.DefaultOperation,
	}).Debug("Processing edit-config request")

	// Apply configuration changes through FCAPS manager
	err := o1.fcapsManager.EditConfig(sessionID, req.Datastore, req.DefaultOperation, req.Config)
	if err != nil {
		o1.metrics.O1Metrics.NetconfOperations.WithLabelValues("edit-config", "failed").Inc()
		return fmt.Errorf("failed to edit config: %w", err)
	}

	// Update metrics
	o1.metrics.O1Metrics.NetconfOperations.WithLabelValues("edit-config", "success").Inc()

	return nil
}

// HandleGet handles NETCONF get operations
func (o1 *O1Interface) HandleGet(sessionID uint32, filter string) (interface{}, error) {
	o1.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"filter":     filter != "",
	}).Debug("Processing get request")

	// Get operational data
	data := o1.getOperationalData(filter)

	// Update metrics
	o1.metrics.O1Metrics.NetconfOperations.WithLabelValues("get", "success").Inc()

	return data, nil
}

// HandleCopyConfig handles NETCONF copy-config operations
func (o1 *O1Interface) HandleCopyConfig(sessionID uint32, source, target DatastoreType) error {
	o1.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"source":     source,
		"target":     target,
	}).Debug("Processing copy-config request")

	err := o1.fcapsManager.CopyConfig(source, target)
	if err != nil {
		o1.metrics.O1Metrics.NetconfOperations.WithLabelValues("copy-config", "failed").Inc()
		return fmt.Errorf("failed to copy config: %w", err)
	}

	// Update metrics
	o1.metrics.O1Metrics.NetconfOperations.WithLabelValues("copy-config", "success").Inc()

	return nil
}

// HandleDeleteConfig handles NETCONF delete-config operations
func (o1 *O1Interface) HandleDeleteConfig(sessionID uint32, target DatastoreType) error {
	o1.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"target":     target,
	}).Debug("Processing delete-config request")

	// Only allow deletion of candidate datastore
	if target != DatastoreCandidate {
		return fmt.Errorf("delete-config only allowed for candidate datastore")
	}

	err := o1.fcapsManager.CopyConfig(DatastoreRunning, DatastoreCandidate)
	if err != nil {
		o1.metrics.O1Metrics.NetconfOperations.WithLabelValues("delete-config", "failed").Inc()
		return fmt.Errorf("failed to delete config: %w", err)
	}

	// Update metrics
	o1.metrics.O1Metrics.NetconfOperations.WithLabelValues("delete-config", "success").Inc()

	return nil
}

// HandleLock handles NETCONF lock operations
func (o1 *O1Interface) HandleLock(sessionID uint32, target DatastoreType) error {
	o1.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"target":     target,
	}).Debug("Processing lock request")

	// TODO: Implement proper datastore locking
	// For now, just acknowledge the lock

	// Update metrics
	o1.metrics.O1Metrics.NetconfOperations.WithLabelValues("lock", "success").Inc()

	return nil
}

// HandleUnlock handles NETCONF unlock operations
func (o1 *O1Interface) HandleUnlock(sessionID uint32, target DatastoreType) error {
	o1.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"target":     target,
	}).Debug("Processing unlock request")

	// TODO: Implement proper datastore unlocking
	// For now, just acknowledge the unlock

	// Update metrics
	o1.metrics.O1Metrics.NetconfOperations.WithLabelValues("unlock", "success").Inc()

	return nil
}

// HandleCloseSession handles NETCONF close-session operations
func (o1 *O1Interface) HandleCloseSession(sessionID uint32) error {
	o1.logger.WithField("session_id", sessionID).Debug("Processing close-session request")

	// Send event
	for _, handler := range o1.eventHandlers {
		go handler.OnNetconfSessionClosed(sessionID)
	}

	// Update metrics
	o1.metrics.O1Metrics.NetconfOperations.WithLabelValues("close-session", "success").Inc()

	return nil
}

// HandleKillSession handles NETCONF kill-session operations
func (o1 *O1Interface) HandleKillSession(sessionID uint32, targetSessionID uint32) error {
	o1.logger.WithFields(logrus.Fields{
		"session_id":        sessionID,
		"target_session_id": targetSessionID,
	}).Debug("Processing kill-session request")

	// TODO: Implement proper session killing
	// For now, just acknowledge the kill-session

	// Update metrics
	o1.metrics.O1Metrics.NetconfOperations.WithLabelValues("kill-session", "success").Inc()

	return nil
}

// FCAPSEventHandler interface implementation

func (o1 *O1Interface) OnAlarmRaised(alarm *Alarm) {
	o1.logger.WithFields(logrus.Fields{
		"alarm_id":       alarm.AlarmID,
		"alarm_type":     alarm.AlarmType,
		"managed_object": alarm.ManagedObject,
		"severity":       alarm.Severity,
	}).Info("Alarm raised")

	for _, handler := range o1.eventHandlers {
		go handler.OnAlarmRaised(alarm)
	}
}

func (o1 *O1Interface) OnAlarmCleared(alarm *Alarm) {
	o1.logger.WithFields(logrus.Fields{
		"alarm_id":       alarm.AlarmID,
		"alarm_type":     alarm.AlarmType,
		"managed_object": alarm.ManagedObject,
	}).Info("Alarm cleared")

	for _, handler := range o1.eventHandlers {
		go handler.OnAlarmCleared(alarm)
	}
}

func (o1 *O1Interface) OnConfigurationChanged(change *ConfigurationChange) {
	o1.logger.WithFields(logrus.Fields{
		"change_id":  change.ChangeID,
		"session_id": change.SessionID,
		"datastore":  change.Datastore,
		"operation":  change.Operation,
	}).Info("Configuration changed")

	for _, handler := range o1.eventHandlers {
		go handler.OnConfigurationChanged(change)
	}
}

func (o1 *O1Interface) OnPerformanceReportGenerated(report *PerformanceReport) {
	o1.logger.WithFields(logrus.Fields{
		"report_id":      report.ReportID,
		"managed_object": report.ManagedObject,
		"report_type":    report.ReportType,
		"metrics_count":  len(report.Metrics),
	}).Info("Performance report generated")

	for _, handler := range o1.eventHandlers {
		go handler.OnPerformanceReportGenerated(report)
	}
}

func (o1 *O1Interface) OnSecurityPolicyUpdated(policy *SecurityPolicy) {
	o1.logger.WithFields(logrus.Fields{
		"policy_id":   policy.PolicyID,
		"policy_name": policy.PolicyName,
		"policy_type": policy.PolicyType,
	}).Info("Security policy updated")
}

func (o1 *O1Interface) OnSoftwareInstalled(software *SoftwareVersion) {
	o1.logger.WithFields(logrus.Fields{
		"name":    software.Name,
		"version": software.Version,
		"vendor":  software.Vendor,
	}).Info("Software installed")
}

func (o1 *O1Interface) OnFileTransferCompleted(transfer *FileTransfer) {
	o1.logger.WithFields(logrus.Fields{
		"transfer_id": transfer.TransferID,
		"operation":   transfer.Operation,
		"status":      transfer.Status,
	}).Info("File transfer completed")
}

// Background workers

func (o1 *O1Interface) monitoringWorker() {
	defer o1.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-o1.ctx.Done():
			return
		case <-ticker.C:
			o1.collectInterfaceMetrics()
		}
	}
}

func (o1 *O1Interface) collectInterfaceMetrics() {
	// Collect NETCONF server metrics
	netconfStats := o1.netconfServer.GetStats()
	o1.metrics.O1Metrics.NetconfSessions.Set(float64(netconfStats.ActiveSessions))
	o1.metrics.O1Metrics.NetconfMessages.Add(float64(netconfStats.TotalMessages))

	// Collect system resource metrics
	o1.fcapsManager.CollectPerformanceMetric("o1_interface_uptime", "o1-interface", "gauge",
		time.Since(o1.startTime).Seconds(), "seconds", nil)

	// Log statistics
	o1.logger.WithFields(logrus.Fields{
		"active_sessions":    netconfStats.ActiveSessions,
		"total_sessions":     netconfStats.TotalSessions,
		"total_messages":     netconfStats.TotalMessages,
		"successful_rpcs":    netconfStats.SuccessfulRPCs,
		"failed_rpcs":        netconfStats.FailedRPCs,
		"active_alarms":      len(o1.fcapsManager.GetActiveAlarms()),
		"uptime":             time.Since(o1.startTime),
	}).Debug("O1 interface statistics")
}

// Helper methods

func (o1 *O1Interface) getOperationalData(filter string) map[string]interface{} {
	data := map[string]interface{}{
		"ietf-netconf-monitoring:netconf-state": map[string]interface{}{
			"capabilities": map[string]interface{}{
				"capability": o1.getCapabilities(),
			},
			"sessions": map[string]interface{}{
				"session": o1.getSessionsData(),
			},
			"statistics": map[string]interface{}{
				"netconf-start-time": o1.startTime.Format(time.RFC3339),
				"in-rpcs":            o1.netconfServer.GetStats().TotalMessages,
				"in-bad-rpcs":        o1.netconfServer.GetStats().FailedRPCs,
				"out-rpc-errors":     o1.netconfServer.GetStats().FailedRPCs,
				"out-notifications":  0,
			},
		},
		"o-ran-sc-ric:ric-state": map[string]interface{}{
			"alarms": o1.getAlarmsData(),
			"performance": map[string]interface{}{
				"metrics": o1.getPerformanceData(),
			},
			"software": o1.fcapsManager.GetSoftwareInventory(),
		},
	}

	// Apply filter if specified (simplified implementation)
	if filter != "" {
		return o1.applyFilter(data, filter)
	}

	return data
}

func (o1 *O1Interface) getCapabilities() []string {
	var caps []string
	for _, cap := range o1.netconfServer.capabilities {
		caps = append(caps, cap.URI)
	}
	return caps
}

func (o1 *O1Interface) getSessionsData() []map[string]interface{} {
	var sessions []map[string]interface{}
	netconfStats := o1.netconfServer.GetStats()
	
	// Return basic session statistics
	if netconfStats.ActiveSessions > 0 {
		sessions = append(sessions, map[string]interface{}{
			"session-id":     1,
			"transport":      "netconf-ssh",
			"username":       "netconf",
			"source-host":    "127.0.0.1",
			"login-time":     time.Now().Format(time.RFC3339),
			"in-rpcs":        netconfStats.TotalMessages,
			"in-bad-rpcs":    netconfStats.FailedRPCs,
			"out-rpc-errors": netconfStats.FailedRPCs,
		})
	}
	
	return sessions
}

func (o1 *O1Interface) getAlarmsData() []map[string]interface{} {
	var alarms []map[string]interface{}
	for _, alarm := range o1.fcapsManager.GetActiveAlarms() {
		alarms = append(alarms, map[string]interface{}{
			"alarm-id":         alarm.AlarmID,
			"alarm-type":       alarm.AlarmType,
			"managed-object":   alarm.ManagedObject,
			"severity":         string(alarm.Severity),
			"status":           string(alarm.Status),
			"probable-cause":   alarm.ProbableCause,
			"specific-problem": alarm.SpecificProblem,
			"raised-time":      alarm.RaisedTime.Format(time.RFC3339),
		})
	}
	return alarms
}

func (o1 *O1Interface) getPerformanceData() []map[string]interface{} {
	var metrics []map[string]interface{}
	// Return basic performance statistics
	netconfStats := o1.netconfServer.GetStats()
	
	metrics = append(metrics, map[string]interface{}{
		"metric-name":     "netconf-sessions",
		"metric-value":    netconfStats.ActiveSessions,
		"metric-unit":     "count",
		"timestamp":       time.Now().Format(time.RFC3339),
	})
	
	metrics = append(metrics, map[string]interface{}{
		"metric-name":     "netconf-messages",
		"metric-value":    netconfStats.TotalMessages,
		"metric-unit":     "count",
		"timestamp":       time.Now().Format(time.RFC3339),
	})
	
	return metrics
}

func (o1 *O1Interface) applyFilter(data map[string]interface{}, filter string) map[string]interface{} {
	// Simplified filter implementation
	// In production, implement proper XPath filtering
	return data
}

// Public interface methods

// IsRunning returns whether the O1 interface is running
func (o1 *O1Interface) IsRunning() bool {
	o1.mutex.RLock()
	defer o1.mutex.RUnlock()
	return o1.running
}

// AddEventHandler adds an event handler
func (o1 *O1Interface) AddEventHandler(handler O1InterfaceEventHandler) {
	o1.eventHandlers = append(o1.eventHandlers, handler)
}

// GetStats returns O1 interface statistics
func (o1 *O1Interface) GetStats() map[string]interface{} {
	netconfStats := o1.netconfServer.GetStats()
	
	return map[string]interface{}{
		"running":              o1.IsRunning(),
		"uptime":               time.Since(o1.startTime),
		"netconf_stats":        netconfStats,
		"active_alarms":        len(o1.fcapsManager.GetActiveAlarms()),
		"software_inventory":   o1.fcapsManager.GetSoftwareInventory(),
		"file_transfers":       len(o1.fcapsManager.GetFileTransfers()),
	}
}

// HealthCheck performs a health check of the O1 interface
func (o1 *O1Interface) HealthCheck() error {
	if !o1.IsRunning() {
		return fmt.Errorf("O1 interface is not running")
	}
	
	if !o1.netconfServer.IsRunning() {
		return fmt.Errorf("NETCONF server is not running")
	}
	
	return nil
}

// GetAlarms returns current alarms
func (o1 *O1Interface) GetAlarms(filter map[string]interface{}) []*Alarm {
	return o1.fcapsManager.GetAlarms(filter)
}

// RaiseAlarm raises a new alarm
func (o1 *O1Interface) RaiseAlarm(alarmType, managedObject, source string, severity AlarmSeverity, probableCause, specificProblem, additionalText string) (*Alarm, error) {
	return o1.fcapsManager.RaiseAlarm(alarmType, managedObject, source, severity, probableCause, specificProblem, additionalText)
}

// ClearAlarm clears an existing alarm
func (o1 *O1Interface) ClearAlarm(alarmID string) error {
	return o1.fcapsManager.ClearAlarm(alarmID)
}

// GetSoftwareInventory returns software inventory
func (o1 *O1Interface) GetSoftwareInventory() *SoftwareInventory {
	return o1.fcapsManager.GetSoftwareInventory()
}

// InstallSoftware installs new software
func (o1 *O1Interface) InstallSoftware(name, version, vendor, description string, dependencies []string, checksums map[string]string) (*SoftwareVersion, error) {
	return o1.fcapsManager.InstallSoftware(name, version, vendor, description, dependencies, checksums)
}

// StartFileTransfer starts a file transfer
func (o1 *O1Interface) StartFileTransfer(operation, localPath, remotePath, protocol string) (*FileTransfer, error) {
	return o1.fcapsManager.StartFileTransfer(operation, localPath, remotePath, protocol)
}