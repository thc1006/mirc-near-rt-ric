package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// SecurityConfig holds all security-related configuration
type SecurityConfig struct {
	// Authentication & Authorization
	JWTSigningKey        string        `json:"jwt_signing_key"`
	JWTExpiry            time.Duration `json:"jwt_expiry"`
	EnableAuthentication bool          `json:"enable_authentication"`
	EnableAuthorization  bool          `json:"enable_authorization"`

	// TLS Configuration
	TLSEnabled    bool   `json:"tls_enabled"`
	TLSCertPath   string `json:"tls_cert_path"`
	TLSKeyPath    string `json:"tls_key_path"`
	TLSCAPath     string `json:"tls_ca_path"`
	TLSMinVersion string `json:"tls_min_version"`

	// E2 Interface Security
	E2TLSEnabled         bool     `json:"e2_tls_enabled"`
	E2AllowedNodeCertCNs []string `json:"e2_allowed_node_cert_cns"`
	E2RequireClientCert  bool     `json:"e2_require_client_cert"`

	// A1 Interface Security
	A1TLSEnabled        bool     `json:"a1_tls_enabled"`
	A1AllowedClientCNs  []string `json:"a1_allowed_client_cns"`
	A1RequireClientCert bool     `json:"a1_require_client_cert"`

	// API Security
	APIRateLimitEnabled   bool        `json:"api_rate_limit_enabled"`
	APIRateLimitPerMinute int         `json:"api_rate_limit_per_minute"`
	APITimeouts           APITimeouts `json:"api_timeouts"`
	CORSAllowedOrigins    []string    `json:"cors_allowed_origins"`

	// Audit Logging
	AuditLoggingEnabled bool   `json:"audit_logging_enabled"`
	AuditLogPath        string `json:"audit_log_path"`
	AuditLogMaxSize     int    `json:"audit_log_max_size_mb"`
	AuditLogMaxAge      int    `json:"audit_log_max_age_days"`

	// Password Policy
	PasswordPolicy PasswordPolicy `json:"password_policy"`

	// Session Management
	SessionTimeout    time.Duration `json:"session_timeout"`
	MaxActiveSessions int           `json:"max_active_sessions"`

	// Security Headers
	EnableSecurityHeaders bool `json:"enable_security_headers"`
	HSTSMaxAge            int  `json:"hsts_max_age"`
	ContentTypeNoSniff    bool `json:"content_type_no_sniff"`
	FrameDeny             bool `json:"frame_deny"`

	// Input Validation
	MaxRequestSize        int64 `json:"max_request_size"`
	EnableInputValidation bool  `json:"enable_input_validation"`

	// Encryption
	DataEncryptionKey   string `json:"data_encryption_key"`
	EncryptionAlgorithm string `json:"encryption_algorithm"`
}

