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
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/google/uuid"
	
	"github.com/kubernetes/dashboard/src/app/backend/errors"
	"github.com/kubernetes/dashboard/src/app/backend/handler"
)

// FLAPIHandler provides REST API endpoints for federated learning coordination
type FLAPIHandler struct {
	coordinator FederatedLearningManager
	logger      *slog.Logger
}

// NewFLAPIHandler creates a new FL API handler
func NewFLAPIHandler(coordinator FederatedLearningManager, logger *slog.Logger) *FLAPIHandler {
	return &FLAPIHandler{
		coordinator: coordinator,
		logger:      logger,
	}
}

// RegisterRoutes registers all federated learning API routes
func (h *FLAPIHandler) RegisterRoutes(ws *restful.WebService) {
	// Client management endpoints
	ws.Route(ws.POST("/fl/clients").
		To(h.registerClient).
		Doc("Register a new federated learning client (xApp)").
		Metadata("rrm-task", "client-management").
		Reads(ClientRegistrationRequest{}).
		Writes(ClientRegistrationResponse{}).
		Returns(201, "Created", ClientRegistrationResponse{}).
		Returns(400, "Bad Request", errors.HTTPError{}).
		Returns(409, "Conflict", errors.HTTPError{}))

	ws.Route(ws.GET("/fl/clients").
		To(h.listClients).
		Doc("List registered federated learning clients").
		Param(ws.QueryParameter("rrm_task", "Filter by RRM task type").DataType("string")).
		Param(ws.QueryParameter("status", "Filter by client status").DataType("string")).
		Param(ws.QueryParameter("trust_score_min", "Minimum trust score").DataType("number")).
		Param(ws.QueryParameter("network_slice", "Filter by network slice").DataType("string")).
		Writes([]FLClient{}).
		Returns(200, "OK", []FLClient{}))

	ws.Route(ws.GET("/fl/clients/{client-id}").
		To(h.getClient).
		Doc("Get details of a specific federated learning client").
		Param(ws.PathParameter("client-id", "Client identifier")).
		Writes(FLClient{}).
		Returns(200, "OK", FLClient{}).
		Returns(404, "Not Found", errors.HTTPError{}))

	ws.Route(ws.PUT("/fl/clients/{client-id}").
		To(h.updateClient).
		Doc("Update federated learning client information").
		Param(ws.PathParameter("client-id", "Client identifier")).
		Reads(ClientUpdateRequest{}).
		Writes(FLClient{}).
		Returns(200, "OK", FLClient{}).
		Returns(404, "Not Found", errors.HTTPError{}))

	ws.Route(ws.DELETE("/fl/clients/{client-id}").
		To(h.unregisterClient).
		Doc("Unregister a federated learning client").
		Param(ws.PathParameter("client-id", "Client identifier")).
		Returns(204, "No Content", nil).
		Returns(404, "Not Found", errors.HTTPError{}).
		Returns(409, "Conflict", errors.HTTPError{}))

	ws.Route(ws.POST("/fl/clients/{client-id}/heartbeat").
		To(h.clientHeartbeat).
		Doc("Client heartbeat to maintain connection").
		Param(ws.PathParameter("client-id", "Client identifier")).
		Reads(HeartbeatRequest{}).
		Writes(HeartbeatResponse{}).
		Returns(200, "OK", HeartbeatResponse{}).
		Returns(404, "Not Found", errors.HTTPError{}))

	// Model management endpoints
	ws.Route(ws.POST("/fl/models").
		To(h.createGlobalModel).
		Doc("Create a new global federated learning model").
		Reads(CreateModelRequest{}).
		Writes(GlobalModel{}).
		Returns(201, "Created", GlobalModel{}).
		Returns(400, "Bad Request", errors.HTTPError{}))

	ws.Route(ws.GET("/fl/models").
		To(h.listGlobalModels).
		Doc("List global federated learning models").
		Param(ws.QueryParameter("rrm_task", "Filter by RRM task type").DataType("string")).
		Param(ws.QueryParameter("status", "Filter by training status").DataType("string")).
		Writes([]GlobalModel{}).
		Returns(200, "OK", []GlobalModel{}))

	ws.Route(ws.GET("/fl/models/{model-id}").
		To(h.getGlobalModel).
		Doc("Get details of a specific global model").
		Param(ws.PathParameter("model-id", "Model identifier")).
		Writes(GlobalModel{}).
		Returns(200, "OK", GlobalModel{}).
		Returns(404, "Not Found", errors.HTTPError{}))

	ws.Route(ws.GET("/fl/models/{model-id}/download").
		To(h.downloadGlobalModel).
		Doc("Download global model parameters for training").
		Param(ws.PathParameter("model-id", "Model identifier")).
		Param(ws.HeaderParameter("X-Client-ID", "Client identifier")).
		Produces(restful.MIME_OCTET).
		Returns(200, "OK", []byte{}).
		Returns(404, "Not Found", errors.HTTPError{}).
		Returns(403, "Forbidden", errors.HTTPError{}))

	// Training coordination endpoints
	ws.Route(ws.POST("/fl/training").
		To(h.startTraining).
		Doc("Start a new federated learning training job").
		Reads(StartTrainingRequest{}).
		Writes(TrainingJob{}).
		Returns(201, "Created", TrainingJob{}).
		Returns(400, "Bad Request", errors.HTTPError{}).
		Returns(503, "Service Unavailable", errors.HTTPError{}))

	ws.Route(ws.GET("/fl/training").
		To(h.listTrainingJobs).
		Doc("List federated learning training jobs").
		Param(ws.QueryParameter("status", "Filter by training status").DataType("string")).
		Param(ws.QueryParameter("rrm_task", "Filter by RRM task type").DataType("string")).
		Writes([]TrainingJob{}).
		Returns(200, "OK", []TrainingJob{}))

	ws.Route(ws.GET("/fl/training/{job-id}").
		To(h.getTrainingJob).
		Doc("Get details of a specific training job").
		Param(ws.PathParameter("job-id", "Training job identifier")).
		Writes(TrainingJob{}).
		Returns(200, "OK", TrainingJob{}).
		Returns(404, "Not Found", errors.HTTPError{}))

	ws.Route(ws.POST("/fl/training/{job-id}/stop").
		To(h.stopTraining).
		Doc("Stop a running training job").
		Param(ws.PathParameter("job-id", "Training job identifier")).
		Returns(200, "OK", nil).
		Returns(404, "Not Found", errors.HTTPError{}))

	// Model update endpoints (for xApps to submit local updates)
	ws.Route(ws.POST("/fl/training/{job-id}/updates").
		To(h.submitModelUpdate).
		Doc("Submit local model update from xApp client").
		Param(ws.PathParameter("job-id", "Training job identifier")).
		Param(ws.HeaderParameter("X-Client-ID", "Client identifier")).
		Reads(ModelUpdateRequest{}).
		Writes(ModelUpdateResponse{}).
		Returns(200, "OK", ModelUpdateResponse{}).
		Returns(400, "Bad Request", errors.HTTPError{}).
		Returns(404, "Not Found", errors.HTTPError{}).
		Returns(403, "Forbidden", errors.HTTPError{}))

	ws.Route(ws.GET("/fl/training/{job-id}/round").
		To(h.getCurrentRound).
		Doc("Get current training round information").
		Param(ws.PathParameter("job-id", "Training job identifier")).
		Param(ws.HeaderParameter("X-Client-ID", "Client identifier")).
		Writes(RoundInfo{}).
		Returns(200, "OK", RoundInfo{}).
		Returns(404, "Not Found", errors.HTTPError{}))

	// Metrics and monitoring endpoints
	ws.Route(ws.GET("/fl/metrics/models/{model-id}").
		To(h.getModelMetrics).
		Doc("Get performance metrics for a specific model").
		Param(ws.PathParameter("model-id", "Model identifier")).
		Writes(ModelMetrics{}).
		Returns(200, "OK", ModelMetrics{}).
		Returns(404, "Not Found", errors.HTTPError{}))

	ws.Route(ws.GET("/fl/metrics/clients/{client-id}").
		To(h.getClientMetrics).
		Doc("Get performance metrics for a specific client").
		Param(ws.PathParameter("client-id", "Client identifier")).
		Writes(ClientMetrics{}).
		Returns(200, "OK", ClientMetrics{}).
		Returns(404, "Not Found", errors.HTTPError{}))

	ws.Route(ws.GET("/fl/metrics/training/{job-id}").
		To(h.getTrainingMetrics).
		Doc("Get training metrics for a specific job").
		Param(ws.PathParameter("job-id", "Training job identifier")).
		Writes(TrainingMetrics{}).
		Returns(200, "OK", TrainingMetrics{}).
		Returns(404, "Not Found", errors.HTTPError{}))

	// Health and status endpoints
	ws.Route(ws.GET("/fl/health").
		To(h.healthCheck).
		Doc("Health check for federated learning service").
		Writes(HealthStatus{}).
		Returns(200, "OK", HealthStatus{}))

	ws.Route(ws.GET("/fl/status").
		To(h.systemStatus).
		Doc("Get federated learning system status").
		Writes(SystemStatus{}).
		Returns(200, "OK", SystemStatus{}))
}

