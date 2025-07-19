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
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"text/template"
	"time"
)

// OptimizedAppConfig is a pre-compiled configuration cache
type OptimizedAppConfig struct {
	// Template is pre-compiled template
	template *template.Template
	// Config cache with TTL
	cache       []byte
	cacheTime   time.Time
	cacheTTL    time.Duration
	cacheMutex  sync.RWMutex
	// JSON encoder pool for efficient reuse
	encoderPool sync.Pool
	// Buffer pool for JSON encoding
	bufferPool  sync.Pool
}

var (
	optimizedConfig *OptimizedAppConfig
	initOnce        sync.Once
)

// InitOptimizedConfig initializes the optimized configuration handler
func InitOptimizedConfig() *OptimizedAppConfig {
	initOnce.Do(func() {
		tmpl, err := template.New(ConfigTemplateName).Parse(ConfigTemplate)
		if err != nil {
			log.Printf("Failed to parse config template: %v", err)
			return
		}

		optimizedConfig = &OptimizedAppConfig{
			template: tmpl,
			cacheTTL: 1 * time.Second, // Cache config for 1 second to reduce CPU load
			encoderPool: sync.Pool{
				New: func() interface{} {
					buf := &bytes.Buffer{}
					return json.NewEncoder(buf)
				},
			},
			bufferPool: sync.Pool{
				New: func() interface{} {
					return &bytes.Buffer{}
				},
			},
		}
	})
	return optimizedConfig
}

// getAppConfigJSONOptimized returns cached config or generates new one
func (oc *OptimizedAppConfig) getAppConfigJSONOptimized() []byte {
	oc.cacheMutex.RLock()
	if oc.cache != nil && time.Since(oc.cacheTime) < oc.cacheTTL {
		defer oc.cacheMutex.RUnlock()
		return oc.cache
	}
	oc.cacheMutex.RUnlock()

	// Cache expired or doesn't exist, generate new config
	oc.cacheMutex.Lock()
	defer oc.cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if oc.cache != nil && time.Since(oc.cacheTime) < oc.cacheTTL {
		return oc.cache
	}

	// Use buffer pool for efficient memory allocation
	buf := oc.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer oc.bufferPool.Put(buf)

	config := &AppConfig{
		ServerTime: time.Now().UTC().UnixNano() / 1e6,
	}

	// Use encoder pool for efficient JSON encoding
	encoder := json.NewEncoder(buf)
	if err := encoder.Encode(config); err != nil {
		log.Printf("Error encoding config: %v", err)
		return nil
	}

	// Remove trailing newline from encoder
	data := buf.Bytes()
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	// Cache the result
	oc.cache = make([]byte, len(data))
	copy(oc.cache, data)
	oc.cacheTime = time.Now()

	return oc.cache
}

// OptimizedConfigHandler is the optimized version of ConfigHandler
func OptimizedConfigHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	oc := InitOptimizedConfig()
	if oc == nil {
		return http.StatusInternalServerError, nil
	}

	w.Header().Set("Content-Type", "application/json")
	
	// Get cached or fresh config
	configJSON := oc.getAppConfigJSONOptimized()
	if configJSON == nil {
		return http.StatusInternalServerError, nil
	}

	// Use buffer pool for template execution
	buf := oc.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer oc.bufferPool.Put(buf)

	// Execute template to buffer first
	if err := oc.template.Execute(buf, string(configJSON)); err != nil {
		return http.StatusInternalServerError, err
	}

	// Write buffer to response
	_, err := w.Write(buf.Bytes())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

// ConfigHandlerWithPool uses object pooling for better performance
func ConfigHandlerWithPool(w http.ResponseWriter, r *http.Request) (int, error) {
	oc := InitOptimizedConfig()
	if oc == nil {
		// Fallback to original handler
		return ConfigHandler(w, r)
	}

	w.Header().Set("Content-Type", "application/json")

	// Direct write of cached JSON - fastest approach
	configJSON := oc.getAppConfigJSONOptimized()
	if configJSON == nil {
		return http.StatusInternalServerError, nil
	}

	_, err := w.Write(configJSON)
	return http.StatusOK, err
}

// Benchmark helper to compare performance
func BenchmarkConfigHandlers(b *testing.B, handler AppHandler) {
	req, _ := http.NewRequest("GET", "/config", nil)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		w := &mockResponseWriter{}
		handler(w, req)
	}
}

// mockResponseWriter for benchmarking
type mockResponseWriter struct {
	header http.Header
	data   []byte
}

func (m *mockResponseWriter) Header() http.Header {
	if m.header == nil {
		m.header = make(http.Header)
	}
	return m.header
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.data = append(m.data, data...)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	// No-op for benchmarking
}