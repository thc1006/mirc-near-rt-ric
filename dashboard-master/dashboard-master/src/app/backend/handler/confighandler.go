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
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"text/template"
	"time"
)

// AppHandler is an application handler.
type AppHandler func(http.ResponseWriter, *http.Request) (int, error)

// AppConfig is a global configuration of application.
type AppConfig struct {
	// ServerTime is current server time.
	ServerTime int64 `json:"serverTime"`
}

const (
	// ConfigTemplateName is a name of config template
	ConfigTemplateName = "appConfig"
	// ConfigTemplate is a template of a config
	ConfigTemplate = "{{.}}"
)

var (
	// configTemplate is pre-compiled template for better performance
	configTemplate *template.Template
	configOnce     sync.Once
	// configCache caches the JSON string with TTL
	configCache     string
	configCacheTime time.Time
	configCacheMutex sync.RWMutex
	configCacheTTL   = 1 * time.Second // Cache for 1 second to reduce CPU load
)

// ServeHTTP serves HTTP endpoint with application configuration.
func (fn AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := fn(w, r); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
	}
}

// initConfigTemplate initializes the template once
func initConfigTemplate() {
	configOnce.Do(func() {
		var err error
		configTemplate, err = template.New(ConfigTemplateName).Parse(ConfigTemplate)
		if err != nil {
			log.Printf("Error parsing config template: %v", err)
		}
	})
}

func getAppConfigJSON() string {
	// Check cache first
	configCacheMutex.RLock()
	if configCache != "" && time.Since(configCacheTime) < configCacheTTL {
		configCacheMutex.RUnlock()
		return configCache
	}
	configCacheMutex.RUnlock()

	// Cache miss or expired, generate new config
	configCacheMutex.Lock()
	defer configCacheMutex.Unlock()

	// Double-check after acquiring write lock
	if configCache != "" && time.Since(configCacheTime) < configCacheTTL {
		return configCache
	}

	// Log only when generating new config to reduce log spam
	config := &AppConfig{
		ServerTime: time.Now().UTC().UnixNano() / 1e6,
	}

	jsonConfig, err := json.Marshal(config)
	if err != nil {
		log.Printf("Error marshaling config: %v", err)
		return "{}"
	}

	// Update cache
	configCache = string(jsonConfig)
	configCacheTime = time.Now()

	return configCache
}

func ConfigHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	// Initialize template once
	initConfigTemplate()
	
	w.Header().Set("Content-Type", "application/json")
	
	if configTemplate == nil {
		return http.StatusInternalServerError, nil
	}
	
	// Use cached config JSON
	configJSON := getAppConfigJSON()
	return http.StatusOK, configTemplate.Execute(w, configJSON)
}
