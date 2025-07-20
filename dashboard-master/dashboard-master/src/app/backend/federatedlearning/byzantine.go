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
	"fmt"
	"log/slog"
	"sort"
	"sync"

	"gonum.org/v1/gonum/stat"
)

// ByzantineFTStrategy implements Byzantine fault-tolerant aggregation
type ByzantineFTStrategy struct {
	logger               *slog.Logger
	config               *ByzantineFTConfig
	faultToleranceEngine *ByzantineFaultToleranceEngine
	anomalyDetector      *AnomalyDetector
	reputationSystem     *ReputationSystem
	mutex                sync.RWMutex
}

// ByzantineFTConfig configures the Byzantine fault-tolerant strategy
type ByzantineFTConfig struct {
	Krum         bool    `yaml:"krum"`
	TrimmedMean  bool    `yaml:"trimmed_mean"`
	TrimmedMeanP float64 `yaml:"trimmed_mean_p"`
}

// NewByzantineFTStrategy creates a new ByzantineFTStrategy
func NewByzantineFTStrategy(logger *slog.Logger, config *ByzantineFTConfig) (*ByzantineFTStrategy, error) {
	return &ByzantineFTStrategy{
		logger: logger,
		config: config,
	}, nil
}

// Krum selects the model update that is closest to its k nearest neighbors.
func (s *ByzantineFTStrategy) Krum(updates [][]float64, k int) ([]float64, error) {
	if len(updates) < k+1 {
		return nil, fmt.Errorf("not enough updates for Krum")
	}

	scores := make([]float64, len(updates))
	for i := range updates {
		distances := make([]float64, len(updates))
		for j := range updates {
			distances[j] = euclideanDistance(updates[i], updates[j])
		}
		sort.Float64s(distances)
		scores[i] = sum(distances[:k+1])
	}

	bestIndex := 0
	for i := 1; i < len(scores); i++ {
		if scores[i] < scores[bestIndex] {
			bestIndex = i
		}
	}

	return updates[bestIndex], nil
}

// TrimmedMean removes a certain percentage of the smallest and largest model updates before averaging the rest.
func (s *ByzantineFTStrategy) TrimmedMean(updates [][]float64, p float64) ([]float64, error) {
	if p < 0 || p >= 0.5 {
		return nil, fmt.Errorf("p must be in [0, 0.5)")
	}

	n := len(updates)
	k := int(float64(n) * p)

	if n <= 2*k {
		return nil, fmt.Errorf("not enough updates for Trimmed Mean")
	}

	trimmedUpdates := make([][]float64, n-2*k)
	for i := 0; i < len(updates[0]); i++ {
		values := make([]float64, n)
		for j := 0; j < n; j++ {
			values[j] = updates[j][i]
		}
		sort.Float64s(values)
		trimmedValues := values[k : n-k]
		for j := 0; j < len(trimmedValues); j++ {
			if i == 0 {
				trimmedUpdates[j] = make([]float64, len(updates[0]))
			}
			trimmedUpdates[j][i] = trimmedValues[j]
		}
	}

	return average(trimmedUpdates), nil
}

// AnomalyDetector detects anomalous model updates.
type AnomalyDetector struct {
	logger *slog.Logger
}

// DetectAnomalies detects anomalies in the given model updates.
func (d *AnomalyDetector) DetectAnomalies(updates [][]float64) ([]int, error) {
	var anomalies []int
	// Monitor gradient magnitude deviations
	magnitudes := make([]float64, len(updates))
	for i, update := range updates {
		magnitudes[i] = magnitude(update)
	}
	mean, stdDev := stat.MeanStdDev(magnitudes, nil)
	for i, magnitude := range magnitudes {
		if magnitude > mean+3*stdDev {
			anomalies = append(anomalies, i)
		}
	}
	return anomalies, nil
}

// ReputationSystem tracks the reputation of clients.
type ReputationSystem struct {
	scores map[string]float64
	mutex  sync.RWMutex
}

// UpdateScore updates the reputation score of a client.
func (s *ReputationSystem) UpdateScore(clientID string, score float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.scores[clientID] = score
}

// GetScore retrieves the reputation score of a client.
func (s *ReputationSystem) GetScore(clientID string) float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.scores[clientID]
}

// euclideanDistance calculates the Euclidean distance between two vectors.
func euclideanDistance(a, b []float64) float64 {
	var sum float64
	for i := range a {
		sum += (a[i] - b[i]) * (a[i] - b[i])
	}
	return sum
}

// sum calculates the sum of a slice of floats.
func sum(data []float64) float64 {
	var total float64
	for _, v := range data {
		total += v
	}
	return total
}

// average calculates the average of a slice of vectors.
func average(data [][]float64) []float64 {
	if len(data) == 0 {
		return nil
	}
	avg := make([]float64, len(data[0]))
	for _, v := range data {
		for i := range v {
			avg[i] += v[i]
		}
	}
	for i := range avg {
		avg[i] /= float64(len(data))
	}
	return avg
}

// magnitude calculates the magnitude of a vector.
func magnitude(data []float64) float64 {
	var sum float64
	for _, v := range data {
		sum += v * v
	}
	return sum
}
