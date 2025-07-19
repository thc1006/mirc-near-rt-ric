// Copyright 2017 The Kubernetes Authors.
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

package sync

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	"github.com/kubernetes/dashboard/src/app/backend/errors"
	syncApi "github.com/kubernetes/dashboard/src/app/backend/sync/api"
)

// ModernSecretSynchronizer implements a modern, context-aware secret synchronizer with improved concurrency patterns
type ModernSecretSynchronizer struct {
	namespace string
	name      string

	// Protected by mutex
	secret *v1.Secret
	mu     sync.RWMutex

	// Concurrency and lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	client kubernetes.Interface
	logger *slog.Logger

	// Action handlers with proper synchronization
	actionHandlers map[watch.EventType][]syncApi.ActionHandlerFunction
	handlersMu     sync.RWMutex

	// Error reporting
	errChan chan error
	
	// Polling management
	poller   syncApi.Poller
	pollerMu sync.RWMutex

	// Lifecycle management
	started bool
	done    chan struct{}
	wg      sync.WaitGroup
}

// NewModernSecretSynchronizer creates a new modern secret synchronizer with context support
func NewModernSecretSynchronizer(ctx context.Context, namespace, name string, client kubernetes.Interface, logger *slog.Logger) *ModernSecretSynchronizer {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = slog.Default()
	}

	childCtx, cancel := context.WithCancel(ctx)

	return &ModernSecretSynchronizer{
		namespace:      namespace,
		name:           name,
		ctx:            childCtx,
		cancel:         cancel,
		client:         client,
		logger:         logger,
		actionHandlers: make(map[watch.EventType][]syncApi.ActionHandlerFunction),
		errChan:        make(chan error, 1), // Buffered to prevent blocking
		done:           make(chan struct{}),
	}
}

// Name returns the synchronizer name with namespace prefix
func (s *ModernSecretSynchronizer) Name() string {
	return fmt.Sprintf("secret-%s-%s", s.name, s.namespace)
}

// Start begins the synchronization process with proper lifecycle management
func (s *ModernSecretSynchronizer) Start() {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		s.logger.Warn("Secret synchronizer already started", "name", s.Name())
		return
	}
	s.started = true
	s.mu.Unlock()

	s.logger.Info("Starting secret synchronizer", "name", s.Name())

	s.wg.Add(1)
	go s.watchLoop()
}

// Stop gracefully stops the synchronizer
func (s *ModernSecretSynchronizer) Stop() {
	s.logger.Info("Stopping secret synchronizer", "name", s.Name())
	s.cancel()
	
	// Wait for goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Debug("Secret synchronizer stopped gracefully", "name", s.Name())
	case <-time.After(5 * time.Second):
		s.logger.Warn("Timeout waiting for secret synchronizer to stop", "name", s.Name())
	}

	close(s.errChan)
	close(s.done)
}

// watchLoop is the main watch loop with proper error handling and context cancellation
func (s *ModernSecretSynchronizer) watchLoop() {
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Secret synchronizer watch loop panicked",
				"name", s.Name(),
				"panic", r)
			select {
			case s.errChan <- fmt.Errorf("watch loop panicked: %v", r):
			case <-s.ctx.Done():
			}
		}
	}()

	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Debug("Secret synchronizer context cancelled", "name", s.Name())
			return
		default:
		}

		watcher, err := s.createWatcher()
		if err != nil {
			s.logger.Error("Failed to create watcher",
				"name", s.Name(),
				"error", err,
				"retry_in", backoff)
			
			select {
			case s.errChan <- err:
			case <-s.ctx.Done():
				return
			}

			// Exponential backoff with jitter
			select {
			case <-time.After(backoff):
				backoff = min(backoff*2, maxBackoff)
			case <-s.ctx.Done():
				return
			}
			continue
		}

		// Reset backoff on successful connection
		backoff = time.Second

		// Watch for events
		s.processEvents(watcher)
		watcher.Stop()
	}
}

// createWatcher creates a new watcher with proper error handling
func (s *ModernSecretSynchronizer) createWatcher() (watch.Interface, error) {
	s.pollerMu.RLock()
	poller := s.poller
	s.pollerMu.RUnlock()

	if poller == nil {
		// Use default polling mechanism
		return s.watchDirect()
	}

	return poller.Poll(5 * time.Minute), nil
}

// watchDirect creates a direct Kubernetes watch with context
func (s *ModernSecretSynchronizer) watchDirect() (watch.Interface, error) {
	watchCtx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	listOptions := metaV1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", s.name),
		Watch:         true,
	}

	return s.client.CoreV1().Secrets(s.namespace).Watch(watchCtx, listOptions)
}

// processEvents processes watch events with proper error handling
func (s *ModernSecretSynchronizer) processEvents(watcher watch.Interface) {
	defer watcher.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case event, ok := <-watcher.ResultChan():
			if !ok {
				s.logger.Debug("Watch channel closed", "name", s.Name())
				return
			}

			if err := s.handleEvent(event); err != nil {
				s.logger.Error("Error handling watch event",
					"name", s.Name(),
					"event_type", event.Type,
					"error", err)
				
				select {
				case s.errChan <- err:
				case <-s.ctx.Done():
					return
				}
				return
			}
		}
	}
}

