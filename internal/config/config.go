package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

// Config represents the complete configuration for the Near-RT RIC
type Config struct {
	// Global settings
	Global   GlobalConfig   `mapstructure:"global"`
	
	// O-RAN interface configurations
	E2       E2Config       `mapstructure:"e2"`
	A1       A1Config       `mapstructure:"a1"`
	O1       O1Config       `mapstructure:"o1"`
	
	// Application framework
	XApp     XAppConfig     `mapstructure:"xapp"`
	
	// Infrastructure
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Security SecurityConfig `mapstructure:"security"`
	
	// Observability
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Tracing  TracingConfig  `mapstructure:"tracing"`
}

// GlobalConfig contains global application settings
type GlobalConfig struct {
	Environment string `mapstructure:"environment" validate:"required,oneof=development staging production"`
	NodeID      string `mapstructure:"node_id" validate:"required"`
	ClusterID   string `mapstructure:"cluster_id" validate:"required"`
	LogLevel    string `mapstructure:"log_level" default:"info"`
}

// E2Config contains E2 interface configuration
type E2Config struct {
	Enabled         bool          `mapstructure:"enabled" default:"true"`
	ListenAddress   string        `mapstructure:"listen_address" default:"0.0.0.0"`
	Port            int           `mapstructure:"port" default:"36421"`
	MaxConnections  int           `mapstructure:"max_connections" default:"1000"`
	
	// SCTP configuration
	SCTP            SCTPConfig    `mapstructure:"sctp"`
	
	// ASN.1 configuration
	ASN1            ASN1Config    `mapstructure:"asn1"`
	
	// Timeout settings
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout" default:"30s"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval" default:"30s"`
	MessageTimeout    time.Duration `mapstructure:"message_timeout" default:"10s"`
	
	// Performance settings
	BufferSize      int           `mapstructure:"buffer_size" default:"65536"`
	WorkerPoolSize  int           `mapstructure:"worker_pool_size" default:"10"`
	
	// Service Models
	ServiceModels   []ServiceModelConfig `mapstructure:"service_models"`
}

// SCTPConfig contains SCTP-specific configuration
type SCTPConfig struct {
	NumStreams        int           `mapstructure:"num_streams" default:"10"`
	MaxRetransmits    int           `mapstructure:"max_retransmits" default:"5"`
	MaxInitAttempts   int           `mapstructure:"max_init_attempts" default:"3"`
	RTOInitial        time.Duration `mapstructure:"rto_initial" default:"3s"`
	RTOMin            time.Duration `mapstructure:"rto_min" default:"1s"`
	RTOMax            time.Duration `mapstructure:"rto_max" default:"60s"`
	SackDelay         time.Duration `mapstructure:"sack_delay" default:"200ms"`
}

// ASN1Config contains ASN.1 encoding configuration
type ASN1Config struct {
	ValidateMessages bool   `mapstructure:"validate_messages" default:"true"`
	StrictDecoding   bool   `mapstructure:"strict_decoding" default:"true"`
	MaxMessageSize   int    `mapstructure:"max_message_size" default:"1048576"` // 1MB
}

// ServiceModelConfig contains configuration for E2 Service Models
type ServiceModelConfig struct {
	ID          string            `mapstructure:"id" validate:"required"`
	Name        string            `mapstructure:"name" validate:"required"`
	Version     string            `mapstructure:"version" validate:"required"`
	OID         string            `mapstructure:"oid" validate:"required"`
	Enabled     bool              `mapstructure:"enabled" default:"true"`
	Config      map[string]interface{} `mapstructure:"config"`
}

// A1Config contains A1 interface configuration
type A1Config struct {
	Enabled         bool          `mapstructure:"enabled" default:"true"`
	ListenAddress   string        `mapstructure:"listen_address" default:"0.0.0.0"`
	Port            int           `mapstructure:"port" default:"10020"`
	
	// HTTP server configuration
	ReadTimeout     time.Duration `mapstructure:"read_timeout" default:"30s"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" default:"30s"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" default:"120s"`
	
	// TLS configuration
	TLS             TLSConfig     `mapstructure:"tls"`
	
	// Authentication configuration
	Auth            AuthConfig    `mapstructure:"auth"`
	
	// Rate limiting
	RateLimit       RateLimitConfig `mapstructure:"rate_limit"`
	
	// Policy management
	PolicyStorage   PolicyStorageConfig `mapstructure:"policy_storage"`
}

// O1Config contains O1 interface configuration
type O1Config struct {
	Enabled         bool          `mapstructure:"enabled" default:"true"`
	ListenAddress   string        `mapstructure:"listen_address" default:"0.0.0.0"`
	Port            int           `mapstructure:"port" default:"830"`
	
	// NETCONF configuration
	NETCONF         NETCONFConfig `mapstructure:"netconf"`
	
	// YANG models configuration
	YANG            YANGConfig    `mapstructure:"yang"`
	
	// SSH configuration for NETCONF
	SSH             SSHConfig     `mapstructure:"ssh"`
}

// NETCONFConfig contains NETCONF protocol configuration
type NETCONFConfig struct {
	Capabilities    []string      `mapstructure:"capabilities"`
	SessionTimeout  time.Duration `mapstructure:"session_timeout" default:"600s"`
	MaxSessions     int           `mapstructure:"max_sessions" default:"100"`
	HelloTimeout    time.Duration `mapstructure:"hello_timeout" default:"30s"`
}

// YANGConfig contains YANG model configuration
type YANGConfig struct {
	ModelPath       string        `mapstructure:"model_path" default:"/etc/near-rt-ric/yang"`
	Models          []YANGModel   `mapstructure:"models"`
	ValidateData    bool          `mapstructure:"validate_data" default:"true"`
}

// YANGModel represents a YANG model configuration
type YANGModel struct {
	Name        string   `mapstructure:"name" validate:"required"`
	Namespace   string   `mapstructure:"namespace" validate:"required"`
	Prefix      string   `mapstructure:"prefix" validate:"required"`
	Version     string   `mapstructure:"version" validate:"required"`
	Features    []string `mapstructure:"features"`
	Deviations  []string `mapstructure:"deviations"`
}

// SSHConfig contains SSH server configuration for NETCONF
type SSHConfig struct {
	HostKey         string        `mapstructure:"host_key" validate:"required"`
	AuthMethods     []string      `mapstructure:"auth_methods" default:"password,publickey"`
	MaxAuthTries    int           `mapstructure:"max_auth_tries" default:"3"`
	LoginTimeout    time.Duration `mapstructure:"login_timeout" default:"30s"`
}

// XAppConfig contains xApp framework configuration
type XAppConfig struct {
	Enabled         bool          `mapstructure:"enabled" default:"true"`
	Registry        RegistryConfig `mapstructure:"registry"`
	Lifecycle       LifecycleConfig `mapstructure:"lifecycle"`
	Messaging       MessagingConfig `mapstructure:"messaging"`
}

// RegistryConfig contains xApp registry configuration
type RegistryConfig struct {
	URL             string        `mapstructure:"url" default:"http://localhost:8080"`
	Timeout         time.Duration `mapstructure:"timeout" default:"30s"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval" default:"60s"`
}

