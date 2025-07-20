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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// discoverRegionalCoordinators discovers regional coordinators using Kubernetes services.
func (c *FLCoordinator) discoverRegionalCoordinators(ctx context.Context) ([]*RegionalCoordinatorClient, error) {
	if c.role != MasterCoordinatorRole {
		return nil, fmt.Errorf("only master coordinator can discover regional coordinators")
	}

	services, err := c.kubeClient.CoreV1().Services("").List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/component=fl-coordinator,role=regional",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list regional coordinator services: %w", err)
	}

	var regionalCoordinators []*RegionalCoordinatorClient
	for _, service := range services.Items {
		regionalCoordinators = append(regionalCoordinators, &RegionalCoordinatorClient{
			ID:       service.Name,
			Endpoint: fmt.Sprintf("%s:%d", service.Spec.ClusterIP, service.Spec.Ports[0].Port),
		})
	}

	c.logger.Info("Discovered regional coordinators", slog.Int("count", len(regionalCoordinators)))
	return regionalCoordinators, nil
}
