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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/kubernetes/dashboard/src/app/backend/args"
	"github.com/kubernetes/dashboard/src/app/backend/auth"
	"github.com/kubernetes/dashboard/src/app/backend/client"
	"github.com/kubernetes/dashboard/src/app/backend/handler"
	"github.com/kubernetes/dashboard/src/app/backend/integration"
	"github.com/kubernetes/dashboard/src/app/backend/settings"
	"github.com/kubernetes/dashboard/src/app/backend/systembanner"
)

// TestPayload represents a test request payload
type TestPayload struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
	Data      map[string]string `json:"data"`
}

func BenchmarkHTTPRequests(b *testing.B) {
	// Setup test environment
	initArgHolderForTest()
	
	// Create mock managers
	clientManager := client.NewClientManager("", "")
	authManager := auth.NewAuthManager(clientManager, nil, nil, false)
	settingsManager := settings.NewSettingsManager()
	systemBannerManager := systembanner.NewSystemBannerManager("", "")
	integrationManager := integration.NewIntegrationManager(clientManager)
	
	// Create API handler
	apiHandler, err := handler.CreateHTTPAPIHandler(
		integrationManager,
		clientManager,
		authManager,
		settingsManager,
		systemBannerManager,
	)
	if err != nil {
		b.Fatalf("Failed to create API handler: %v", err)
	}
	
	// Create test server
	server := httptest.NewServer(apiHandler)
	defer server.Close()
	
	// Test payload
	payload := TestPayload{
		Name:      "test-resource",
		Namespace: "default",
		Labels:    map[string]string{"app": "test", "version": "1.0"},
		Data:      map[string]string{"key1": "value1", "key2": "value2"},
	}
	
	jsonPayload, _ := json.Marshal(payload)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	// Benchmark with various concurrency levels
	concurrencyLevels := []int{1, 10, 50, 100, 500, 1000}
	
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency-%d", concurrency), func(b *testing.B) {
			benchmarkConcurrentRequests(b, server.URL, jsonPayload, concurrency)
		})
	}
}

func benchmarkConcurrentRequests(b *testing.B, serverURL string, payload []byte, concurrency int) {
	// Channel to control concurrency
	sem := make(chan struct{}, concurrency)
	
	// Track memory usage
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	start := time.Now()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sem <- struct{}{} // Acquire semaphore
			
			go func() {
				defer func() { <-sem }() // Release semaphore
				
				// Make HTTP request
				req, err := http.NewRequest("GET", serverURL+"/api/v1/namespace", bytes.NewReader(payload))
				if err != nil {
					b.Error(err)
					return
				}
				req.Header.Set("Content-Type", "application/json")
				
				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					b.Error(err)
					return
				}
				defer resp.Body.Close()
				
				// Read response to simulate processing
				_, err = io.ReadAll(resp.Body)
				if err != nil {
					b.Error(err)
					return
				}
			}()
		}
	})
	
	// Wait for all goroutines to complete
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}
	
	elapsed := time.Since(start)
	runtime.GC()
	runtime.ReadMemStats(&m2)
	
	// Calculate metrics
	requestsPerSecond := float64(b.N) / elapsed.Seconds()
	memoryUsed := m2.Alloc - m1.Alloc
	
	b.ReportMetric(requestsPerSecond, "req/s")
	b.ReportMetric(float64(memoryUsed), "bytes/op")
	b.ReportMetric(float64(memoryUsed)/float64(b.N), "bytes/req")
}

func BenchmarkJSONSerialization(b *testing.B) {
	// Test different payload sizes
	payloadSizes := []int{100, 1000, 10000, 100000}
	
	for _, size := range payloadSizes {
		b.Run(fmt.Sprintf("PayloadSize-%d", size), func(b *testing.B) {
			benchmarkJSONOperations(b, size)
		})
	}
}

func benchmarkJSONOperations(b *testing.B, payloadSize int) {
	// Create test data
	data := make(map[string]interface{})
	for i := 0; i < payloadSize; i++ {
		data[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.Run("Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(data)
			if err != nil {
				b.Error(err)
			}
		}
	})
	
	// Marshal once for unmarshal test
	jsonData, _ := json.Marshal(data)
	
	b.Run("Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result map[string]interface{}
			err := json.Unmarshal(jsonData, &result)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

func BenchmarkConfigHandling(b *testing.B) {
	initArgHolderForTest()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.Run("ConfigGeneration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate config generation (similar to confighandler.go)
			config := map[string]interface{}{
				"serverTime": time.Now().Unix(),
			}
			
			_, err := json.Marshal(config)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

func initArgHolderForTest() {
	builder := args.GetHolderBuilder()
	builder.SetInsecurePort(9090)
	builder.SetPort(8443)
	builder.SetTokenTTL(900)
	builder.SetMetricClientCheckPeriod(30)
	builder.SetInsecureBindAddress(net.IPv4(127, 0, 0, 1))
	builder.SetBindAddress(net.IPv4(0, 0, 0, 0))
	builder.SetDefaultCertDir("/certs")
	builder.SetCertFile("")
	builder.SetKeyFile("")
	builder.SetApiServerHost("")
	builder.SetMetricsProvider("none")
	builder.SetHeapsterHost("")
	builder.SetSidecarHost("")
	builder.SetKubeConfigFile("")
	builder.SetSystemBanner("")
	builder.SetSystemBannerSeverity("INFO")
	builder.SetAPILogLevel("INFO")
	builder.SetAuthenticationMode([]string{"token"})
	builder.SetAutoGenerateCertificates(false)
	builder.SetEnableInsecureLogin(false)
	builder.SetDisableSettingsAuthorizer(false)
	builder.SetEnableSkipLogin(false)
	builder.SetNamespace("kube-system")
	builder.SetLocaleConfig("./locale_conf.json")
}

// CPU Profiling benchmark
func BenchmarkCPUIntensive(b *testing.B) {
	initArgHolderForTest()
	
	// Create large payload
	largePayload := make(map[string]interface{})
	for i := 0; i < 10000; i++ {
		largePayload[fmt.Sprintf("key%d", i)] = map[string]interface{}{
			"nested1": fmt.Sprintf("value%d", i),
			"nested2": []string{fmt.Sprintf("item%d", i), fmt.Sprintf("item%d", i+1)},
			"nested3": map[string]string{
				"subkey1": fmt.Sprintf("subvalue%d", i),
				"subkey2": fmt.Sprintf("subvalue%d", i+1),
			},
		}
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Marshal
		data, err := json.Marshal(largePayload)
		if err != nil {
			b.Error(err)
		}
		
		// Unmarshal
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		if err != nil {
			b.Error(err)
		}
	}
}