// APITimeouts defines timeout configurations for various API operations
type APITimeouts struct {
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`
	IdleTimeout    time.Duration `json:"idle_timeout"`
	RequestTimeout time.Duration `json:"request_timeout"`
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength        int  `json:"min_length"`
	RequireUppercase bool `json:"require_uppercase"`
	RequireLowercase bool `json:"require_lowercase"`
	RequireNumbers   bool `json:"require_numbers"`
	RequireSpecial   bool `json:"require_special"`
	MaxAge           int  `json:"max_age_days"`
	HistoryCount     int  `json:"history_count"`
}

// LoadSecurityConfig loads security configuration from environment variables and defaults
func LoadSecurityConfig() (*SecurityConfig, error) {
	config := &SecurityConfig{
		// Authentication & Authorization defaults
		JWTSigningKey:        getEnvOrDefault("JWT_SIGNING_KEY", ""),
		JWTExpiry:            getDurationEnvOrDefault("JWT_EXPIRY", 24*time.Hour),
		EnableAuthentication: getBoolEnvOrDefault("ENABLE_AUTHENTICATION", true),
		EnableAuthorization:  getBoolEnvOrDefault("ENABLE_AUTHORIZATION", true),

		// TLS Configuration defaults
		TLSEnabled:    getBoolEnvOrDefault("TLS_ENABLED", true),
		TLSCertPath:   getEnvOrDefault("TLS_CERT_PATH", "/certs/tls.crt"),
		TLSKeyPath:    getEnvOrDefault("TLS_KEY_PATH", "/certs/tls.key"),
		TLSCAPath:     getEnvOrDefault("TLS_CA_PATH", "/certs/ca.crt"),
		TLSMinVersion: getEnvOrDefault("TLS_MIN_VERSION", "1.3"),

		// E2 Interface Security defaults
		E2TLSEnabled:         getBoolEnvOrDefault("E2_TLS_ENABLED", true),
		E2AllowedNodeCertCNs: getStringSliceEnvOrDefault("E2_ALLOWED_NODE_CERT_CNS", []string{}),
		E2RequireClientCert:  getBoolEnvOrDefault("E2_REQUIRE_CLIENT_CERT", true),

		// A1 Interface Security defaults
		A1TLSEnabled:        getBoolEnvOrDefault("A1_TLS_ENABLED", true),
		A1AllowedClientCNs:  getStringSliceEnvOrDefault("A1_ALLOWED_CLIENT_CNS", []string{}),
		A1RequireClientCert: getBoolEnvOrDefault("A1_REQUIRE_CLIENT_CERT", true),

		// API Security defaults
		APIRateLimitEnabled:   getBoolEnvOrDefault("API_RATE_LIMIT_ENABLED", true),
		APIRateLimitPerMinute: getIntEnvOrDefault("API_RATE_LIMIT_PER_MINUTE", 1000),
		APITimeouts: APITimeouts{
			ReadTimeout:    getDurationEnvOrDefault("API_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:   getDurationEnvOrDefault("API_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:    getDurationEnvOrDefault("API_IDLE_TIMEOUT", 60*time.Second),
			RequestTimeout: getDurationEnvOrDefault("API_REQUEST_TIMEOUT", 30*time.Second),
		},
		CORSAllowedOrigins: getStringSliceEnvOrDefault("CORS_ALLOWED_ORIGINS", []string{"https://localhost:8443"}),

		// Audit Logging defaults
		AuditLoggingEnabled: getBoolEnvOrDefault("AUDIT_LOGGING_ENABLED", true),
		AuditLogPath:        getEnvOrDefault("AUDIT_LOG_PATH", "/var/log/oran-ric/audit.log"),
		AuditLogMaxSize:     getIntEnvOrDefault("AUDIT_LOG_MAX_SIZE_MB", 100),
		AuditLogMaxAge:      getIntEnvOrDefault("AUDIT_LOG_MAX_AGE_DAYS", 30),

		// Password Policy defaults
		PasswordPolicy: PasswordPolicy{
			MinLength:        getIntEnvOrDefault("PASSWORD_MIN_LENGTH", 12),
			RequireUppercase: getBoolEnvOrDefault("PASSWORD_REQUIRE_UPPERCASE", true),
			RequireLowercase: getBoolEnvOrDefault("PASSWORD_REQUIRE_LOWERCASE", true),
			RequireNumbers:   getBoolEnvOrDefault("PASSWORD_REQUIRE_NUMBERS", true),
			RequireSpecial:   getBoolEnvOrDefault("PASSWORD_REQUIRE_SPECIAL", true),
			MaxAge:           getIntEnvOrDefault("PASSWORD_MAX_AGE_DAYS", 90),
			HistoryCount:     getIntEnvOrDefault("PASSWORD_HISTORY_COUNT", 12),
		},

		// Session Management defaults
		SessionTimeout:    getDurationEnvOrDefault("SESSION_TIMEOUT", 30*time.Minute),
		MaxActiveSessions: getIntEnvOrDefault("MAX_ACTIVE_SESSIONS", 10),

		// Security Headers defaults
		EnableSecurityHeaders: getBoolEnvOrDefault("ENABLE_SECURITY_HEADERS", true),
		HSTSMaxAge:            getIntEnvOrDefault("HSTS_MAX_AGE", 31536000), // 1 year
		ContentTypeNoSniff:    getBoolEnvOrDefault("CONTENT_TYPE_NO_SNIFF", true),
		FrameDeny:             getBoolEnvOrDefault("FRAME_DENY", true),

		// Input Validation defaults
		MaxRequestSize:        int64(getIntEnvOrDefault("MAX_REQUEST_SIZE", 10*1024*1024)), // 10MB
		EnableInputValidation: getBoolEnvOrDefault("ENABLE_INPUT_VALIDATION", true),

		// Encryption defaults
		DataEncryptionKey:   getEnvOrDefault("DATA_ENCRYPTION_KEY", ""),
		EncryptionAlgorithm: getEnvOrDefault("ENCRYPTION_ALGORITHM", "AES-256-GCM"),
	}

	// Generate JWT signing key if not provided
	if config.JWTSigningKey == "" {
		key, err := generateRandomKey(64)
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWT signing key: %w", err)
		}
		config.JWTSigningKey = key
	}

	// Generate data encryption key if not provided
	if config.DataEncryptionKey == "" {
		key, err := generateRandomKey(32) // 256-bit key
		if err != nil {
			return nil, fmt.Errorf("failed to generate data encryption key: %w", err)
		}
		config.DataEncryptionKey = key
	}

	return config, nil
}

// ValidateConfig validates the security configuration
func (sc *SecurityConfig) ValidateConfig() error {
	// Validate JWT configuration
	if sc.EnableAuthentication && sc.JWTSigningKey == "" {
		return fmt.Errorf("JWT signing key is required when authentication is enabled")
	}

	// Validate TLS configuration
	if sc.TLSEnabled {
		if sc.TLSCertPath == "" || sc.TLSKeyPath == "" {
			return fmt.Errorf("TLS certificate and key paths are required when TLS is enabled")
		}
	}

	// Validate password policy
	if sc.PasswordPolicy.MinLength < 8 {
		return fmt.Errorf("password minimum length must be at least 8 characters")
	}

	// Validate API timeouts
	if sc.APITimeouts.RequestTimeout <= 0 {
		return fmt.Errorf("API request timeout must be positive")
	}

	// Validate encryption key
	if sc.DataEncryptionKey == "" {
		return fmt.Errorf("data encryption key is required")
	}

	return nil
}

// Helper functions for environment variable parsing

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getDurationEnvOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getStringSliceEnvOrDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// generateRandomKey generates a cryptographically secure random key
func generateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}
