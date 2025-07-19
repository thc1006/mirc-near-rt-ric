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

// Package auth provides modernized authentication management
package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"k8s.io/client-go/tools/clientcmd/api"

	authApi "github.com/kubernetes/dashboard/src/app/backend/auth/api"
	clientapi "github.com/kubernetes/dashboard/src/app/backend/client/api"
	"github.com/kubernetes/dashboard/src/app/backend/errors"
)

// ModernAuthManager implements the AuthManager interface with modern Go patterns
type ModernAuthManager struct {
	tokenManager            authApi.TokenManager
	clientManager           clientapi.ClientManager
	authenticationModes     authApi.AuthenticationModes
	authenticationSkippable bool
	logger                  *slog.Logger
	
	// Modern additions
	healthCheckTimeout time.Duration
	metrics           *AuthMetrics
}

// AuthMetrics tracks authentication-related metrics
type AuthMetrics struct {
	LoginAttempts     int64
	SuccessfulLogins  int64
	FailedLogins      int64
	TokenRefreshes    int64
	HealthChecks      int64
}

// AuthOptions configures the auth manager
type AuthOptions struct {
	TokenManager            authApi.TokenManager
	ClientManager           clientapi.ClientManager
	AuthenticationModes     authApi.AuthenticationModes
	AuthenticationSkippable bool
	Logger                  *slog.Logger
	HealthCheckTimeout      time.Duration
}

// NewModernAuthManager creates a new modernized auth manager
func NewModernAuthManager(opts AuthOptions) authApi.AuthManager {
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	
	if opts.HealthCheckTimeout == 0 {
		opts.HealthCheckTimeout = 10 * time.Second
	}
	
	return &ModernAuthManager{
		tokenManager:            opts.TokenManager,
		clientManager:           opts.ClientManager,
		authenticationModes:     opts.AuthenticationModes,
		authenticationSkippable: opts.AuthenticationSkippable,
		logger:                  opts.Logger.With("component", "auth_manager"),
		healthCheckTimeout:      opts.HealthCheckTimeout,
		metrics:                 &AuthMetrics{},
	}
}

// Login implements auth manager with modern error handling and logging
func (am *ModernAuthManager) Login(spec *authApi.LoginSpec) (*authApi.AuthResponse, error) {
	ctx := context.Background()
	startTime := time.Now()
	am.metrics.LoginAttempts++
	
	am.logger.Info("login attempt started",
		"has_token", len(spec.Token) > 0,
		"has_credentials", len(spec.Username) > 0 && len(spec.Password) > 0,
		"has_kubeconfig", len(spec.KubeConfig) > 0)
	
	// Get authenticator with context
	authenticator, err := am.getAuthenticator(spec)
	if err != nil {
		am.metrics.FailedLogins++
		am.logger.Error("failed to create authenticator",
			"error", err,
			"duration", time.Since(startTime))
		return nil, fmt.Errorf("authenticator creation failed: %w", err)
	}
	
	// Get auth info
	authInfo, err := authenticator.GetAuthInfo()
	if err != nil {
		am.metrics.FailedLogins++
		am.logger.Error("failed to get auth info",
			"error", err,
			"authenticator_type", fmt.Sprintf("%T", authenticator),
			"duration", time.Since(startTime))
		return nil, fmt.Errorf("auth info extraction failed: %w", err)
	}
	
	// Perform health check with timeout
	ctx, cancel := context.WithTimeout(ctx, am.healthCheckTimeout)
	defer cancel()
	
	username, err := am.healthCheckWithContext(ctx, authInfo)
	if err != nil {
		am.metrics.FailedLogins++
		
		// Use modern error handling
		nonCriticalErrors, criticalError := errors.HandleErrors([]error{err})
		if criticalError != nil {
			am.logger.Error("critical error during health check",
				"error", criticalError,
				"duration", time.Since(startTime))
			return &authApi.AuthResponse{Errors: nonCriticalErrors}, criticalError
		}
		
		if len(nonCriticalErrors) > 0 {
			am.logger.Warn("non-critical errors during health check",
				"errors", nonCriticalErrors,
				"duration", time.Since(startTime))
			return &authApi.AuthResponse{Errors: nonCriticalErrors}, nil
		}
	}
	
	// Generate token
	token, err := am.tokenManager.Generate(authInfo)
	if err != nil {
		am.metrics.FailedLogins++
		am.logger.Error("failed to generate token",
			"error", err,
			"username", username,
			"duration", time.Since(startTime))
		return nil, fmt.Errorf("token generation failed: %w", err)
	}
	
	am.metrics.SuccessfulLogins++
	am.logger.Info("login successful",
		"username", username,
		"duration", time.Since(startTime),
		"token_length", len(token))
	
	return &authApi.AuthResponse{
		JWEToken: token,
		Name:     username,
		Errors:   nil,
	}, nil
}

