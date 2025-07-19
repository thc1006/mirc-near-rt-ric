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

// Package errors provides modernized error handling for the dashboard backend
package errors

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// Standard error definitions following modern Go practices
var (
	// Authentication and authorization errors
	ErrUnauthorized         = errors.New("unauthorized access")
	ErrTokenExpired         = errors.New("token has expired")
	ErrTokenInvalid         = errors.New("token is invalid")
	ErrTokenMissing         = errors.New("token is missing")
	ErrInsufficientAccess   = errors.New("insufficient access permissions")
	
	// Input validation errors
	ErrInvalidInput         = errors.New("invalid input")
	ErrMissingRequired      = errors.New("missing required field")
	ErrInvalidFormat        = errors.New("invalid format")
	ErrValueOutOfRange      = errors.New("value out of range")
	
	// Resource errors
	ErrResourceNotFound     = errors.New("resource not found")
	ErrResourceExists       = errors.New("resource already exists")
	ErrResourceConflict     = errors.New("resource conflict")
	ErrResourceLocked       = errors.New("resource is locked")
	
	// System errors
	ErrInternalServer       = errors.New("internal server error")
	ErrServiceUnavailable   = errors.New("service unavailable")
	ErrTimeout              = errors.New("operation timeout")
	ErrClientConfig         = errors.New("client configuration error")
	
	// Authentication mode errors
	ErrAllAuthDisabled      = errors.New("all authentication options disabled")
	ErrInvalidAuthMode      = errors.New("invalid authentication mode")
	ErrAuthDataInsufficient = errors.New("insufficient authentication data")
)

// HTTPError wraps an error with HTTP status information
type HTTPError struct {
	err        error
	statusCode int
	details    map[string]interface{}
	source     string
}

// Error implements the error interface
func (e *HTTPError) Error() string {
	if e.source != "" {
		return fmt.Sprintf("[%s] %s", e.source, e.err.Error())
	}
	return e.err.Error()
}

// Unwrap returns the underlying error
func (e *HTTPError) Unwrap() error {
	return e.err
}

// StatusCode returns the HTTP status code
func (e *HTTPError) StatusCode() int {
	return e.statusCode
}

// Details returns additional error details
func (e *HTTPError) Details() map[string]interface{} {
	return e.details
}

// Source returns the error source
func (e *HTTPError) Source() string {
	return e.source
}

// WithDetails adds details to the error
func (e *HTTPError) WithDetails(key string, value interface{}) *HTTPError {
	if e.details == nil {
		e.details = make(map[string]interface{})
	}
	e.details[key] = value
	return e
}

// ErrorBuilder provides a fluent API for building errors
type ErrorBuilder struct {
	baseErr    error
	statusCode int
	details    map[string]interface{}
	source     string
	wrapped    []error
}

// NewErrorBuilder creates a new error builder
func NewErrorBuilder(baseErr error) *ErrorBuilder {
	return &ErrorBuilder{
		baseErr:    baseErr,
		details:    make(map[string]interface{}),
		statusCode: http.StatusInternalServerError,
	}
}

// WithStatus sets the HTTP status code
func (eb *ErrorBuilder) WithStatus(code int) *ErrorBuilder {
	eb.statusCode = code
	return eb
}

// WithDetail adds a detail to the error
func (eb *ErrorBuilder) WithDetail(key string, value interface{}) *ErrorBuilder {
	eb.details[key] = value
	return eb
}

// WithSource sets the error source
func (eb *ErrorBuilder) WithSource(source string) *ErrorBuilder {
	eb.source = source
	return eb
}

// WithCaller automatically sets the source from the calling function
func (eb *ErrorBuilder) WithCaller() *ErrorBuilder {
	if pc, file, line, ok := runtime.Caller(1); ok {
		if fn := runtime.FuncForPC(pc); fn != nil {
			name := fn.Name()
			if idx := strings.LastIndex(name, "."); idx != -1 {
				name = name[idx+1:]
			}
			eb.source = fmt.Sprintf("%s:%d:%s", file[strings.LastIndex(file, "/")+1:], line, name)
		}
	}
	return eb
}

