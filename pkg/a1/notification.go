package a1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// A1NotificationServiceImpl implements the A1 notification service
type A1NotificationServiceImpl struct {
	webhooks       map[string]string // policyID -> webhook URL
	httpClient     *http.Client
	logger         *logrus.Logger
	mutex          sync.RWMutex
	retryAttempts  int
	retryDelay     time.Duration
	timeout        time.Duration
}

// NewA1NotificationService creates a new A1 notification service
func NewA1NotificationService() A1NotificationService {
	return &A1NotificationServiceImpl{
		webhooks: make(map[string]string),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:        logrus.WithField("component", "a1-notification").Logger,
		retryAttempts: 3,
		retryDelay:    2 * time.Second,
		timeout:       30 * time.Second,
	}
}

// SendNotification sends a notification to registered webhooks
func (ns *A1NotificationServiceImpl) SendNotification(notification *A1Notification) error {
	ns.mutex.RLock()
	webhookURL, exists := ns.webhooks[notification.PolicyID]
	ns.mutex.RUnlock()

	if !exists {
		// No webhook registered for this policy - log and continue
		ns.logger.WithField("policy_id", notification.PolicyID).Debug("No webhook registered for policy")
		return nil
	}

	// Send notification with retry logic
	return ns.sendNotificationWithRetry(notification, webhookURL)
}

// RegisterWebhook registers a webhook URL for a specific policy
func (ns *A1NotificationServiceImpl) RegisterWebhook(policyID string, webhookURL string) error {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	// Validate webhook URL format
	if webhookURL == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	// Basic URL validation
	if len(webhookURL) < 7 || (webhookURL[:7] != "http://" && webhookURL[:8] != "https://") {
		return fmt.Errorf("webhook URL must start with http:// or https://")
	}

	ns.webhooks[policyID] = webhookURL
	
	ns.logger.WithFields(logrus.Fields{
		"policy_id":   policyID,
		"webhook_url": webhookURL,
	}).Info("Webhook registered")

	return nil
}

// UnregisterWebhook removes a webhook registration for a specific policy
func (ns *A1NotificationServiceImpl) UnregisterWebhook(policyID string) error {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	if _, exists := ns.webhooks[policyID]; !exists {
		return fmt.Errorf("no webhook registered for policy %s", policyID)
	}

	delete(ns.webhooks, policyID)
	
	ns.logger.WithField("policy_id", policyID).Info("Webhook unregistered")

	return nil
}

// ListWebhooks returns all registered webhooks
func (ns *A1NotificationServiceImpl) ListWebhooks() (map[string]string, error) {
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()

	// Return a copy to prevent external modifications
	webhooksCopy := make(map[string]string)
	for policyID, webhookURL := range ns.webhooks {
		webhooksCopy[policyID] = webhookURL
	}

	return webhooksCopy, nil
}

// sendNotificationWithRetry sends a notification with retry logic
func (ns *A1NotificationServiceImpl) sendNotificationWithRetry(notification *A1Notification, webhookURL string) error {
	var lastErr error

	for attempt := 0; attempt < ns.retryAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retrying
			time.Sleep(ns.retryDelay * time.Duration(attempt))
		}

		err := ns.sendNotificationHTTP(notification, webhookURL)
		if err == nil {
			ns.logger.WithFields(logrus.Fields{
				"notification_id": notification.NotificationID,
				"policy_id":       notification.PolicyID,
				"webhook_url":     webhookURL,
				"attempt":         attempt + 1,
			}).Info("Notification sent successfully")
			return nil
		}

		lastErr = err
		ns.logger.WithFields(logrus.Fields{
			"notification_id": notification.NotificationID,
			"policy_id":       notification.PolicyID,
			"webhook_url":     webhookURL,
			"attempt":         attempt + 1,
			"error":           err,
		}).Warn("Notification sending failed, will retry")
	}

	ns.logger.WithFields(logrus.Fields{
		"notification_id": notification.NotificationID,
		"policy_id":       notification.PolicyID,
		"webhook_url":     webhookURL,
		"attempts":        ns.retryAttempts,
		"error":           lastErr,
	}).Error("Failed to send notification after all retry attempts")

	return fmt.Errorf("failed to send notification after %d attempts: %w", ns.retryAttempts, lastErr)
}

// sendNotificationHTTP sends a single HTTP notification
func (ns *A1NotificationServiceImpl) sendNotificationHTTP(notification *A1Notification, webhookURL string) error {
	// Create HTTP request body
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Create HTTP request with context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), ns.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "O-RAN-Near-RT-RIC-A1/1.0")
	req.Header.Set("X-A1-Notification-ID", notification.NotificationID)
	req.Header.Set("X-A1-Policy-ID", notification.PolicyID)
	req.Header.Set("X-A1-Notification-Type", string(notification.NotificationType))

	// Send HTTP request
	resp, err := ns.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}

// GetNotificationStatistics returns statistics about notification sending
func (ns *A1NotificationServiceImpl) GetNotificationStatistics() map[string]interface{} {
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()

	return map[string]interface{}{
		"registered_webhooks": len(ns.webhooks),
		"retry_attempts":      ns.retryAttempts,
		"retry_delay_ms":      ns.retryDelay.Milliseconds(),
		"timeout_ms":          ns.timeout.Milliseconds(),
	}
}

