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
	"log/slog"

	"k8s.io/client-go/kubernetes"
)

// NewResourceManager creates a new resource manager.
func NewResourceManager(logger *slog.Logger, kubeClient kubernetes.Interface) *AdaptiveResourceManager {
	return NewAdaptiveResourceManager(logger, kubeClient)
}

// ReserveResources reserves resources for a training job.
func (rm *AdaptiveResourceManager) ReserveResources(ctx context.Context, requirements *ResourceRequirements) error {
	// 1. Predict resource needs
	predictedRequirements, err := rm.capacityPredictor.PredictRequiredResources(&Workload{})
	if err != nil {
		return err
	}

	// 2. Enforce quotas
	// TODO: Get slice ID from requirements.
	if err := rm.quotaManager.EnforceQuotas("slice-id", predictedRequirements); err != nil {
		return err
	}

	// 3. Scale resources
	// TODO: Get deployment name from requirements.
	if err := rm.autoScaler.Scale("deployment-name", 5); err != nil {
		return err
	}

	return nil
}

// ReleaseResources releases previously allocated resources.
func (rm *AdaptiveResourceManager) ReleaseResources(ctx context.Context, allocationID string) error {
	// TODO: Implement resource release logic.
	return nil
}