// LifecycleConfig contains xApp lifecycle management configuration
type LifecycleConfig struct {
	DeploymentTimeout time.Duration `mapstructure:"deployment_timeout" default:"300s"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval" default:"30s"`
	MaxRestartAttempts int           `mapstructure:"max_restart_attempts" default:"3"`
}

// MessagingConfig contains xApp messaging configuration
type MessagingConfig struct {
	Protocol        string        `mapstructure:"protocol" default:"rmr"`
	BufferSize      int           `mapstructure:"buffer_size" default:"4096"`
	ThreadCount     int           `mapstructure:"thread_count" default:"4"`
}

// TLSConfig contains TLS configuration
type TLSConfig struct {
	Enabled         bool          `mapstructure:"enabled" default:"false"`
	CertFile        string        `mapstructure:"cert_file"`
	KeyFile         string        `mapstructure:"key_file"`
	CAFile          string        `mapstructure:"ca_file"`
	MinVersion      string        `mapstructure:"min_version" default:"1.2"`
	CipherSuites    []string      `mapstructure:"cipher_suites"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	Enabled         bool          `mapstructure:"enabled" default:"true"`
	Method          string        `mapstructure:"method" default:"jwt" validate:"oneof=jwt oauth2 basic"`
	
	// JWT configuration
	JWT             JWTConfig     `mapstructure:"jwt"`
	
	// OAuth2 configuration
	OAuth2          OAuth2Config  `mapstructure:"oauth2"`
}

// JWTConfig contains JWT authentication configuration
type JWTConfig struct {
	SecretKey       string        `mapstructure:"secret_key" validate:"required"`
	PublicKeyFile   string        `mapstructure:"public_key_file"`
	PrivateKeyFile  string        `mapstructure:"private_key_file"`
	Algorithm       string        `mapstructure:"algorithm" default:"HS256"`
	TokenExpiry     time.Duration `mapstructure:"token_expiry" default:"24h"`
	RefreshExpiry   time.Duration `mapstructure:"refresh_expiry" default:"168h"`
	Issuer          string        `mapstructure:"issuer" default:"near-rt-ric"`
	Audience        []string      `mapstructure:"audience"`
}

// OAuth2Config contains OAuth2 configuration
type OAuth2Config struct {
	Provider        string        `mapstructure:"provider" validate:"required"`
	ClientID        string        `mapstructure:"client_id" validate:"required"`
	ClientSecret    string        `mapstructure:"client_secret" validate:"required"`
	RedirectURL     string        `mapstructure:"redirect_url" validate:"required"`
	Scopes          []string      `mapstructure:"scopes"`
	AuthURL         string        `mapstructure:"auth_url"`
	TokenURL        string        `mapstructure:"token_url"`
	UserInfoURL     string        `mapstructure:"user_info_url"`
}

// RateLimitConfig contains rate limiting configuration
type RateLimitConfig struct {
	Enabled         bool          `mapstructure:"enabled" default:"true"`
	RequestsPerSecond int         `mapstructure:"requests_per_second" default:"100"`
	BurstSize       int           `mapstructure:"burst_size" default:"200"`
	WindowSize      time.Duration `mapstructure:"window_size" default:"60s"`
}

// PolicyStorageConfig contains policy storage configuration
type PolicyStorageConfig struct {
	Backend         string        `mapstructure:"backend" default:"database" validate:"oneof=database redis memory"`
	RetentionPeriod time.Duration `mapstructure:"retention_period" default:"168h"`
	MaxPolicies     int           `mapstructure:"max_policies" default:"10000"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	// PostgreSQL configuration
	Host            string        `mapstructure:"host" default:"localhost"`
	Port            int           `mapstructure:"port" default:"5432"`
	Database        string        `mapstructure:"database" default:"near_rt_ric"`
	Username        string        `mapstructure:"username" default:"postgres"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode" default:"prefer"`
	
	// Connection pool settings
	MaxOpenConns    int           `mapstructure:"max_open_conns" default:"25"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" default:"5"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" default:"300s"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" default:"60s"`
	
	// Migration settings
	AutoMigrate     bool          `mapstructure:"auto_migrate" default:"true"`
	MigrationsPath  string        `mapstructure:"migrations_path" default:"/etc/near-rt-ric/migrations"`
}

