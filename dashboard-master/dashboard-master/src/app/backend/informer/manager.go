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

package informer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
)

// InformerManager manages Kubernetes informers for efficient resource watching
type InformerManager struct {
	client          kubernetes.Interface
	informerFactory informers.SharedInformerFactory
	logger          *slog.Logger
	
	// Informers for different resource types
	podInformer        cache.SharedIndexInformer
	deploymentInformer cache.SharedIndexInformer
	serviceInformer    cache.SharedIndexInformer
	replicaSetInformer cache.SharedIndexInformer
	nodeInformer       cache.SharedIndexInformer
	eventInformer      cache.SharedIndexInformer
	
	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	
	// State tracking
	started bool
	synced  bool
	mu      sync.RWMutex
	
	// Metrics
	eventCount map[string]int64
	eventMu    sync.RWMutex
}

// NewInformerManager creates a new informer manager
func NewInformerManager(ctx context.Context, client kubernetes.Interface, logger *slog.Logger) *InformerManager {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = slog.Default()
	}
	
	childCtx, cancel := context.WithCancel(ctx)
	
	im := &InformerManager{
		client:     client,
		logger:     logger,
		ctx:        childCtx,
		cancel:     cancel,
		eventCount: make(map[string]int64),
	}
	
	// Create informer factory with 30 second resync period
	im.informerFactory = informers.NewSharedInformerFactory(client, 30*time.Second)
	
	// Setup informers
	im.setupInformers()
	
	return im
}

// setupInformers configures all the informers
func (im *InformerManager) setupInformers() {
	// Pod informer
	im.podInformer = im.informerFactory.Core().V1().Pods().Informer()
	im.podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    im.onPodAdd,
		UpdateFunc: im.onPodUpdate,
		DeleteFunc: im.onPodDelete,
	})
	
	// Deployment informer
	im.deploymentInformer = im.informerFactory.Apps().V1().Deployments().Informer()
	im.deploymentInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    im.onDeploymentAdd,
		UpdateFunc: im.onDeploymentUpdate,
		DeleteFunc: im.onDeploymentDelete,
	})
	
	// Service informer
	im.serviceInformer = im.informerFactory.Core().V1().Services().Informer()
	im.serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    im.onServiceAdd,
		UpdateFunc: im.onServiceUpdate,
		DeleteFunc: im.onServiceDelete,
	})
	
	// ReplicaSet informer
	im.replicaSetInformer = im.informerFactory.Apps().V1().ReplicaSets().Informer()
	im.replicaSetInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    im.onReplicaSetAdd,
		UpdateFunc: im.onReplicaSetUpdate,
		DeleteFunc: im.onReplicaSetDelete,
	})
	
	// Node informer
	im.nodeInformer = im.informerFactory.Core().V1().Nodes().Informer()
	im.nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    im.onNodeAdd,
		UpdateFunc: im.onNodeUpdate,
		DeleteFunc: im.onNodeDelete,
	})
	
	// Event informer
	im.eventInformer = im.informerFactory.Core().V1().Events().Informer()
	im.eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    im.onEventAdd,
		UpdateFunc: im.onEventUpdate,
		DeleteFunc: im.onEventDelete,
	})
}

// Start begins the informer operations
func (im *InformerManager) Start() error {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	if im.started {
		return fmt.Errorf("informer manager already started")
	}
	
	im.logger.Info("Starting informer manager")
	
	// Start all informers
	im.informerFactory.Start(im.ctx.Done())
	
	// Wait for cache sync with timeout
	syncTimeout := 30 * time.Second
	syncCtx, cancel := context.WithTimeout(im.ctx, syncTimeout)
	defer cancel()
	
	informers := map[string]cache.SharedIndexInformer{
		"pods":        im.podInformer,
		"deployments": im.deploymentInformer,
		"services":    im.serviceInformer,
		"replicasets": im.replicaSetInformer,
		"nodes":       im.nodeInformer,
		"events":      im.eventInformer,
	}
	
	for name, informer := range informers {
		im.logger.Debug("Waiting for cache sync", "informer", name)
		if !cache.WaitForCacheSync(syncCtx.Done(), informer.HasSynced) {
			return fmt.Errorf("failed to sync cache for %s informer", name)
		}
		im.logger.Debug("Cache synced", "informer", name)
	}
	
	im.started = true
	im.synced = true
	im.logger.Info("Informer manager started successfully")
	
	return nil
}

