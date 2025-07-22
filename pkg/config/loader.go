package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config represents the overall application configuration
type Config struct {
	LogLevel string `mapstructure:"log_level"`
	Server   ServerConfig
	Security SecurityConfig
	Database DatabaseConfig
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	ConnectionString string `mapstructure:"connection_string"`
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
	Timeout int    `mapstructure:"timeout"`
	TLS     TLSConfig
}

// TLSConfig represents the TLS configuration
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertPath string `mapstructure:"cert_path"`
	KeyPath  string `mapstructure:"key_path"`
}

// SecurityConfig represents the security configuration
type SecurityConfig struct {
	AuthEnabled   bool   `mapstructure:"auth_enabled"`
	JWTSigningKey string `mapstructure:"jwt_signing_key"`
}

// LoadConfig loads the configuration for the specified service
func LoadConfig(serviceName string) (*Config, error) {
	viper.SetConfigName(strings.ToLower(serviceName) + "-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/ric")

	// Set defaults
	viper.SetDefault("log_level", "info")
	viper.SetDefault("server.address", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.timeout", 30)
	viper.SetDefault("server.tls.enabled", false)
	viper.SetDefault("security.auth_enabled", false)
	viper.SetDefault("security.jwt_signing_key", "")
	viper.SetDefault("database.connection_string", "postgres://user:password@localhost:5432/database?sslmode=disable")

	// Enable environment variable overriding
	viper.SetEnvPrefix(strings.ToUpper(serviceName))
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if we have env vars
		} else {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
