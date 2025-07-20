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
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

const (
	clientKeyPrefix = "fl:client:"
	modelKeyPrefix  = "fl:model:"
)

// RedisClientStore implements the ClientStore interface using Redis.
type RedisClientStore struct {
	client *redis.Client
}

// NewRedisClientStore creates a new Redis-based client store.
func NewRedisClientStore(client *redis.Client) *RedisClientStore {
	return &RedisClientStore{client: client}
}

// Store stores a client in Redis.
func (s *RedisClientStore) Store(ctx context.Context, client *FLClient) error {
	key := clientKeyPrefix + client.ID
	data, err := json.Marshal(client)
	if err != nil {
		return fmt.Errorf("failed to marshal client: %w", err)
	}
	return s.client.Set(ctx, key, data, 0).Err()
}

// Get retrieves a client from Redis.
func (s *RedisClientStore) Get(ctx context.Context, clientID string) (*FLClient, error) {
	key := clientKeyPrefix + clientID
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	var client FLClient
	if err := json.Unmarshal(data, &client); err != nil {
		return nil, fmt.Errorf("failed to unmarshal client: %w", err)
	}
	return &client, nil
}

// List retrieves all clients from Redis.
func (s *RedisClientStore) List(ctx context.Context) ([]*FLClient, error) {
	keys, err := s.client.Keys(ctx, clientKeyPrefix+"*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list clients: %w", err)
	}

	var clients []*FLClient
	for _, key := range keys {
		data, err := s.client.Get(ctx, key).Bytes()
		if err != nil {
			return nil, fmt.Errorf("failed to get client: %w", err)
		}
		var client FLClient
		if err := json.Unmarshal(data, &client); err != nil {
			return nil, fmt.Errorf("failed to unmarshal client: %w", err)
		}
		clients = append(clients, &client)
	}
	return clients, nil
}

// Delete removes a client from Redis.
func (s *RedisClientStore) Delete(ctx context.Context, clientID string) error {
	key := clientKeyPrefix + clientID
	return s.client.Del(ctx, key).Err()
}

// RedisModelStore implements the ModelStore interface using Redis.
type RedisModelStore struct {
	client *redis.Client
}

// NewRedisModelStore creates a new Redis-based model store.
func NewRedisModelStore(client *redis.Client) *RedisModelStore {
	return &RedisModelStore{client: client}
}

// CreateGlobalModel creates a new global model in Redis.
func (s *RedisModelStore) CreateGlobalModel(ctx context.Context, model *GlobalModel) error {
	key := modelKeyPrefix + model.ID
	data, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("failed to marshal model: %w", err)
	}
	return s.client.Set(ctx, key, data, 0).Err()
}

// GetGlobalModel retrieves a global model from Redis.
func (s *RedisModelStore) GetGlobalModel(ctx context.Context, modelID string) (*GlobalModel, error) {
	key := modelKeyPrefix + modelID
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}
	var model GlobalModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to unmarshal model: %w", err)
	}
	return &model, nil
}

// UpdateGlobalModel updates a global model in Redis.
func (s *RedisModelStore) UpdateGlobalModel(ctx context.Context, model *GlobalModel) error {
	return s.CreateGlobalModel(ctx, model) // Same as create
}
