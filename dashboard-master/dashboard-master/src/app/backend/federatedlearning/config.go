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
	"fmt"
	"log/slog"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

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
