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
	"log/slog"
	"sort"
)

// AdaptiveCompressionManager manages adaptive compression strategies.
type AdaptiveCompressionManager struct {
	logger *slog.Logger
}

// NewAdaptiveCompressionManager creates a new AdaptiveCompressionManager.
func NewAdaptiveCompressionManager(logger *slog.Logger) *AdaptiveCompressionManager {
	return &AdaptiveCompressionManager{logger: logger}
}

// CompressModel applies adaptive compression to the given model.
func (m *AdaptiveCompressionManager) CompressModel(model *GlobalModel, networkConditions *NetworkConditions) (*GlobalModel, error) {
	// 1. Determine compression level based on network conditions.
	compressionLevel := m.determineCompressionLevel(networkConditions)

	// 2. Apply quantization.
	quantizedModel, err := m.quantize(model, compressionLevel)
	if err != nil {
		return nil, err
	}

	// 3. Apply sparsification.
	sparsifiedModel, err := m.sparsify(quantizedModel, compressionLevel)
	if err != nil {
		return nil, err
	}

	return sparsifiedModel, nil
}

// determineCompressionLevel determines the appropriate compression level based on network conditions.
func (m *AdaptiveCompressionManager) determineCompressionLevel(networkConditions *NetworkConditions) CompressionLevel {
	// Placeholder for adaptive compression logic.
	if networkConditions.E2Latency > 100 {
		return CompressionLevelHigh
	}
	return CompressionLevelMedium
}

// quantize applies quantization to the model parameters.
func (m *AdaptiveCompressionManager) quantize(model *GlobalModel, level CompressionLevel) (*GlobalModel, error) {
	// Placeholder for quantization logic.
	return model, nil
}

// sparsify applies sparsification to the model gradients.
func (m *AdaptiveCompressionManager) sparsify(model *GlobalModel, level CompressionLevel) (*GlobalModel, error) {
	// Placeholder for sparsification logic.
	return model, nil
}

// TopKSparsification performs top-k sparsification on the given data.
func (m *AdaptiveCompressionManager) TopKSparsification(data []float32, k int) ([]float32, error) {
	if k > len(data) {
		k = len(data)
	}

	// Create a copy to avoid modifying the original data.
	sortedData := make([]float32, len(data))
	copy(sortedData, data)
	sort.Slice(sortedData, func(i, j int) bool {
		return sortedData[i] > sortedData[j]
	})

	// Determine the threshold.
	threshold := sortedData[k-1]

	// Apply sparsification.
	sparsifiedData := make([]float32, len(data))
	for i, v := range data {
		if v >= threshold {
			sparsifiedData[i] = v
		}
	}

	return sparsifiedData, nil
}

// NetworkConditions represents the current network conditions.
type NetworkConditions struct {
	E2Latency float64
	SliceID   string
}

// CompressionLevel represents the level of compression to apply.
type CompressionLevel string

const (
	CompressionLevelLow    CompressionLevel = "low"
	CompressionLevelMedium CompressionLevel = "medium"
	CompressionLevelHigh   CompressionLevel = "high"
)
