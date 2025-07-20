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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Service implements the gRPC service for the federated learning coordinator.
type Service struct {
	UnimplementedFederatedLearningServer

	coordinator *FLCoordinator
	logger      *slog.Logger
}

// NewService creates a new gRPC service.
func NewService(coordinator *FLCoordinator, logger *slog.Logger) *Service {
	return &Service{
		coordinator: coordinator,
		logger:      logger,
	}
}

// Start starts the gRPC service.
func (s *Service) Start(ctx context.Context, config *CoordinatorConfig) error {
	lis, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	var opts []grpc.ServerOption
	if config.TLSEnabled {
		creds, err := credentials.NewServerTLSFromFile(config.CertificatePath, config.PrivateKeyPath)
		if err != nil {
			return fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	grpcServer := grpc.NewServer(opts...)
	RegisterFederatedLearningServer(grpcServer, s)

	s.coordinator.grpcServer = grpcServer

	s.logger.Info("gRPC server started", "address", config.ListenAddress)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			s.logger.Error("gRPC server failed", "error", err)
		}
	}()

	<-ctx.Done()
	s.logger.Info("Shutting down gRPC server")
	grpcServer.GracefulStop()

	return nil
}

// HandleTraining is the gRPC stream handler for training.
func (s *Service) HandleTraining(stream FederatedLearning_HandleTrainingServer) error {
	return s.coordinator.HandleTraining(stream)
}

// UnimplementedFederatedLearningServer is a placeholder for the unimplemented gRPC server.
type UnimplementedFederatedLearningServer struct{}

// RegisterFederatedLearningServer is a placeholder for the gRPC server registration.
func RegisterFederatedLearningServer(s *grpc.Server, srv *Service) {}