// handleEvent processes individual watch events with proper concurrency control
func (s *ModernSecretSynchronizer) handleEvent(event watch.Event) error {
	s.logger.Debug("Handling watch event",
		"name", s.Name(),
		"event_type", event.Type)

	// Call registered action handlers
	s.handlersMu.RLock()
	handlers := make([]syncApi.ActionHandlerFunction, len(s.actionHandlers[event.Type]))
	copy(handlers, s.actionHandlers[event.Type])
	s.handlersMu.RUnlock()

	for _, handler := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("Action handler panicked",
						"name", s.Name(),
						"event_type", event.Type,
						"panic", r)
				}
			}()
			handler(event.Object)
		}()
	}

	// Process the event
	switch event.Type {
	case watch.Added, watch.Modified:
		secret, ok := event.Object.(*v1.Secret)
		if !ok {
			return errors.NewInternal(fmt.Sprintf("Expected secret got %s", reflect.TypeOf(event.Object)))
		}
		s.updateSecret(*secret)

	case watch.Deleted:
		s.clearSecret()

	case watch.Error:
		return errors.NewUnexpectedObject(event.Object)

	default:
		s.logger.Warn("Unknown watch event type",
			"name", s.Name(),
			"event_type", event.Type)
	}

	return nil
}

// updateSecret safely updates the stored secret
func (s *ModernSecretSynchronizer) updateSecret(secret v1.Secret) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.secret != nil && reflect.DeepEqual(s.secret, &secret) {
		// Skip update if existing object is the same as new one
		return
	}

	s.secret = &secret
	s.logger.Debug("Updated secret",
		"name", s.Name(),
		"resource_version", secret.ResourceVersion)
}

// clearSecret safely clears the stored secret
func (s *ModernSecretSynchronizer) clearSecret() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secret = nil
	s.logger.Debug("Cleared secret", "name", s.Name())
}

// Error returns the error channel for monitoring
func (s *ModernSecretSynchronizer) Error() chan error {
	return s.errChan
}

// Get safely retrieves the current secret with fallback to API call
func (s *ModernSecretSynchronizer) Get() runtime.Object {
	s.mu.RLock()
	secret := s.secret
	s.mu.RUnlock()

	if secret != nil {
		return secret
	}

	// Fallback to synchronous API call with context
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	secret, err := s.client.CoreV1().Secrets(s.namespace).Get(ctx, s.name, metaV1.GetOptions{})
	if err != nil {
		s.logger.Error("Failed to get secret synchronously",
			"name", s.Name(),
			"error", err)
		return nil
	}

	s.logger.Debug("Retrieved secret synchronously", "name", s.Name())
	
	// Update cached value
	s.mu.Lock()
	s.secret = secret
	s.mu.Unlock()

	return secret
}

// Create creates a new secret with context support
func (s *ModernSecretSynchronizer) Create(obj runtime.Object) error {
	secret := s.getSecret(obj)
	
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	_, err := s.client.CoreV1().Secrets(secret.Namespace).Create(ctx, secret, metaV1.CreateOptions{})
	if err != nil {
		s.logger.Error("Failed to create secret",
			"name", s.Name(),
			"error", err)
		return err
	}

	s.logger.Info("Created secret", "name", s.Name())
	return nil
}

// Update updates an existing secret with context support
func (s *ModernSecretSynchronizer) Update(obj runtime.Object) error {
	secret := s.getSecret(obj)
	
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	_, err := s.client.CoreV1().Secrets(secret.Namespace).Update(ctx, secret, metaV1.UpdateOptions{})
	if err != nil {
		s.logger.Error("Failed to update secret",
			"name", s.Name(),
			"error", err)
		return err
	}

	s.logger.Info("Updated secret", "name", s.Name())
	return nil
}

// Delete deletes the secret with context support
func (s *ModernSecretSynchronizer) Delete() error {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	err := s.client.CoreV1().Secrets(s.namespace).Delete(ctx, s.name,
		metaV1.DeleteOptions{GracePeriodSeconds: new(int64)})
	if err != nil {
		s.logger.Error("Failed to delete secret",
			"name", s.Name(),
			"error", err)
		return err
	}

	s.logger.Info("Deleted secret", "name", s.Name())
	return nil
}

// RegisterActionHandler registers an action handler with proper concurrency control
func (s *ModernSecretSynchronizer) RegisterActionHandler(handler syncApi.ActionHandlerFunction, events ...watch.EventType) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()

	for _, ev := range events {
		if _, exists := s.actionHandlers[ev]; !exists {
			s.actionHandlers[ev] = make([]syncApi.ActionHandlerFunction, 0)
		}
		s.actionHandlers[ev] = append(s.actionHandlers[ev], handler)
	}

	s.logger.Debug("Registered action handler",
		"name", s.Name(),
		"events", events)
}

// Refresh manually refreshes the secret from the API
func (s *ModernSecretSynchronizer) Refresh() {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	secret, err := s.client.CoreV1().Secrets(s.namespace).Get(ctx, s.name, metaV1.GetOptions{})
	if err != nil {
		s.logger.Error("Failed to refresh secret",
			"name", s.Name(),
			"error", err)
		return
	}

	s.mu.Lock()
	s.secret = secret
	s.mu.Unlock()

	s.logger.Debug("Refreshed secret", "name", s.Name())
}

// SetPoller sets the custom poller with proper synchronization
func (s *ModernSecretSynchronizer) SetPoller(poller syncApi.Poller) {
	s.pollerMu.Lock()
	defer s.pollerMu.Unlock()
	s.poller = poller
}

// getSecret safely converts runtime.Object to *v1.Secret
func (s *ModernSecretSynchronizer) getSecret(obj runtime.Object) *v1.Secret {
	secret, ok := obj.(*v1.Secret)
	if !ok {
		panic("Provided object has to be a secret. Most likely this is a programming error")
	}
	return secret
}

// Context returns the synchronizer's context for cancellation propagation
func (s *ModernSecretSynchronizer) Context() context.Context {
	return s.ctx
}

// min returns the minimum of two time.Duration values
func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}