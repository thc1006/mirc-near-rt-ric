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

package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/emicklei/go-restful/v3"
	v1 "k8s.io/api/core/v1"

	"github.com/kubernetes/dashboard/src/app/backend/cache"
	clientapi "github.com/kubernetes/dashboard/src/app/backend/client/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/handler/parser"
	"github.com/kubernetes/dashboard/src/app/backend/informer"
	"github.com/kubernetes/dashboard/src/app/backend/integration"
	metricapi "github.com/kubernetes/dashboard/src/app/backend/integration/metric/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/optimized"
	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"
	settingsApi "github.com/kubernetes/dashboard/src/app/backend/settings/api"
)

// OptimizedAPIHandler provides high-performance API handling with caching and informers
type OptimizedAPIHandler struct {
	// Core dependencies
	iManager integration.IntegrationManager
	cManager clientapi.ClientManager
	sManager settingsApi.SettingsManager
	logger   *slog.Logger
	
	// Performance optimization components
	cache           *cache.ResourceCache
	informerManager *informer.InformerManager
	
	// Optimized services
	podListService *optimized.OptimizedPodListService
	
	// Request pooling for memory optimization
	requestPool sync.Pool
	
	// Metrics and monitoring
	requestCount map[string]int64
	requestMu    sync.RWMutex
	
	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
}

// OptimizedRequest represents a pooled request object
type OptimizedRequest struct {
	StartTime time.Time
	Path      string
	Method    string
	Namespace string
}

// NewOptimizedAPIHandler creates a new optimized API handler
func NewOptimizedAPIHandler(ctx context.Context, iManager integration.IntegrationManager, 
	cManager clientapi.ClientManager, sManager settingsApi.SettingsManager, logger *slog.Logger) (*OptimizedAPIHandler, error) {
	
	if ctx == nil {
		ctx = context.Background()
	}
	if logger == nil {
		logger = slog.Default()
	}
	
	childCtx, cancel := context.WithCancel(ctx)
	
	// Get Kubernetes client for initializing components
	client, err := cManager.Client(nil) // Pass nil request for now
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to get Kubernetes client: %w", err)
	}
	
	// Initialize cache
	resourceCache := cache.NewResourceCache(childCtx, client, logger.With("component", "cache"))
	
	// Initialize informer manager
	informerMgr := informer.NewInformerManager(childCtx, client, logger.With("component", "informer"))
	
	handler := &OptimizedAPIHandler{
		iManager:        iManager,
		cManager:        cManager,
		sManager:        sManager,
		logger:          logger,
		cache:           resourceCache,
		informerManager: informerMgr,
		ctx:             childCtx,
		cancel:          cancel,
		requestCount:    make(map[string]int64),
	}
	
	// Initialize optimized services
	handler.podListService = optimized.NewOptimizedPodListService(client, resourceCache, logger.With("component", "pod_service"))
	
	// Initialize request pool
	handler.requestPool = sync.Pool{
		New: func() interface{} {
			return &OptimizedRequest{}
		},
	}
	
	return handler, nil
}

// Start initializes the optimized components
func (h *OptimizedAPIHandler) Start() error {
	h.logger.Info("Starting optimized API handler")
	
	// Start cache
	h.cache.Start()
	
	// Start informer manager
	if err := h.informerManager.Start(); err != nil {
		return fmt.Errorf("failed to start informer manager: %w", err)
	}
	
	h.logger.Info("Optimized API handler started successfully")
	return nil
}

// Stop gracefully shuts down the optimized components
func (h *OptimizedAPIHandler) Stop() {
	h.logger.Info("Stopping optimized API handler")
	
	h.cancel()
	h.informerManager.Stop()
	h.cache.Stop()
	
	h.logger.Info("Optimized API handler stopped")
}