// Wrap wraps another error
func (eb *ErrorBuilder) Wrap(err error) *ErrorBuilder {
	eb.wrapped = append(eb.wrapped, err)
	return eb
}

// Build creates the final HTTPError
func (eb *ErrorBuilder) Build() *HTTPError {
	finalErr := eb.baseErr
	
	// Wrap any additional errors
	for _, wrapped := range eb.wrapped {
		finalErr = fmt.Errorf("%w: %v", finalErr, wrapped)
	}
	
	return &HTTPError{
		err:        finalErr,
		statusCode: eb.statusCode,
		details:    eb.details,
		source:     eb.source,
	}
}

// Modern error constructors using the builder pattern

// NewUnauthorized creates an unauthorized error
func NewUnauthorized(message string) error {
	return NewErrorBuilder(ErrUnauthorized).
		WithStatus(http.StatusUnauthorized).
		WithDetail("message", message).
		WithCaller().
		Build()
}

// NewInvalid creates an invalid input error
func NewInvalid(message string) error {
	return NewErrorBuilder(ErrInvalidInput).
		WithStatus(http.StatusBadRequest).
		WithDetail("message", message).
		WithCaller().
		Build()
}

// NewNotFound creates a not found error
func NewNotFound(resource, name string) error {
	return NewErrorBuilder(ErrResourceNotFound).
		WithStatus(http.StatusNotFound).
		WithDetail("resource", resource).
		WithDetail("name", name).
		WithCaller().
		Build()
}

// NewTokenExpired creates a token expired error
func NewTokenExpired(message string) error {
	return NewErrorBuilder(ErrTokenExpired).
		WithStatus(http.StatusUnauthorized).
		WithDetail("message", message).
		WithDetail("expired", true).
		WithCaller().
		Build()
}

// NewInternal creates an internal server error
func NewInternal(message string) error {
	return NewErrorBuilder(ErrInternalServer).
		WithStatus(http.StatusInternalServerError).
		WithDetail("message", message).
		WithCaller().
		Build()
}

// NewTimeout creates a timeout error
func NewTimeout(operation string, duration string) error {
	return NewErrorBuilder(ErrTimeout).
		WithStatus(http.StatusRequestTimeout).
		WithDetail("operation", operation).
		WithDetail("duration", duration).
		WithCaller().
		Build()
}

// Error handling utilities

// HandleError processes an error and returns appropriate HTTP status and message
func HandleError(err error) (statusCode int, message string, details map[string]interface{}) {
	if err == nil {
		return http.StatusOK, "", nil
	}

	// Check for HTTPError
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode(), httpErr.Error(), httpErr.Details()
	}

	// Check for Kubernetes API errors
	if k8sErr := k8serrors.APIStatus(nil); errors.As(err, &k8sErr) {
		status := k8sErr.Status()
		return int(status.Code), status.Message, map[string]interface{}{
			"reason": status.Reason,
			"kind":   status.Details.Kind if status.Details != nil else "",
		}
	}

	// Check for standard errors
	switch {
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, err.Error(), nil
	case errors.Is(err, ErrTokenExpired):
		return http.StatusUnauthorized, err.Error(), map[string]interface{}{"expired": true}
	case errors.Is(err, ErrInvalidInput):
		return http.StatusBadRequest, err.Error(), nil
	case errors.Is(err, ErrResourceNotFound):
		return http.StatusNotFound, err.Error(), nil
	case errors.Is(err, ErrTimeout):
		return http.StatusRequestTimeout, err.Error(), nil
	case errors.Is(err, ErrServiceUnavailable):
		return http.StatusServiceUnavailable, err.Error(), nil
	default:
		return http.StatusInternalServerError, "Internal server error", nil
	}
}