// Client Registration

// registerClient handles xApp client registration
func (h *FLAPIHandler) registerClient(request *restful.Request, response *restful.Response) {
	var req ClientRegistrationRequest
	if err := request.ReadEntity(&req); err != nil {
		errors.HandleError(response.ResponseWriter, 
			errors.NewInvalid(fmt.Sprintf("invalid registration request: %v", err)))
		return
	}

	// Validate request
	if err := h.validateClientRegistrationRequest(&req); err != nil {
		errors.HandleError(response.ResponseWriter, err)
		return
	}

	// Create FL client from request
	client := &FLClient{
		XAppName:         req.XAppName,
		XAppVersion:      req.XAppVersion,
		Namespace:        req.Namespace,
		Endpoint:         req.Endpoint,
		RRMTasks:         req.RRMTasks,
		ModelFormats:     req.ModelFormats,
		ComputeResources: req.ComputeResources,
		NetworkSlices:    req.NetworkSlices,
		Metadata:         req.Metadata,
	}

	// Register client
	if err := h.coordinator.RegisterClient(request.Request.Context(), client); err != nil {
		errors.HandleError(response.ResponseWriter, err)
		return
	}

	// Prepare response
	resp := ClientRegistrationResponse{
		ClientID:    client.ID,
		Status:      string(client.Status),
		TrustScore:  client.TrustScore,
		Certificate: client.Certificate,
		Endpoint:    fmt.Sprintf("/fl/clients/%s", client.ID),
	}

	response.WriteHeaderAndEntity(http.StatusCreated, resp)
	
	h.logger.Info("Client registered successfully",
		slog.String("client_id", client.ID),
		slog.String("xapp_name", client.XAppName))
}