// UpdateConfiguration updates the notification service configuration
func (ns *A1NotificationServiceImpl) UpdateConfiguration(retryAttempts int, retryDelay, timeout time.Duration) {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	ns.retryAttempts = retryAttempts
	ns.retryDelay = retryDelay
	ns.timeout = timeout

	// Update HTTP client timeout
	ns.httpClient.Timeout = timeout

	ns.logger.WithFields(logrus.Fields{
		"retry_attempts": retryAttempts,
		"retry_delay":    retryDelay,
		"timeout":        timeout,
	}).Info("Notification service configuration updated")
}

// TestWebhook tests a webhook URL by sending a test notification
func (ns *A1NotificationServiceImpl) TestWebhook(webhookURL string) error {
	testNotification := &A1Notification{
		NotificationID:   fmt.Sprintf("test-%d", time.Now().Unix()),
		PolicyID:         "test-policy",
		PolicyTypeID:     "test-policy-type",
		NotificationType: A1NotificationPolicyCreated,
		Timestamp:        time.Now(),
		Data: map[string]interface{}{
			"test": true,
			"message": "This is a test notification from O-RAN Near-RT RIC A1 interface",
		},
	}

	return ns.sendNotificationHTTP(testNotification, webhookURL)
}

// BatchNotificationSender handles sending multiple notifications efficiently
type BatchNotificationSender struct {
	service       *A1NotificationServiceImpl
	batchSize     int
	batchTimeout  time.Duration
	notifications chan *A1Notification
	done          chan struct{}
	logger        *logrus.Logger
}

// NewBatchNotificationSender creates a new batch notification sender
func NewBatchNotificationSender(service *A1NotificationServiceImpl, batchSize int, batchTimeout time.Duration) *BatchNotificationSender {
	return &BatchNotificationSender{
		service:       service,
		batchSize:     batchSize,
		batchTimeout:  batchTimeout,
		notifications: make(chan *A1Notification, batchSize*2),
		done:          make(chan struct{}),
		logger:        logrus.WithField("component", "batch-notification-sender").Logger,
	}
}

// Start starts the batch notification sender
func (bns *BatchNotificationSender) Start() {
	go bns.processNotifications()
}

// Stop stops the batch notification sender
func (bns *BatchNotificationSender) Stop() {
	close(bns.done)
}

// QueueNotification queues a notification for batch sending
func (bns *BatchNotificationSender) QueueNotification(notification *A1Notification) {
	select {
	case bns.notifications <- notification:
	default:
		bns.logger.Warn("Notification queue is full, dropping notification")
	}
}

// processNotifications processes notifications in batches
func (bns *BatchNotificationSender) processNotifications() {
	ticker := time.NewTicker(bns.batchTimeout)
	defer ticker.Stop()

	var batch []*A1Notification

	for {
		select {
		case notification := <-bns.notifications:
			batch = append(batch, notification)
			
			if len(batch) >= bns.batchSize {
				bns.sendBatch(batch)
				batch = nil
			}

		case <-ticker.C:
			if len(batch) > 0 {
				bns.sendBatch(batch)
				batch = nil
			}

		case <-bns.done:
			// Send remaining notifications before stopping
			if len(batch) > 0 {
				bns.sendBatch(batch)
			}
			return
		}
	}
}

// sendBatch sends a batch of notifications
func (bns *BatchNotificationSender) sendBatch(batch []*A1Notification) {
	bns.logger.WithField("batch_size", len(batch)).Debug("Sending notification batch")

	// Send notifications concurrently
	var wg sync.WaitGroup
	for _, notification := range batch {
		wg.Add(1)
		go func(n *A1Notification) {
			defer wg.Done()
			if err := bns.service.SendNotification(n); err != nil {
				bns.logger.WithFields(logrus.Fields{
					"notification_id": n.NotificationID,
					"error":           err,
				}).Error("Failed to send notification in batch")
			}
		}(notification)
	}

	wg.Wait()
	bns.logger.WithField("batch_size", len(batch)).Debug("Notification batch sent")
}

// NotificationMetrics tracks metrics for notification sending
type NotificationMetrics struct {
	TotalSent           int64     `json:"total_sent"`
	TotalFailed         int64     `json:"total_failed"`
	AverageLatency      float64   `json:"average_latency_ms"`
	LastNotificationTime time.Time `json:"last_notification_time"`
	mutex               sync.RWMutex
}

// NewNotificationMetrics creates a new notification metrics tracker
func NewNotificationMetrics() *NotificationMetrics {
	return &NotificationMetrics{}
}

// RecordSuccess records a successful notification
func (nm *NotificationMetrics) RecordSuccess(latency time.Duration) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	nm.TotalSent++
	nm.LastNotificationTime = time.Now()
	
	// Update average latency (simple moving average)
	if nm.TotalSent == 1 {
		nm.AverageLatency = float64(latency.Milliseconds())
	} else {
		nm.AverageLatency = (nm.AverageLatency*float64(nm.TotalSent-1) + float64(latency.Milliseconds())) / float64(nm.TotalSent)
	}
}

// RecordFailure records a failed notification
func (nm *NotificationMetrics) RecordFailure() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	nm.TotalFailed++
}

// GetMetrics returns current metrics
func (nm *NotificationMetrics) GetMetrics() map[string]interface{} {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()

	successRate := float64(0)
	if nm.TotalSent+nm.TotalFailed > 0 {
		successRate = float64(nm.TotalSent) / float64(nm.TotalSent+nm.TotalFailed) * 100
	}

	return map[string]interface{}{
		"total_sent":             nm.TotalSent,
		"total_failed":           nm.TotalFailed,
		"success_rate_percent":   successRate,
		"average_latency_ms":     nm.AverageLatency,
		"last_notification_time": nm.LastNotificationTime,
	}
}