// Stop gracefully stops the informer manager
func (im *InformerManager) Stop() {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	if !im.started {
		return
	}
	
	im.logger.Info("Stopping informer manager")
	im.cancel()
	im.started = false
	im.synced = false
}

// IsReady returns true if all informers are synced and ready
func (im *InformerManager) IsReady() bool {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.started && im.synced
}

// GetPods returns pods from the local cache
func (im *InformerManager) GetPods(namespace string, selector labels.Selector) ([]*corev1.Pod, error) {
	if !im.IsReady() {
		return nil, fmt.Errorf("informer not ready")
	}
	
	lister := im.informerFactory.Core().V1().Pods().Lister()
	
	if namespace == "" {
		return lister.List(selector)
	}
	
	return lister.Pods(namespace).List(selector)
}

// GetPodsFromCache efficiently retrieves pods using informer cache
func (im *InformerManager) GetPodsFromCache(nsQuery *common.NamespaceQuery) ([]*corev1.Pod, error) {
	if !im.IsReady() {
		return nil, fmt.Errorf("informer not ready")
	}
	
	var pods []*corev1.Pod
	var err error
	
	if nsQuery.ToRequestParam() == "" {
		// All namespaces
		pods, err = im.informerFactory.Core().V1().Pods().Lister().List(labels.Everything())
	} else {
		// Specific namespace
		pods, err = im.informerFactory.Core().V1().Pods().Lister().Pods(nsQuery.ToRequestParam()).List(labels.Everything())
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to list pods from cache: %w", err)
	}
	
	// Filter by namespace query if needed
	if nsQuery != nil {
		filteredPods := make([]*corev1.Pod, 0, len(pods))
		for _, pod := range pods {
			if nsQuery.Matches(pod.Namespace) {
				filteredPods = append(filteredPods, pod)
			}
		}
		pods = filteredPods
	}
	
	return pods, nil
}

// GetDeploymentsFromCache efficiently retrieves deployments using informer cache
func (im *InformerManager) GetDeploymentsFromCache(nsQuery *common.NamespaceQuery) ([]*v1.Deployment, error) {
	if !im.IsReady() {
		return nil, fmt.Errorf("informer not ready")
	}
	
	var deployments []*v1.Deployment
	var err error
	
	if nsQuery.ToRequestParam() == "" {
		deployments, err = im.informerFactory.Apps().V1().Deployments().Lister().List(labels.Everything())
	} else {
		deployments, err = im.informerFactory.Apps().V1().Deployments().Lister().Deployments(nsQuery.ToRequestParam()).List(labels.Everything())
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments from cache: %w", err)
	}
	
	// Filter by namespace query
	if nsQuery != nil {
		filteredDeployments := make([]*v1.Deployment, 0, len(deployments))
		for _, deployment := range deployments {
			if nsQuery.Matches(deployment.Namespace) {
				filteredDeployments = append(filteredDeployments, deployment)
			}
		}
		deployments = filteredDeployments
	}
	
	return deployments, nil
}

// GetServicesFromCache efficiently retrieves services using informer cache
func (im *InformerManager) GetServicesFromCache(nsQuery *common.NamespaceQuery) ([]*corev1.Service, error) {
	if !im.IsReady() {
		return nil, fmt.Errorf("informer not ready")
	}
	
	var services []*corev1.Service
	var err error
	
	if nsQuery.ToRequestParam() == "" {
		services, err = im.informerFactory.Core().V1().Services().Lister().List(labels.Everything())
	} else {
		services, err = im.informerFactory.Core().V1().Services().Lister().Services(nsQuery.ToRequestParam()).List(labels.Everything())
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to list services from cache: %w", err)
	}
	
	// Filter by namespace query
	if nsQuery != nil {
		filteredServices := make([]*corev1.Service, 0, len(services))
		for _, service := range services {
			if nsQuery.Matches(service.Namespace) {
				filteredServices = append(filteredServices, service)
			}
		}
		services = filteredServices
	}
	
	return services, nil
}

