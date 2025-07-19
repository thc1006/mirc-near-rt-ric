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

package common

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubernetes/dashboard/src/app/backend/api"
)

// ModernResourceChannel provides a context-aware, cancellable resource channel with proper error handling
type ModernResourceChannel[T any] struct {
	Data   chan T
	Error  chan error
	Cancel context.CancelFunc
	ctx    context.Context
	logger *slog.Logger
}

// NewModernResourceChannel creates a new modern resource channel with context and cancellation support
func NewModernResourceChannel[T any](ctx context.Context, bufferSize int, logger *slog.Logger) *ModernResourceChannel[T] {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = slog.Default()
	}

	childCtx, cancel := context.WithCancel(ctx)
	
	return &ModernResourceChannel[T]{
		Data:   make(chan T, bufferSize),
		Error:  make(chan error, 1), // Error channel should be buffered to prevent blocking
		Cancel: cancel,
		ctx:    childCtx,
		logger: logger,
	}
}

// Close safely closes the channels and cancels the context
func (mrc *ModernResourceChannel[T]) Close() {
	mrc.Cancel()
	close(mrc.Data)
	close(mrc.Error)
}

// Context returns the channel's context for cancellation propagation
func (mrc *ModernResourceChannel[T]) Context() context.Context {
	return mrc.ctx
}

// ModernPodListChannel is a modernized pod list channel with context support
type ModernPodListChannel = ModernResourceChannel[*v1.PodList]

// GetModernPodListChannel returns a context-aware pod list channel with proper concurrency management
func GetModernPodListChannel(ctx context.Context, client kubernetes.Interface, nsQuery *NamespaceQuery, logger *slog.Logger) *ModernPodListChannel {
	channel := NewModernResourceChannel[*v1.PodList](ctx, 1, logger)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				channel.logger.Error("Pod list channel goroutine panicked",
					"panic", r)
				select {
				case channel.Error <- fmt.Errorf("pod list channel panicked: %v", r):
				case <-channel.ctx.Done():
				}
			}
		}()

		// Use context with timeout for the Kubernetes API call
		listCtx, cancel := context.WithTimeout(channel.ctx, 30*time.Second)
		defer cancel()

		list, err := client.CoreV1().Pods(nsQuery.ToRequestParam()).List(listCtx, api.ListEverything)
		if err != nil {
			channel.logger.Error("Failed to list pods",
				"namespace", nsQuery.ToRequestParam(),
				"error", err)
			select {
			case channel.Error <- err:
			case <-channel.ctx.Done():
				channel.logger.Debug("Context cancelled during error send")
			}
			return
		}

		// Filter items based on namespace query
		var filteredItems []v1.Pod
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems

		// Send result with context cancellation support
		select {
		case channel.Data <- list:
			channel.logger.Debug("Successfully sent pod list",
				"pod_count", len(list.Items))
		case <-channel.ctx.Done():
			channel.logger.Debug("Context cancelled during data send")
			return
		}

		// Always send nil error on success
		select {
		case channel.Error <- nil:
		case <-channel.ctx.Done():
		}
	}()

	return channel
}

// ModernServiceListChannel is a modernized service list channel
type ModernServiceListChannel = ModernResourceChannel[*v1.ServiceList]

// GetModernServiceListChannel returns a context-aware service list channel
func GetModernServiceListChannel(ctx context.Context, client kubernetes.Interface, nsQuery *NamespaceQuery, logger *slog.Logger) *ModernServiceListChannel {
	channel := NewModernResourceChannel[*v1.ServiceList](ctx, 1, logger)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				channel.logger.Error("Service list channel goroutine panicked",
					"panic", r)
				select {
				case channel.Error <- fmt.Errorf("service list channel panicked: %v", r):
				case <-channel.ctx.Done():
				}
			}
		}()

		listCtx, cancel := context.WithTimeout(channel.ctx, 30*time.Second)
		defer cancel()

		list, err := client.CoreV1().Services(nsQuery.ToRequestParam()).List(listCtx, api.ListEverything)
		if err != nil {
			channel.logger.Error("Failed to list services",
				"namespace", nsQuery.ToRequestParam(),
				"error", err)
			select {
			case channel.Error <- err:
			case <-channel.ctx.Done():
			}
			return
		}

		var filteredItems []v1.Service
		for _, item := range list.Items {
			if nsQuery.Matches(item.ObjectMeta.Namespace) {
				filteredItems = append(filteredItems, item)
			}
		}
		list.Items = filteredItems

		select {
		case channel.Data <- list:
			channel.logger.Debug("Successfully sent service list",
				"service_count", len(list.Items))
		case <-channel.ctx.Done():
			return
		}

		select {
		case channel.Error <- nil:
		case <-channel.ctx.Done():
		}
	}()

	return channel
}

// ModernResourceChannelManager manages multiple resource channels with proper lifecycle management
type ModernResourceChannelManager struct {
	channels []ChannelCloser
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	logger   *slog.Logger
	mu       sync.RWMutex
}

// ChannelCloser interface for channels that can be closed
type ChannelCloser interface {
	Close()
}

// NewModernResourceChannelManager creates a new channel manager
func NewModernResourceChannelManager(ctx context.Context, logger *slog.Logger) *ModernResourceChannelManager {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = slog.Default()
	}

	childCtx, cancel := context.WithCancel(ctx)
	
	return &ModernResourceChannelManager{
		channels: make([]ChannelCloser, 0),
		ctx:      childCtx,
		cancel:   cancel,
		logger:   logger,
	}
}

// AddChannel adds a channel to be managed
func (m *ModernResourceChannelManager) AddChannel(channel ChannelCloser) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels = append(m.channels, channel)
}

// Close safely closes all managed channels
func (m *ModernResourceChannelManager) Close() {
	m.cancel() // Cancel context first to signal shutdown

	m.mu.Lock()
	channels := make([]ChannelCloser, len(m.channels))
	copy(channels, m.channels)
	m.mu.Unlock()

	// Close all channels
	for _, channel := range channels {
		if channel != nil {
			channel.Close()
		}
	}

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.logger.Debug("All channels closed successfully")
	case <-time.After(5 * time.Second):
		m.logger.Warn("Timeout waiting for channels to close")
	}
}

// Context returns the manager's context
func (m *ModernResourceChannelManager) Context() context.Context {
	return m.ctx
}

// Wait waits for all managed goroutines to complete
func (m *ModernResourceChannelManager) Wait() {
	m.wg.Wait()
}