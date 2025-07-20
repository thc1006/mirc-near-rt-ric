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

package main

import (
	"context"
	"log/slog"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/go-redis/redis/v8"
	"github.com/hctsai1006/near-rt-ric/dashboard-master/src/app/backend/federatedlearning"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("Failed to get in-cluster config", "error", err)
		os.Exit(1)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error("Failed to create Kubernetes client", "error", err)
		os.Exit(1)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	configMapName := os.Getenv("CONFIG_MAP_NAME")
	if configMapName == "" {
		configMapName = "fl-coordinator-config"
	}

	api, err := federatedlearning.NewAPI(logger, kubeClient, namespace, configMapName, redisClient)
	if err != nil {
		logger.Error("Failed to create API", "error", err)
		os.Exit(1)
	}

	if err := api.Start(context.Background()); err != nil {
		logger.Error("Failed to start API", "error", err)
		os.Exit(1)
	}
}