// listClients handles listing of registered clients
func (h *FLAPIHandler) listClients(request *restful.Request, response *restful.Response) {
	// Parse query parameters
	selector := ClientSelector{}

	if rrmTask := request.QueryParameter("rrm_task"); rrmTask != "" {
		selector.MatchRRMTasks = []RRMTaskType{RRMTaskType(rrmTask)}
	}

	if minTrustScore := request.QueryParameter("trust_score_min"); minTrustScore != "" {
		if score, err := strconv.ParseFloat(minTrustScore, 64); err == nil {
			selector.MinTrustScore = score
		}
	}

	if networkSlice := request.QueryParameter("network_slice"); networkSlice != "" {
		selector.MatchLabels = map[string]string{"network_slice": networkSlice}
	}

	// Get clients
	clients, err := h.coordinator.ListClients(request.Request.Context(), selector)
	if err != nil {
		errors.HandleError(response.ResponseWriter, err)
		return
	}

	// Filter by status if specified
	if statusFilter := request.QueryParameter("status"); statusFilter != "" {
		var filteredClients []*FLClient
		for _, client := range clients {
			if string(client.Status) == statusFilter {
				filteredClients = append(filteredClients, client)
			}
		}
		clients = filteredClients
	}

	response.WriteEntity(clients)
}

// getClient handles getting a specific client
func (h *FLAPIHandler) getClient(request *restful.Request, response *restful.Response) {
	clientID := request.PathParameter("client-id")
	
	client, err := h.coordinator.GetClient(request.Request.Context(), clientID)
	if err != nil {
		errors.HandleError(response.ResponseWriter, err)
		return
	}

	response.WriteEntity(client)
}

