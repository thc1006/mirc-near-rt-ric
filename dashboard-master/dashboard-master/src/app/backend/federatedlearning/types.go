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
	"sync"
	"time"

	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CoordinatorRole defines the role of a coordinator in the hierarchy.
type CoordinatorRole string

const (
	// MasterCoordinatorRole is the role for the master coordinator.
	MasterCoordinatorRole CoordinatorRole = "master"
	// RegionalCoordinatorRole is the role for a regional coordinator.
	RegionalCoordinatorRole CoordinatorRole = "regional"
)

// MasterCoordinator manages the global state of the federated learning process.
type MasterCoordinator struct {
	*FLCoordinator
	regionalCoordinators map[string]*RegionalCoordinatorClient
	regionalMutex        sync.RWMutex
}

// RegionalCoordinator manages a subset of clients in a specific region.
type RegionalCoordinator struct {
	*FLCoordinator
	masterClient *MasterCoordinatorClient
}

// RegionalCoordinatorClient is a client for a regional coordinator.
type RegionalCoordinatorClient struct {
	ID         string
	Endpoint   string
	Client     *grpc.ClientConn
	LastSeen   time.Time
	TrustScore float64
}

// MasterCoordinatorClient is a client for the master coordinator.
type MasterCoordinatorClient struct {
	Endpoint string
	Client   *grpc.ClientConn
}

// TrainingJob represents a federated learning training job.
type TrainingJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TrainingJobSpec   `json:"spec,omitempty"`
	Status            TrainingJobStatus `json:"status,omitempty"`
}

// TrainingJobSpec defines the desired state of a TrainingJob.
type TrainingJobSpec struct {
	ModelID        string                `json:"modelId"`
	RRMTask        RRMTaskType           `json:"rrmTask"`
	TrainingConfig TrainingConfiguration `json:"trainingConfig"`
	ClientSelector ClientSelector        `json:"clientSelector"`
	Resources      ResourceRequirements  `json:"resources"`
}

// TrainingJobStatus defines the observed state of a TrainingJob.
type TrainingJobStatus struct {
	Phase                TrainingStatus `json:"phase,omitempty"`
	Message              string         `json:"message,omitempty"`
	LastUpdate           metav1.Time    `json:"lastUpdate,omitempty"`
	ParticipatingClients []string       `json:"participatingClients,omitempty"`
}

// RRMTaskType defines the type of RRM task.
type RRMTaskType string

// ClientSelector defines the criteria for selecting clients for a training job.
type ClientSelector struct {
	MinTrustScore    float64     `json:"minTrustScore"`
	GeographicZone   string      `json:"geographicZone"`
	MatchRRMTasks    []RRMTaskType `json:"matchRrmTasks"`
	MaxClients       int         `json:"maxClients"`
}

// ResourceRequirements defines the resource requirements for a training job.
type ResourceRequirements struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	GPU    string `json:"gpu"`
}