// RedisConfig contains Redis configuration
type RedisConfig struct {
	// Connection settings
	Address         string        `mapstructure:"address" default:"localhost:6379"`
	Password        string        `mapstructure:"password"`
	Database        int           `mapstructure:"database" default:"0"`
	
	// Pool settings
	PoolSize        int           `mapstructure:"pool_size" default:"10"`
	MinIdleConns    int           `mapstructure:"min_idle_conns" default:"2"`
	
	// Timeout settings
	DialTimeout     time.Duration `mapstructure:"dial_timeout" default:"5s"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" default:"3s"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" default:"3s"`
	
	// Cluster settings (for Redis Cluster)
	Cluster         bool          `mapstructure:"cluster" default:"false"`
	ClusterAddrs    []string      `mapstructure:"cluster_addrs"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level           string        `mapstructure:"level" default:"info"`
	Format          string        `mapstructure:"format" default:"json" validate:"oneof=json text"`
	Output          string        `mapstructure:"output" default:"stdout"`
	File            string        `mapstructure:"file"`
	MaxSize         int           `mapstructure:"max_size" default:"100"`      // MB
	MaxBackups      int           `mapstructure:"max_backups" default:"5"`
	MaxAge          int           `mapstructure:"max_age" default:"30"`        // days
	Compress        bool          `mapstructure:"compress" default:"true"`
	ReportCaller    bool          `mapstructure:"report_caller" default:"false"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	// Encryption settings
	EncryptionKey   string        `mapstructure:"encryption_key" validate:"required"`
	HashCost        int           `mapstructure:"hash_cost" default:"12"`
	
	// Session settings
	SessionTimeout  time.Duration `mapstructure:"session_timeout" default:"1800s"`
	SessionSecret   string        `mapstructure:"session_secret" validate:"required"`
	SecureCookies   bool          `mapstructure:"secure_cookies" default:"true"`
	
	// CORS settings
	CORS            CORSConfig    `mapstructure:"cors"`
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	Enabled         bool          `mapstructure:"enabled" default:"true"`
	AllowedOrigins  []string      `mapstructure:"allowed_origins"`
	AllowedMethods  []string      `mapstructure:"allowed_methods" default:"GET,POST,PUT,DELETE,OPTIONS"`
	AllowedHeaders  []string      `mapstructure:"allowed_headers"`
	ExposedHeaders  []string      `mapstructure:"exposed_headers"`
	AllowCredentials bool         `mapstructure:"allow_credentials" default:"true"`
	MaxAge          int           `mapstructure:"max_age" default:"3600"`
}