// updateClient handles updating client information
func (h *FLAPIHandler) updateClient(request *restful.Request, response *restful.Response) {
	clientID := request.PathParameter("client-id")
	
	var req ClientUpdateRequest
	if err := request.ReadEntity(&req); err != nil {
		errors.HandleError(response.ResponseWriter, 
			errors.NewInvalid(fmt.Sprintf("invalid update request: %v", err)))
		return
	}

	// Get existing client
	client, err := h.coordinator.GetClient(request.Request.Context(), clientID)
	if err != nil {
		errors.HandleError(response.ResponseWriter, err)
		return
	}

	// Update fields
	if req.Endpoint != "" {
		client.Endpoint = req.Endpoint
	}
	if req.ComputeResources != nil {
		client.ComputeResources = *req.ComputeResources
	}
	if req.NetworkSlices != nil {
		client.NetworkSlices = req.NetworkSlices
	}
	if req.Metadata != nil {
		client.Metadata = req.Metadata
	}
	if req.Status != "" {
		if err := h.coordinator.UpdateClientStatus(request.Request.Context(), clientID, FLClientStatus(req.Status)); err != nil {
			errors.HandleError(response.ResponseWriter, err)
			return
		}
	}

	response.WriteEntity(client)
}

// unregisterClient handles client unregistration
func (h *FLAPIHandler) unregisterClient(request *restful.Request, response *restful.Response) {
	clientID := request.PathParameter("client-id")
	
	if err := h.coordinator.UnregisterClient(request.Request.Context(), clientID); err != nil {
		errors.HandleError(response.ResponseWriter, err)
		return
	}

	response.WriteHeader(http.StatusNoContent)
	
	h.logger.Info("Client unregistered successfully", slog.String("client_id", clientID))
}

// clientHeartbeat handles client heartbeat requests
func (h *FLAPIHandler) clientHeartbeat(request *restful.Request, response *restful.Response) {
	clientID := request.PathParameter("client-id")
	
	var req HeartbeatRequest
	if err := request.ReadEntity(&req); err != nil {
		errors.HandleError(response.ResponseWriter, 
			errors.NewInvalid(fmt.Sprintf("invalid heartbeat request: %v", err)))
		return
	}

	// Update client status to indicate it's alive
	if err := h.coordinator.UpdateClientStatus(request.Request.Context(), clientID, FLClientStatusIdle); err != nil {
		errors.HandleError(response.ResponseWriter, err)
		return
	}

	// Prepare heartbeat response
	resp := HeartbeatResponse{
		Timestamp:    time.Now(),
		Status:       "OK",
		Instructions: []string{}, // Could include training instructions
	}

	response.WriteEntity(resp)
}

// Model Management

