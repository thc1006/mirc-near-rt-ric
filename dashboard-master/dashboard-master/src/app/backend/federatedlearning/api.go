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
	"log/slog"

	"k8s.io/client-go/kubernetes"

	"github.com/go-redis/redis/v8"
)

// API provides the entry points for the federated learning coordinator.
type API struct {
	coordinator *FLCoordinator
	logger      *slog.Logger
}

// NewAPI creates a new API.
func NewAPI(logger *slog.Logger, kubeClient kubernetes.Interface, namespace, configMapName string, redisClient *redis.Client) (*API, error) {
	coordinator, err := NewFLCoordinator(logger, kubeClient, namespace, configMapName, redisClient)
	if err != nil {
		return nil, err
	}

	return &API{
		coordinator: coordinator,
		logger:      logger,
	}, nil
}

// Start starts the federated learning coordinator.
func (a *API) Start(ctx context.Context) error {
	config := a.coordinator.configManager.GetConfig()
	service := NewService(a.coordinator, a.logger)
	return service.Start(ctx, config)
}

// Coordinator returns the underlying FLCoordinator.
func (a *API) Coordinator() *FLCoordinator {
	return a.coordinator
}