// MetricsConfig contains metrics configuration
type MetricsConfig struct {
	Enabled         bool          `mapstructure:"enabled" default:"true"`
	Path            string        `mapstructure:"path" default:"/metrics"`
	Port            int           `mapstructure:"port" default:"9090"`
	
	// Prometheus configuration
	Prometheus      PrometheusConfig `mapstructure:"prometheus"`
}

// PrometheusConfig contains Prometheus-specific configuration
type PrometheusConfig struct {
	Namespace       string        `mapstructure:"namespace" default:"near_rt_ric"`
	Subsystem       string        `mapstructure:"subsystem"`
	Labels          map[string]string `mapstructure:"labels"`
}

// TracingConfig contains distributed tracing configuration
type TracingConfig struct {
	Enabled         bool          `mapstructure:"enabled" default:"true"`
	ServiceName     string        `mapstructure:"service_name" default:"near-rt-ric"`
	
	// Jaeger configuration
	Jaeger          JaegerConfig  `mapstructure:"jaeger"`
}

// JaegerConfig contains Jaeger tracing configuration
type JaegerConfig struct {
	Endpoint        string        `mapstructure:"endpoint" default:"http://localhost:14268/api/traces"`
	SamplingRate    float64       `mapstructure:"sampling_rate" default:"0.1"`
	MaxTagValueLength int         `mapstructure:"max_tag_value_length" default:"256"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configFile string) (*Config, error) {
	// Set up Viper
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
	
	// Enable environment variable support
	viper.AutomaticEnv()
	viper.SetEnvPrefix("RIC")
	
	// Set defaults
	setDefaults()
	
	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional in development
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}
	
	// Unmarshal into config struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Process sensitive configuration
	if err := processSecrets(&config); err != nil {
		return nil, fmt.Errorf("failed to process secrets: %w", err)
	}
	
	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Global defaults
	viper.SetDefault("global.environment", "development")
	viper.SetDefault("global.node_id", "near-rt-ric-001")
	viper.SetDefault("global.cluster_id", "cluster-001")
	viper.SetDefault("global.log_level", "info")
	
	// E2 defaults
	viper.SetDefault("e2.enabled", true)
	viper.SetDefault("e2.listen_address", "0.0.0.0")
	viper.SetDefault("e2.port", 36421)
	viper.SetDefault("e2.max_connections", 1000)
	
	// A1 defaults
	viper.SetDefault("a1.enabled", true)
	viper.SetDefault("a1.listen_address", "0.0.0.0")
	viper.SetDefault("a1.port", 10020)
	
	// O1 defaults
	viper.SetDefault("o1.enabled", true)
	viper.SetDefault("o1.listen_address", "0.0.0.0")
	viper.SetDefault("o1.port", 830)
	
	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.database", "near_rt_ric")
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.ssl_mode", "prefer")
	
	// Security defaults
	viper.SetDefault("security.hash_cost", bcrypt.DefaultCost)
	viper.SetDefault("security.secure_cookies", true)
}

// validateConfig validates the loaded configuration
func validateConfig(config *Config) error {
	// Validate environment
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}
	if !validEnvs[config.Global.Environment] {
		return fmt.Errorf("invalid environment: %s", config.Global.Environment)
	}
	
	// Validate required fields for production
	if config.Global.Environment == "production" {
		if config.Security.EncryptionKey == "" {
			return fmt.Errorf("encryption_key is required for production")
		}
		if config.Security.SessionSecret == "" {
			return fmt.Errorf("session_secret is required for production")
		}
		if config.A1.Auth.JWT.SecretKey == "" {
			return fmt.Errorf("JWT secret_key is required for production")
		}
	}
	
	// Validate port numbers
	ports := []int{config.E2.Port, config.A1.Port, config.O1.Port}
	for _, port := range ports {
		if port < 1 || port > 65535 {
			return fmt.Errorf("invalid port number: %d", port)
		}
	}
	
	return nil
}

// processSecrets handles sensitive configuration like passwords and keys
func processSecrets(config *Config) error {
	// Generate default secrets for development
	if config.Global.Environment == "development" {
		if config.Security.EncryptionKey == "" {
			config.Security.EncryptionKey = "dev-encryption-key-change-in-production"
		}
		if config.Security.SessionSecret == "" {
			config.Security.SessionSecret = "dev-session-secret-change-in-production"
		}
		if config.A1.Auth.JWT.SecretKey == "" {
			config.A1.Auth.JWT.SecretKey = "dev-jwt-secret-change-in-production"
		}
	}
	
	// Load secrets from files if specified
	if keyFile := os.Getenv("RIC_ENCRYPTION_KEY_FILE"); keyFile != "" {
		key, err := os.ReadFile(keyFile)
		if err != nil {
			return fmt.Errorf("failed to read encryption key file: %w", err)
		}
		config.Security.EncryptionKey = string(key)
	}
	
	return nil
}

// GetDefaultConfig returns a default configuration for development
func GetDefaultConfig() *Config {
	config := &Config{
		Global: GlobalConfig{
			Environment: "development",
			NodeID:      "near-rt-ric-001",
			ClusterID:   "cluster-001",
			LogLevel:    "info",
		},
		E2: E2Config{
			Enabled:       true,
			ListenAddress: "0.0.0.0",
			Port:          36421,
			MaxConnections: 1000,
			ConnectionTimeout: 30 * time.Second,
			HeartbeatInterval: 30 * time.Second,
			MessageTimeout:    10 * time.Second,
		},
		A1: A1Config{
			Enabled:       true,
			ListenAddress: "0.0.0.0",
			Port:          10020,
			ReadTimeout:   30 * time.Second,
			WriteTimeout:  30 * time.Second,
			IdleTimeout:   120 * time.Second,
		},
		O1: O1Config{
			Enabled:       true,
			ListenAddress: "0.0.0.0",
			Port:          830,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "near_rt_ric",
			Username: "postgres",
			SSLMode:  "prefer",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Security: SecurityConfig{
			EncryptionKey: "dev-encryption-key-change-in-production",
			SessionSecret: "dev-session-secret-change-in-production",
			HashCost:      bcrypt.DefaultCost,
		},
	}
	
	return config
}