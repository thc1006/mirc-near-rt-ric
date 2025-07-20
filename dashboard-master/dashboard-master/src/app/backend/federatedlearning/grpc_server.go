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
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FLServiceServer implements the gRPC server for federated learning xApp communication.
type FLServiceServer struct {
	UnimplementedFLServiceServer // Embed for forward compatibility
	coordinator FederatedLearningManager
	logger      *slog.Logger
}

// NewFLServiceServer creates a new gRPC FL service server.
func NewFLServiceServer(coordinator FederatedLearningManager, logger *slog.Logger) *FLServiceServer {
	return &FLServiceServer{
		coordinator: coordinator,
		logger:      logger,
	}
}

// SubmitModelUpdate handles incoming model updates from xApps via gRPC.
func (s *FLServiceServer) SubmitModelUpdate(ctx context.Context, req *SubmitModelUpdateRequest) (*SubmitModelUpdateResponse, error) {
	s.logger.Info("Received gRPC model update request",
		slog.String("client_id", req.ClientID),
		slog.String("model_id", req.ModelID),
		slog.Int64("round", req.Round))

	// Convert gRPC request to internal ModelUpdate type
	modelUpdate := ModelUpdate{
		ClientID:         req.ClientID,
		ModelID:          req.ModelID,
		Round:            req.Round,
		Parameters:       req.Parameters,
		ParametersHash:   req.ParametersHash,
		DataSamplesCount: req.DataSamplesCount,
		LocalMetrics:     *req.LocalMetrics,
		Signature:        req.Signature,
		Timestamp:        time.Now(), // Use current time for server-side processing
		Metadata:         req.Metadata,
	}

	// Validate the model update using the coordinator's logic
	if err := s.coordinator.ValidateModelUpdate(ctx, modelUpdate); err != nil {
		s.logger.Error("Model update validation failed", slog.String("error", err.Error()))
		return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
	}

	// In a real implementation, this would likely push the update to a channel
	// or queue for asynchronous processing by the aggregation engine.
	// For this mock, we'll just log and return success.

	s.logger.Info("Model update accepted for processing",
		slog.String("client_id", req.ClientID),
		slog.String("model_id", req.ModelID))

	return &SubmitModelUpdateResponse{
		Status:    "ACCEPTED",
		Round:     req.Round,
		Timestamp: &modelUpdate.Timestamp,
		Message:   "Model update received and accepted for aggregation",
	}, nil
}

// StartGRPCServer starts the gRPC server for FL xApp communication.
func StartGRPCServer(coordinator FederatedLearningManager, logger *slog.Logger, port int, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	listenAddr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.Error("Failed to listen for gRPC", slog.String("address", listenAddr), slog.String("error", err.Error()))
		return
	}

	s := grpc.NewServer()
	RegisterFLServiceServer(s, NewFLServiceServer(coordinator, logger))

	logger.Info("gRPC server starting", slog.String("address", listenAddr))

	go func() {
		<-ctx.Done()
		logger.Info("gRPC server shutting down...")
		s.GracefulStop()
		logger.Info("gRPC server stopped.")
	}()

	if err := s.Serve(lis); err != nil {
		logger.Error("gRPC server failed to serve", slog.String("error", err.Error()))
	}
}

// UnimplementedFLServiceServer must be embedded to have forward compatible implementations.
type UnimplementedFLServiceServer struct {
}

func (UnimplementedFLServiceServer) SubmitModelUpdate(context.Context, *SubmitModelUpdateRequest) (*SubmitModelUpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitModelUpdate not implemented")
}
func (UnimplementedFLServiceServer) mustEmbedUnimplementedFLServiceServer() {}

// FLServiceServer is the server API for FLService service.
// All implementations must embed UnimplementedFLServiceServer
// for forward compatibility
type FLService interface {
	SubmitModelUpdate(ctx context.Context, in *SubmitModelUpdateRequest) (*SubmitModelUpdateResponse, error)
	mustEmbedUnimplementedFLServiceServer()
}

func RegisterFLServiceServer(s grpc.ServiceRegistrar, srv FLService) {
	s.RegisterService(&FLService_ServiceDesc, srv)
}

var FLService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "federatedlearning.FLService",
	HandlerType: (*FLServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SubmitModelUpdate",
			Handler:    _FLService_SubmitModelUpdate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc_service.proto", // Placeholder for actual proto file
}

func _FLService_SubmitModelUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitModelUpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FLService).SubmitModelUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "federatedlearning.FLService/SubmitModelUpdate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FLService).SubmitModelUpdate(ctx, req.(*SubmitModelUpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}