// HandleErrors processes multiple errors and returns critical and non-critical errors
func HandleErrors(errs []error) (nonCriticalErrors []error, criticalError error) {
	for _, err := range errs {
		if err == nil {
			continue
		}

		// Determine if error is critical based on HTTP status
		statusCode, _, _ := HandleError(err)
		if statusCode >= 500 {
			criticalError = err
			break
		} else {
			nonCriticalErrors = append(nonCriticalErrors, err)
		}
	}

	return nonCriticalErrors, criticalError
}

// LogError logs an error with appropriate context
func LogError(logger *slog.Logger, err error, context string, attrs ...slog.Attr) {
	if err == nil {
		return
	}

	statusCode, message, details := HandleError(err)
	
	logAttrs := []slog.Attr{
		slog.String("context", context),
		slog.Int("status_code", statusCode),
		slog.String("error", message),
	}
	
	// Add details as attributes
	for key, value := range details {
		logAttrs = append(logAttrs, slog.Any(key, value))
	}
	
	// Add any additional attributes
	logAttrs = append(logAttrs, attrs...)
	
	// Log at appropriate level based on severity
	if statusCode >= 500 {
		logger.Error("critical error occurred", logAttrs...)
	} else if statusCode >= 400 {
		logger.Warn("client error occurred", logAttrs...)
	} else {
		logger.Info("operation completed with non-critical error", logAttrs...)
	}
}

// WrapKubernetesError wraps a Kubernetes API error with additional context
func WrapKubernetesError(err error, operation, resource string) error {
	if err == nil {
		return nil
	}

	if k8serrors.IsNotFound(err) {
		return NewErrorBuilder(ErrResourceNotFound).
			WithStatus(http.StatusNotFound).
			WithDetail("operation", operation).
			WithDetail("resource", resource).
			WithCaller().
			Wrap(err).
			Build()
	}

	if k8serrors.IsUnauthorized(err) {
		return NewErrorBuilder(ErrUnauthorized).
			WithStatus(http.StatusUnauthorized).
			WithDetail("operation", operation).
			WithDetail("resource", resource).
			WithCaller().
			Wrap(err).
			Build()
	}

	if k8serrors.IsForbidden(err) {
		return NewErrorBuilder(ErrInsufficientAccess).
			WithStatus(http.StatusForbidden).
			WithDetail("operation", operation).
			WithDetail("resource", resource).
			WithCaller().
			Wrap(err).
			Build()
	}

	if k8serrors.IsTimeout(err) {
		return NewErrorBuilder(ErrTimeout).
			WithStatus(http.StatusRequestTimeout).
			WithDetail("operation", operation).
			WithDetail("resource", resource).
			WithCaller().
			Wrap(err).
			Build()
	}

	// For other Kubernetes errors, wrap as internal error
	return NewErrorBuilder(ErrInternalServer).
		WithStatus(http.StatusInternalServerError).
		WithDetail("operation", operation).
		WithDetail("resource", resource).
		WithCaller().
		Wrap(err).
		Build()
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", ve.Field, ve.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ves ValidationErrors) Error() string {
	if len(ves) == 0 {
		return "validation failed"
	}
	
	var messages []string
	for _, ve := range ves {
		messages = append(messages, ve.Error())
	}
	
	return fmt.Sprintf("validation failed: %s", strings.Join(messages, "; "))
}

// NewValidationError creates a new validation error
func NewValidationError(field, value, message string) error {
	return NewErrorBuilder(ErrInvalidInput).
		WithStatus(http.StatusBadRequest).
		WithDetail("validation_errors", ValidationErrors{
			{Field: field, Value: value, Message: message},
		}).
		WithCaller().
		Build()
}

// NewValidationErrors creates an error from multiple validation errors
func NewValidationErrors(errors []ValidationError) error {
	return NewErrorBuilder(ErrInvalidInput).
		WithStatus(http.StatusBadRequest).
		WithDetail("validation_errors", ValidationErrors(errors)).
		WithCaller().
		Build()
}