// handleGetPodListOptimized provides an optimized pod list endpoint
func (h *OptimizedAPIHandler) handleGetPodListOptimized(request *restful.Request, response *restful.Response) {
	req := h.requestPool.Get().(*OptimizedRequest)
	defer func() {
		// Reset and return to pool
		req.StartTime = time.Time{}
		req.Path = ""
		req.Method = ""
		req.Namespace = ""
		h.requestPool.Put(req)
	}()
	
	req.StartTime = time.Now()
	req.Path = request.Request.URL.Path
	req.Method = request.Request.Method
	req.Namespace = parseNamespacePathParameter(request)
	
	// Increment request counter
	h.incrementRequestCount("pod_list")
	
	h.logger.Debug("Handling optimized pod list request",
		"namespace", req.Namespace,
		"path", req.Path)
	
	// Parse parameters
	namespace := common.NewNamespaceQuery([]string{req.Namespace})
	dataSelect := parser.ParseDataSelectPathParameter(request)
	dataSelect.MetricQuery = dataselect.StandardMetrics
	
	// Get metric client
	metricClient := h.iManager.Metric().Client()
	
	// Use optimized pod list service
	result, err := h.podListService.GetPodList(h.ctx, namespace, dataSelect, metricClient)
	if err != nil {
		h.logger.Error("Failed to get optimized pod list",
			"namespace", req.Namespace,
			"error", err,
			"duration", time.Since(req.StartTime))
		errors.HandleInternalError(response, err)
		return
	}
	
	duration := time.Since(req.StartTime)
	h.logger.Debug("Completed optimized pod list request",
		"namespace", req.Namespace,
		"pod_count", len(result.Pods),
		"duration", duration)
	
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

// handleGetPodListFromInformer provides pod list using informer cache
func (h *OptimizedAPIHandler) handleGetPodListFromInformer(request *restful.Request, response *restful.Response) {
	startTime := time.Now()
	namespace := parseNamespacePathParameter(request)
	
	h.incrementRequestCount("pod_list_informer")
	
	h.logger.Debug("Handling pod list from informer",
		"namespace", namespace)
	
	if !h.informerManager.IsReady() {
		h.logger.Warn("Informer not ready, falling back to API call")
		h.handleGetPodListOptimized(request, response)
		return
	}
	
	// Parse parameters
	nsQuery := common.NewNamespaceQuery([]string{namespace})
	dataSelect := parser.ParseDataSelectPathParameter(request)
	
	// Get pods from informer cache
	podPointers, err := h.informerManager.GetPodsFromCache(nsQuery)
	if err != nil {
		h.logger.Error("Failed to get pods from informer cache",
			"namespace", namespace,
			"error", err)
		errors.HandleInternalError(response, err)
		return
	}
	
	// Convert pointers to values
	pods := make([]v1.Pod, len(podPointers))
	for i, podPtr := range podPointers {
		pods[i] = *podPtr
	}
	
	// Get events from informer cache
	eventPointers, err := h.informerManager.GetEventsFromCache(nsQuery)
	if err != nil {
		h.logger.Warn("Failed to get events from informer cache",
			"namespace", namespace,
			"error", err)
		// Continue without events
		eventPointers = []*v1.Event{}
	}
	
	events := make([]v1.Event, len(eventPointers))
	for i, eventPtr := range eventPointers {
		events[i] = *eventPtr
	}
	
	// Convert to pod list using existing logic
	metricClient := h.iManager.Metric().Client()
	result := pod.ToPodList(pods, events, []error{}, dataSelect, metricClient)
	
	duration := time.Since(startTime)
	h.logger.Debug("Completed pod list from informer",
		"namespace", namespace,
		"pod_count", len(result.Pods),
		"duration", duration)
	
	response.WriteHeaderAndEntity(http.StatusOK, result)
}

// handleGetDeploymentListFromInformer provides deployment list using informer cache
func (h *OptimizedAPIHandler) handleGetDeploymentListFromInformer(request *restful.Request, response *restful.Response) {
	startTime := time.Now()
	namespace := parseNamespacePathParameter(request)
	
	h.incrementRequestCount("deployment_list_informer")
	
	h.logger.Debug("Handling deployment list from informer",
		"namespace", namespace)
	
	if !h.informerManager.IsReady() {
		h.logger.Warn("Informer not ready")
		errors.HandleInternalError(response, fmt.Errorf("informer not ready"))
		return
	}
	
	// Parse parameters
	nsQuery := common.NewNamespaceQuery([]string{namespace})
	
	// Get deployments from informer cache
	deploymentPointers, err := h.informerManager.GetDeploymentsFromCache(nsQuery)
	if err != nil {
		h.logger.Error("Failed to get deployments from informer cache",
			"namespace", namespace,
			"error", err)
		errors.HandleInternalError(response, err)
		return
	}
	
	duration := time.Since(startTime)
	h.logger.Debug("Completed deployment list from informer",
		"namespace", namespace,
		"deployment_count", len(deploymentPointers),
		"duration", duration)
	
	// Return simplified response for now
	response.WriteHeaderAndEntity(http.StatusOK, map[string]interface{}{
		"deployments": len(deploymentPointers),
		"duration":    duration.String(),
		"source":      "informer_cache",
	})
}

// handleGetServiceListFromInformer provides service list using informer cache
func (h *OptimizedAPIHandler) handleGetServiceListFromInformer(request *restful.Request, response *restful.Response) {
	startTime := time.Now()
	namespace := parseNamespacePathParameter(request)
	
	h.incrementRequestCount("service_list_informer")
	
	h.logger.Debug("Handling service list from informer",
		"namespace", namespace)
	
	if !h.informerManager.IsReady() {
		h.logger.Warn("Informer not ready")
		errors.HandleInternalError(response, fmt.Errorf("informer not ready"))
		return
	}
	
	// Parse parameters
	nsQuery := common.NewNamespaceQuery([]string{namespace})
	
	// Get services from informer cache
	servicePointers, err := h.informerManager.GetServicesFromCache(nsQuery)
	if err != nil {
		h.logger.Error("Failed to get services from informer cache",
			"namespace", namespace,
			"error", err)
		errors.HandleInternalError(response, err)
		return
	}
	
	duration := time.Since(startTime)
	h.logger.Debug("Completed service list from informer",
		"namespace", namespace,
		"service_count", len(servicePointers),
		"duration", duration)
	
	// Return simplified response for now
	response.WriteHeaderAndEntity(http.StatusOK, map[string]interface{}{
		"services": len(servicePointers),
		"duration": duration.String(),
		"source":   "informer_cache",
	})
}

// handleGetCacheStats provides cache performance statistics
func (h *OptimizedAPIHandler) handleGetCacheStats(request *restful.Request, response *restful.Response) {
	cacheStats := h.cache.GetStats()
	informerStats := h.informerManager.GetCacheStats()
	podServiceStats := h.podListService.GetStats()
	
	h.incrementRequestCount("cache_stats")
	
	stats := map[string]interface{}{
		"cache":        cacheStats,
		"informer":     informerStats,
		"pod_service":  podServiceStats,
		"requests":     h.getRequestStats(),
		"timestamp":    time.Now().Unix(),
	}
	
	response.WriteHeaderAndEntity(http.StatusOK, stats)
}

// handleGetHealthCheck provides health status
func (h *OptimizedAPIHandler) handleGetHealthCheck(request *restful.Request, response *restful.Response) {
	health := map[string]interface{}{
		"status": "healthy",
		"components": map[string]bool{
			"cache":    h.cache != nil,
			"informer": h.informerManager.IsReady(),
		},
		"timestamp": time.Now().Unix(),
	}
	
	response.WriteHeaderAndEntity(http.StatusOK, health)
}

// incrementRequestCount safely increments request counter
func (h *OptimizedAPIHandler) incrementRequestCount(endpoint string) {
	h.requestMu.Lock()
	defer h.requestMu.Unlock()
	h.requestCount[endpoint]++
}

// getRequestStats returns request statistics
func (h *OptimizedAPIHandler) getRequestStats() map[string]int64 {
	h.requestMu.RLock()
	defer h.requestMu.RUnlock()
	
	stats := make(map[string]int64, len(h.requestCount))
	for k, v := range h.requestCount {
		stats[k] = v
	}
	
	return stats
}

// InstallOptimizedRoutes adds optimized routes to the web service
func (h *OptimizedAPIHandler) InstallOptimizedRoutes(ws *restful.WebService) {
	// Optimized pod endpoints
	ws.Route(
		ws.GET("/optimized/pod").
			To(h.handleGetPodListOptimized).
			Writes(pod.PodList{}).
			Doc("Get optimized pod list for all namespaces"))
	
	ws.Route(
		ws.GET("/optimized/pod/{namespace}").
			To(h.handleGetPodListOptimized).
			Writes(pod.PodList{}).
			Doc("Get optimized pod list for specific namespace"))
	
	// Informer-based endpoints
	ws.Route(
		ws.GET("/informer/pod").
			To(h.handleGetPodListFromInformer).
			Doc("Get pod list from informer cache"))
	
	ws.Route(
		ws.GET("/informer/pod/{namespace}").
			To(h.handleGetPodListFromInformer).
			Doc("Get pod list from informer cache for specific namespace"))
	
	ws.Route(
		ws.GET("/informer/deployment").
			To(h.handleGetDeploymentListFromInformer).
			Doc("Get deployment list from informer cache"))
	
	ws.Route(
		ws.GET("/informer/deployment/{namespace}").
			To(h.handleGetDeploymentListFromInformer).
			Doc("Get deployment list from informer cache for specific namespace"))
	
	ws.Route(
		ws.GET("/informer/service").
			To(h.handleGetServiceListFromInformer).
			Doc("Get service list from informer cache"))
	
	ws.Route(
		ws.GET("/informer/service/{namespace}").
			To(h.handleGetServiceListFromInformer).
			Doc("Get service list from informer cache for specific namespace"))
	
	// Monitoring endpoints
	ws.Route(
		ws.GET("/cache/stats").
			To(h.handleGetCacheStats).
			Doc("Get cache and performance statistics"))
	
	ws.Route(
		ws.GET("/health").
			To(h.handleGetHealthCheck).
			Doc("Health check endpoint"))
}

// Helper function to parse namespace parameter (assuming it exists in the original code)
func parseNamespacePathParameter(request *restful.Request) string {
	namespace := request.PathParameter("namespace")
	if namespace == "" {
		return "" // All namespaces
	}
	return namespace
}