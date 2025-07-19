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
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// BenchmarkConfigHandlerOriginal benchmarks the original config handler
func BenchmarkConfigHandlerOriginal(b *testing.B) {
	req, _ := http.NewRequest("GET", "/config", nil)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ConfigHandler(w, req)
	}
}

// BenchmarkConfigHandlerOptimized benchmarks the optimized config handler
func BenchmarkConfigHandlerOptimized(b *testing.B) {
	req, _ := http.NewRequest("GET", "/config", nil)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		OptimizedConfigHandler(w, req)
	}
}

// BenchmarkConfigHandlerWithPool benchmarks the pooled config handler
func BenchmarkConfigHandlerWithPool(b *testing.B) {
	req, _ := http.NewRequest("GET", "/config", nil)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		ConfigHandlerWithPool(w, req)
	}
}

// BenchmarkConcurrentConfigHandler tests concurrent access
func BenchmarkConcurrentConfigHandler(b *testing.B) {
	req, _ := http.NewRequest("GET", "/config", nil)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			ConfigHandler(w, req)
		}
	})
}

// BenchmarkJSONPoolOperations benchmarks JSON pool operations
func BenchmarkJSONPoolOperations(b *testing.B) {
	pool := NewJSONPool()
	
	testData := map[string]interface{}{
		"serverTime": 1234567890,
		"config": map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
	
	b.Run("PooledEncode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			_, err := pool.EncodeJSON(testData)
			if err != nil {
				b.Error(err)
			}
		}
	})
	
	// Prepare data for decode benchmark
	jsonData, _ := pool.EncodeJSON(testData)
	
	b.Run("PooledDecode", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		
		for i := 0; i < b.N; i++ {
			var result map[string]interface{}
			err := pool.DecodeJSON(bytes.NewReader(jsonData), &result)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkStreamingJSON benchmarks streaming JSON operations
func BenchmarkStreamingJSON(b *testing.B) {
	items := make([]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = map[string]interface{}{
			"id":   i,
			"name": fmt.Sprintf("item-%d", i),
			"data": map[string]string{
				"key1": fmt.Sprintf("value1-%d", i),
				"key2": fmt.Sprintf("value2-%d", i),
			},
		}
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		encoder := NewStreamingJSONEncoder(buf)
		err := encoder.WriteItems(items)
		if err != nil {
			b.Error(err)
		}
	}
}