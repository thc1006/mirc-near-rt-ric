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
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClientStore is a Redis-based implementation of the ClientStore interface.
type RedisClientStore struct {
	client *redis.Client
}

// NewRedisClientStore creates a new RedisClientStore.
func NewRedisClientStore(client *redis.Client) *RedisClientStore {
	return &RedisClientStore{client: client}
}

// Store stores a client in Redis.
func (s *RedisClientStore) Store(ctx context.Context, client *FLClient) error {
	data, err := json.Marshal(client)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, "client:"+client.ID, data, 0).Err()
}

// Get retrieves a client from Redis.
func (s *RedisClientStore) Get(ctx context.Context, clientID string) (*FLClient, error) {
	data, err := s.client.Get(ctx, "client:"+clientID).Bytes()
	if err != nil {
		return nil, err
	}
	var client FLClient
	if err := json.Unmarshal(data, &client); err != nil {
		return nil, err
	}
	return &client, nil
}

// Delete deletes a client from Redis.
func (s *RedisClientStore) Delete(ctx context.Context, clientID string) error {
	return s.client.Del(ctx, "client:"+clientID).Err()
}

// List lists all clients in Redis.
func (s *RedisClientStore) List(ctx context.Context) ([]*FLClient, error) {
	keys, err := s.client.Keys(ctx, "client:*").Result()
	if err != nil {
		return nil, err
	}
	var clients []*FLClient
	for _, key := range keys {
		data, err := s.client.Get(ctx, key).Bytes()
		if err != nil {
			return nil, err
		}
		var client FLClient
		if err := json.Unmarshal(data, &client); err != nil {
			return nil, err
		}
		clients = append(clients, &client)
	}
	return clients, nil
}

// RedisModelStore is a Redis-based implementation of the ModelStore interface.
type RedisModelStore struct {
	client *redis.Client
}

// NewRedisModelStore creates a new RedisModelStore.
func NewRedisModelStore(client *redis.Client) *RedisModelStore {
	return &RedisModelStore{client: client}
}

// CreateGlobalModel creates a new global model in Redis.
func (s *RedisModelStore) CreateGlobalModel(ctx context.Context, model *GlobalModel) error {
	data, err := json.Marshal(model)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, "model:"+model.ID, data, 0).Err()
}

// GetGlobalModel retrieves a global model from Redis.
func (s *RedisModelStore) GetGlobalModel(ctx context.Context, modelID string) (*GlobalModel, error) {
	data, err := s.client.Get(ctx, "model:"+modelID).Bytes()
	if err != nil {
		return nil, err
	}
	var model GlobalModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, err
	}
	return &model, nil
}

// UpdateGlobalModel updates a global model in Redis.
func (s *RedisModelStore) UpdateGlobalModel(ctx context.Context, model *GlobalModel) error {
	return s.CreateGlobalModel(ctx, model)
}