// createGlobalModel handles creation of new global models
func (h *FLAPIHandler) createGlobalModel(request *restful.Request, response *restful.Response) {
	var req CreateModelRequest
	if err := request.ReadEntity(&req); err != nil {
		errors.HandleError(response.ResponseWriter, 
			errors.NewInvalid(fmt.Sprintf("invalid model creation request: %v", err)))
		return
	}

	// Validate request
	if err := h.validateCreateModelRequest(&req); err != nil {
		errors.HandleError(response.ResponseWriter, err)
		return
	}

	// Create global model
	model := &GlobalModel{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Version:         req.Version,
		RRMTask:         req.RRMTask,
		Format:          req.Format,
		Architecture:    req.Architecture,
		TrainingConfig:  req.TrainingConfig,
		AggregationAlg:  req.AggregationAlgorithm,
		PrivacyMech:     req.PrivacyMechanism,
		PrivacyParams:   req.PrivacyParams,
		MaxRounds:       req.MaxRounds,
		TargetAccuracy:  req.TargetAccuracy,
		Status:          TrainingStatusInitializing,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		AllowedClients:  req.AllowedClients,
		NetworkSlices:   req.NetworkSlices,
		Metadata:        req.Metadata,
	}

	if err := h.coordinator.CreateGlobalModel(request.Request.Context(), model); err != nil {
		errors.HandleError(response.ResponseWriter, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, model)
	
	h.logger.Info("Global model created successfully",
		slog.String("model_id", model.ID),
		slog.String("rrm_task", string(model.RRMTask)))
}

// Training Coordination

// startTraining handles starting new training jobs
func (h *FLAPIHandler) startTraining(request *restful.Request, response *restful.Response) {
	var req StartTrainingRequest
	if err := request.ReadEntity(&req); err != nil {
		errors.HandleError(response.ResponseWriter, 
			errors.NewInvalid(fmt.Sprintf("invalid training request: %v", err)))
		return
	}

	// Validate request
	if err := h.validateStartTrainingRequest(&req); err != nil {
		errors.HandleError(response.ResponseWriter, err)
		return
	}

	// Convert request to job spec
	jobSpec := TrainingJobSpec{
		ModelID:        req.ModelID,
		RRMTask:        req.RRMTask,
		TrainingConfig: req.TrainingConfig,
		ClientSelector: req.ClientSelector,
		PrivacyConfig:  req.PrivacyConfig,
		Schedule:       req.Schedule,
		Priority:       req.Priority,
		Resources:      req.Resources,
		NetworkSlices:  req.NetworkSlices,
	}

	if req.Deadline != "" {
		if deadline, err := time.Parse(time.RFC3339, req.Deadline); err == nil {
			jobSpec.Deadline = &deadline
		}
	}

	// Start training
	job, err := h.coordinator.StartTraining(request.Request.Context(), jobSpec)
	if err != nil {
		errors.HandleError(response.ResponseWriter, err)
		return
	}

	response.WriteHeaderAndEntity(http.StatusCreated, job)
	
	h.logger.Info("Training job started successfully",
		slog.String("job_id", job.Name),
		slog.String("model_id", req.ModelID))
}

// submitModelUpdate handles model updates from xApp clients
func (h *FLAPIHandler) submitModelUpdate(request *restful.Request, response *restful.Response) {
	jobID := request.PathParameter("job-id")
	clientID := request.HeaderParameter("X-Client-ID")

	if clientID == "" {
		errors.HandleError(response.ResponseWriter, 
			errors.NewInvalid("X-Client-ID header is required"))
		return
	}

	var req ModelUpdateRequest
	if err := request.ReadEntity(&req); err != nil {
		errors.HandleError(response.ResponseWriter, 
			errors.NewInvalid(fmt.Sprintf("invalid model update: %v", err)))
		return
	}

	// Create model update
	update := ModelUpdate{
		ClientID:         clientID,
		ModelID:          req.ModelID,
		Round:            req.Round,
		Parameters:       req.Parameters,
		ParametersHash:   req.ParametersHash,
		DataSamplesCount: req.DataSamplesCount,
		LocalMetrics:     req.LocalMetrics,
		Signature:        req.Signature,
		Timestamp:        time.Now(),
		Metadata:         req.Metadata,
	}

	// Validate model update
	if err := h.coordinator.ValidateModelUpdate(request.Request.Context(), update); err != nil {
		errors.HandleError(response.ResponseWriter, err)
		return
	}

	// Process the update (this would typically be handled asynchronously)
	resp := ModelUpdateResponse{
		Status:    "accepted",
		Round:     update.Round,
		Timestamp: time.Now(),
		Message:   "Model update received and validated",
	}

	response.WriteEntity(resp)
	
	h.logger.Debug("Model update submitted",
		slog.String("job_id", jobID),
		slog.String("client_id", clientID),
		slog.Int64("round", update.Round))
}

// Health and Status

// healthCheck handles health check requests
func (h *FLAPIHandler) healthCheck(request *restful.Request, response *restful.Response) {
	status := HealthStatus{
		Status:    "OK",
		Timestamp: time.Now(),
		Services: map[string]string{
			"coordinator": "healthy",
			"storage":     "healthy",
			"crypto":      "healthy",
		},
	}

	response.WriteEntity(status)
}

// Helper methods for validation

func (h *FLAPIHandler) validateClientRegistrationRequest(req *ClientRegistrationRequest) error {
	if req.XAppName == "" {
		return errors.NewInvalid("xapp_name is required")
	}
	if req.Endpoint == "" {
		return errors.NewInvalid("endpoint is required")
	}
	if len(req.RRMTasks) == 0 {
		return errors.NewInvalid("at least one RRM task is required")
	}
	return nil
}

func (h *FLAPIHandler) validateCreateModelRequest(req *CreateModelRequest) error {
	if req.Name == "" {
		return errors.NewInvalid("model name is required")
	}
	if req.RRMTask == "" {
		return errors.NewInvalid("RRM task is required")
	}
	if req.Format == "" {
		return errors.NewInvalid("model format is required")
	}
	return nil
}

func (h *FLAPIHandler) validateStartTrainingRequest(req *StartTrainingRequest) error {
	if req.ModelID == "" {
		return errors.NewInvalid("model_id is required")
	}
	if req.RRMTask == "" {
		return errors.NewInvalid("rrm_task is required")
	}
	if req.TrainingConfig.MinParticipants <= 0 {
		return errors.NewInvalid("min_participants must be greater than 0")
	}
	return nil
}