// Refresh implements auth manager with modern error handling
func (am *ModernAuthManager) Refresh(jweToken string) (string, error) {
	startTime := time.Now()
	am.metrics.TokenRefreshes++
	
	am.logger.Debug("token refresh started",
		"token_length", len(jweToken))
	
	newToken, err := am.tokenManager.Refresh(jweToken)
	if err != nil {
		am.logger.Error("token refresh failed",
			"error", err,
			"duration", time.Since(startTime))
		
		// Check for specific error types
		if errors.Is(err, errors.ErrTokenExpired) {
			return "", errors.NewTokenExpired("refresh failed: token has expired")
		}
		
		if errors.Is(err, errors.ErrTokenInvalid) {
			return "", errors.NewErrorBuilder(errors.ErrTokenInvalid).
				WithStatus(401).
				WithDetail("operation", "refresh").
				Build()
		}
		
		return "", fmt.Errorf("token refresh failed: %w", err)
	}
	
	am.logger.Debug("token refresh successful",
		"duration", time.Since(startTime),
		"new_token_length", len(newToken))
	
	return newToken, nil
}

// AuthenticationModes returns the available authentication modes
func (am *ModernAuthManager) AuthenticationModes() []authApi.AuthenticationMode {
	return am.authenticationModes.Array()
}

// AuthenticationSkippable returns whether authentication can be skipped
func (am *ModernAuthManager) AuthenticationSkippable() bool {
	return am.authenticationSkippable
}

// GetMetrics returns authentication metrics
func (am *ModernAuthManager) GetMetrics() AuthMetrics {
	return *am.metrics
}

// getAuthenticator returns an authenticator based on the login spec
func (am *ModernAuthManager) getAuthenticator(spec *authApi.LoginSpec) (authApi.Authenticator, error) {
	if len(am.authenticationModes) == 0 {
		return nil, fmt.Errorf("%w: check --authentication-modes argument for more information",
			errors.ErrAllAuthDisabled)
	}
	
	// Check for token authentication
	if len(spec.Token) > 0 && am.authenticationModes.IsEnabled(authApi.Token) {
		am.logger.Debug("using token authentication")
		return NewTokenAuthenticator(spec), nil
	}
	
	// Check for basic authentication
	if len(spec.Username) > 0 && len(spec.Password) > 0 && am.authenticationModes.IsEnabled(authApi.Basic) {
		am.logger.Debug("using basic authentication",
			"username", spec.Username)
		return NewBasicAuthenticator(spec), nil
	}
	
	// Check for kubeconfig authentication
	if len(spec.KubeConfig) > 0 {
		am.logger.Debug("using kubeconfig authentication")
		return NewKubeConfigAuthenticator(spec, am.authenticationModes), nil
	}
	
	am.logger.Warn("insufficient authentication data provided",
		"has_token", len(spec.Token) > 0,
		"has_credentials", len(spec.Username) > 0 && len(spec.Password) > 0,
		"has_kubeconfig", len(spec.KubeConfig) > 0,
		"enabled_modes", am.authenticationModes.Array())
	
	return nil, fmt.Errorf("%w: missing token, credentials, or kubeconfig", errors.ErrAuthDataInsufficient)
}

// healthCheckWithContext performs a health check with context support
func (am *ModernAuthManager) healthCheckWithContext(ctx context.Context, authInfo api.AuthInfo) (string, error) {
	am.metrics.HealthChecks++
	
	// Create a channel to receive the result
	resultChan := make(chan struct {
		username string
		err      error
	}, 1)
	
	// Run health check in a goroutine
	go func() {
		username, err := am.clientManager.HasAccess(authInfo)
		resultChan <- struct {
			username string
			err      error
		}{username, err}
	}()
	
	// Wait for result or context cancellation
	select {
	case result := <-resultChan:
		if result.err != nil {
			return "", fmt.Errorf("health check failed: %w", result.err)
		}
		return result.username, nil
		
	case <-ctx.Done():
		return "", fmt.Errorf("health check timeout: %w", ctx.Err())
	}
}

// ValidateAuthenticationModes validates the configured authentication modes
func (am *ModernAuthManager) ValidateAuthenticationModes() error {
	if len(am.authenticationModes) == 0 {
		return fmt.Errorf("%w: no authentication modes configured", errors.ErrInvalidAuthMode)
	}
	
	validModes := []authApi.AuthenticationMode{authApi.Token, authApi.Basic}
	for _, mode := range am.authenticationModes.Array() {
		found := false
		for _, validMode := range validModes {
			if mode == validMode {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: unsupported authentication mode: %s", errors.ErrInvalidAuthMode, mode)
		}
	}
	
	return nil
}

// LogMetrics logs current authentication metrics
func (am *ModernAuthManager) LogMetrics() {
	am.logger.Info("authentication metrics",
		"login_attempts", am.metrics.LoginAttempts,
		"successful_logins", am.metrics.SuccessfulLogins,
		"failed_logins", am.metrics.FailedLogins,
		"token_refreshes", am.metrics.TokenRefreshes,
		"health_checks", am.metrics.HealthChecks,
		"success_rate", am.calculateSuccessRate())
}

// calculateSuccessRate calculates the login success rate
func (am *ModernAuthManager) calculateSuccessRate() float64 {
	if am.metrics.LoginAttempts == 0 {
		return 0.0
	}
	return float64(am.metrics.SuccessfulLogins) / float64(am.metrics.LoginAttempts) * 100
}

// Ensure ModernAuthManager implements the AuthManager interface
var _ authApi.AuthManager = (*ModernAuthManager)(nil)