// GetReplicaSetsFromCache efficiently retrieves replica sets using informer cache
func (im *InformerManager) GetReplicaSetsFromCache(nsQuery *common.NamespaceQuery) ([]*v1.ReplicaSet, error) {
	if !im.IsReady() {
		return nil, fmt.Errorf("informer not ready")
	}
	
	var replicaSets []*v1.ReplicaSet
	var err error
	
	if nsQuery.ToRequestParam() == "" {
		replicaSets, err = im.informerFactory.Apps().V1().ReplicaSets().Lister().List(labels.Everything())
	} else {
		replicaSets, err = im.informerFactory.Apps().V1().ReplicaSets().Lister().ReplicaSets(nsQuery.ToRequestParam()).List(labels.Everything())
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to list replica sets from cache: %w", err)
	}
	
	// Filter by namespace query
	if nsQuery != nil {
		filteredReplicaSets := make([]*v1.ReplicaSet, 0, len(replicaSets))
		for _, rs := range replicaSets {
			if nsQuery.Matches(rs.Namespace) {
				filteredReplicaSets = append(filteredReplicaSets, rs)
			}
		}
		replicaSets = filteredReplicaSets
	}
	
	return replicaSets, nil
}

// GetEventsFromCache efficiently retrieves events using informer cache
func (im *InformerManager) GetEventsFromCache(nsQuery *common.NamespaceQuery) ([]*corev1.Event, error) {
	if !im.IsReady() {
		return nil, fmt.Errorf("informer not ready")
	}
	
	var events []*corev1.Event
	var err error
	
	if nsQuery.ToRequestParam() == "" {
		events, err = im.informerFactory.Core().V1().Events().Lister().List(labels.Everything())
	} else {
		events, err = im.informerFactory.Core().V1().Events().Lister().Events(nsQuery.ToRequestParam()).List(labels.Everything())
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to list events from cache: %w", err)
	}
	
	// Filter by namespace query
	if nsQuery != nil {
		filteredEvents := make([]*corev1.Event, 0, len(events))
		for _, event := range events {
			if nsQuery.Matches(event.Namespace) {
				filteredEvents = append(filteredEvents, event)
			}
		}
		events = filteredEvents
	}
	
	return events, nil
}

// Event handlers
func (im *InformerManager) onPodAdd(obj interface{}) {
	if pod, ok := obj.(*corev1.Pod); ok {
		im.incrementEventCount("pod_add")
		im.logger.Debug("Pod added", "name", pod.Name, "namespace", pod.Namespace)
	}
}

func (im *InformerManager) onPodUpdate(oldObj, newObj interface{}) {
	if pod, ok := newObj.(*corev1.Pod); ok {
		im.incrementEventCount("pod_update")
		im.logger.Debug("Pod updated", "name", pod.Name, "namespace", pod.Namespace)
	}
}

func (im *InformerManager) onPodDelete(obj interface{}) {
	if pod, ok := obj.(*corev1.Pod); ok {
		im.incrementEventCount("pod_delete")
		im.logger.Debug("Pod deleted", "name", pod.Name, "namespace", pod.Namespace)
	}
}

func (im *InformerManager) onDeploymentAdd(obj interface{}) {
	if deployment, ok := obj.(*v1.Deployment); ok {
		im.incrementEventCount("deployment_add")
		im.logger.Debug("Deployment added", "name", deployment.Name, "namespace", deployment.Namespace)
	}
}

func (im *InformerManager) onDeploymentUpdate(oldObj, newObj interface{}) {
	if deployment, ok := newObj.(*v1.Deployment); ok {
		im.incrementEventCount("deployment_update")
		im.logger.Debug("Deployment updated", "name", deployment.Name, "namespace", deployment.Namespace)
	}
}

func (im *InformerManager) onDeploymentDelete(obj interface{}) {
	if deployment, ok := obj.(*v1.Deployment); ok {
		im.incrementEventCount("deployment_delete")
		im.logger.Debug("Deployment deleted", "name", deployment.Name, "namespace", deployment.Namespace)
	}
}

func (im *InformerManager) onServiceAdd(obj interface{}) {
	if service, ok := obj.(*corev1.Service); ok {
		im.incrementEventCount("service_add")
		im.logger.Debug("Service added", "name", service.Name, "namespace", service.Namespace)
	}
}

func (im *InformerManager) onServiceUpdate(oldObj, newObj interface{}) {
	if service, ok := newObj.(*corev1.Service); ok {
		im.incrementEventCount("service_update")
		im.logger.Debug("Service updated", "name", service.Name, "namespace", service.Namespace)
	}
}

func (im *InformerManager) onServiceDelete(obj interface{}) {
	if service, ok := obj.(*corev1.Service); ok {
		im.incrementEventCount("service_delete")
		im.logger.Debug("Service deleted", "name", service.Name, "namespace", service.Namespace)
	}
}

func (im *InformerManager) onReplicaSetAdd(obj interface{}) {
	if rs, ok := obj.(*v1.ReplicaSet); ok {
		im.incrementEventCount("replicaset_add")
		im.logger.Debug("ReplicaSet added", "name", rs.Name, "namespace", rs.Namespace)
	}
}

func (im *InformerManager) onReplicaSetUpdate(oldObj, newObj interface{}) {
	if rs, ok := newObj.(*v1.ReplicaSet); ok {
		im.incrementEventCount("replicaset_update")
		im.logger.Debug("ReplicaSet updated", "name", rs.Name, "namespace", rs.Namespace)
	}
}

func (im *InformerManager) onReplicaSetDelete(obj interface{}) {
	if rs, ok := obj.(*v1.ReplicaSet); ok {
		im.incrementEventCount("replicaset_delete")
		im.logger.Debug("ReplicaSet deleted", "name", rs.Name, "namespace", rs.Namespace)
	}
}

func (im *InformerManager) onNodeAdd(obj interface{}) {
	if node, ok := obj.(*corev1.Node); ok {
		im.incrementEventCount("node_add")
		im.logger.Debug("Node added", "name", node.Name)
	}
}

func (im *InformerManager) onNodeUpdate(oldObj, newObj interface{}) {
	if node, ok := newObj.(*corev1.Node); ok {
		im.incrementEventCount("node_update")
		im.logger.Debug("Node updated", "name", node.Name)
	}
}

func (im *InformerManager) onNodeDelete(obj interface{}) {
	if node, ok := obj.(*corev1.Node); ok {
		im.incrementEventCount("node_delete")
		im.logger.Debug("Node deleted", "name", node.Name)
	}
}

func (im *InformerManager) onEventAdd(obj interface{}) {
	if event, ok := obj.(*corev1.Event); ok {
		im.incrementEventCount("event_add")
		im.logger.Debug("Event added", "name", event.Name, "namespace", event.Namespace)
	}
}

func (im *InformerManager) onEventUpdate(oldObj, newObj interface{}) {
	if event, ok := newObj.(*corev1.Event); ok {
		im.incrementEventCount("event_update")
		im.logger.Debug("Event updated", "name", event.Name, "namespace", event.Namespace)
	}
}

func (im *InformerManager) onEventDelete(obj interface{}) {
	if event, ok := obj.(*corev1.Event); ok {
		im.incrementEventCount("event_delete")
		im.logger.Debug("Event deleted", "name", event.Name, "namespace", event.Namespace)
	}
}

// incrementEventCount safely increments the event count
func (im *InformerManager) incrementEventCount(eventType string) {
	im.eventMu.Lock()
	defer im.eventMu.Unlock()
	im.eventCount[eventType]++
}

// GetEventStats returns event statistics
func (im *InformerManager) GetEventStats() map[string]int64 {
	im.eventMu.RLock()
	defer im.eventMu.RUnlock()
	
	stats := make(map[string]int64, len(im.eventCount))
	for k, v := range im.eventCount {
		stats[k] = v
	}
	
	return stats
}

// GetCacheStats returns cache statistics
func (im *InformerManager) GetCacheStats() map[string]interface{} {
	if !im.IsReady() {
		return map[string]interface{}{
			"ready": false,
		}
	}
	
	stats := map[string]interface{}{
		"ready": true,
		"cache_sizes": map[string]int{
			"pods":        len(im.informerFactory.Core().V1().Pods().Informer().GetStore().List()),
			"deployments": len(im.informerFactory.Apps().V1().Deployments().Informer().GetStore().List()),
			"services":    len(im.informerFactory.Core().V1().Services().Informer().GetStore().List()),
			"replicasets": len(im.informerFactory.Apps().V1().ReplicaSets().Informer().GetStore().List()),
			"nodes":       len(im.informerFactory.Core().V1().Nodes().Informer().GetStore().List()),
			"events":      len(im.informerFactory.Core().V1().Events().Informer().GetStore().List()),
		},
		"event_stats": im.GetEventStats(),
	}
	
	return stats
}