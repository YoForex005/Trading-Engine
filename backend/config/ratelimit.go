package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// RateLimitingConfig holds rate limiting configuration from server.yaml
type RateLimitingConfig struct {
	Enabled          bool              `yaml:"enabled"`
	RequestsPerSecond float64          `yaml:"requests_per_second"`
	RequestsPerMinute float64          `yaml:"requests_per_minute"`
	BurstSize        int               `yaml:"burst_size"`
	CleanupInterval  string            `yaml:"cleanup_interval"`
	ClientTimeout    string            `yaml:"client_timeout"`
	Exclusions       []string          `yaml:"exclusions"`
	Endpoints        map[string]EndpointLimitConfig `yaml:"endpoints"`
}

// EndpointLimitConfig holds rate limit config for specific endpoints
type EndpointLimitConfig struct {
	RequestsPerSecond float64 `yaml:"requests_per_second"`
	BurstSize        int     `yaml:"burst_size"`
}

// KeyBasedRateLimitingConfig holds key-based rate limiting configuration
type KeyBasedRateLimitingConfig struct {
	Enabled            bool                            `yaml:"enabled"`
	RequestsPerSecond  float64                         `yaml:"requests_per_second"`
	RequestsPerMinute  float64                         `yaml:"requests_per_minute"`
	BurstSize         int                             `yaml:"burst_size"`
	CleanupInterval   string                          `yaml:"cleanup_interval"`
	ClientTimeout     string                          `yaml:"client_timeout"`
	Tiers             map[string]TierLimitConfig      `yaml:"tiers"`
}

// TierLimitConfig holds tier-based rate limit configuration
type TierLimitConfig struct {
	RequestsPerSecond float64 `yaml:"requests_per_second"`
	BurstSize        int     `yaml:"burst_size"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host              string `yaml:"host"`
	Port              string `yaml:"port"`
	ReadTimeout       string `yaml:"read_timeout"`
	WriteTimeout      string `yaml:"write_timeout"`
	IdleTimeout       string `yaml:"idle_timeout"`
	MaxRequestSize    int64  `yaml:"max_request_size"`
}

// FullServerYAMLConfig represents the complete server.yaml structure
type FullServerYAMLConfig struct {
	Server                      ServerConfig                   `yaml:"server"`
	RateLimiting                RateLimitingConfig             `yaml:"rate_limiting"`
	KeyBasedRateLimiting        KeyBasedRateLimitingConfig     `yaml:"key_based_rate_limiting"`
	Logging                     interface{}                    `yaml:"logging"`
	Monitoring                  interface{}                    `yaml:"monitoring"`
	APIGateway                  interface{}                    `yaml:"api_gateway"`
	Security                    interface{}                    `yaml:"security"`
	WebSocket                   interface{}                    `yaml:"websocket"`
	Database                    interface{}                    `yaml:"database"`
	Cache                       interface{}                    `yaml:"cache"`
}

// LoadRateLimitingConfig loads rate limiting configuration from server.yaml
func LoadRateLimitingConfig() (RateLimitingConfig, error) {
	configPath := os.Getenv("SERVER_CONFIG")
	if configPath == "" {
		configPath = "config/server.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Return sensible defaults if config file not found
		return RateLimitingConfig{
			Enabled:           true,
			RequestsPerSecond: 10,
			RequestsPerMinute: 500,
			BurstSize:         20,
			CleanupInterval:   "5m",
			ClientTimeout:     "10m",
			Exclusions:        []string{"/health", "/docs", "/swagger.yaml"},
			Endpoints:         make(map[string]EndpointLimitConfig),
		}, nil
	}

	var fullConfig FullServerYAMLConfig
	if err := yaml.Unmarshal(data, &fullConfig); err != nil {
		return RateLimitingConfig{}, fmt.Errorf("failed to parse server.yaml: %w", err)
	}

	return fullConfig.RateLimiting, nil
}

// LoadKeyBasedRateLimitingConfig loads key-based rate limiting configuration
func LoadKeyBasedRateLimitingConfig() (KeyBasedRateLimitingConfig, error) {
	configPath := os.Getenv("SERVER_CONFIG")
	if configPath == "" {
		configPath = "config/server.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Return sensible defaults if config file not found
		return KeyBasedRateLimitingConfig{
			Enabled:           true,
			RequestsPerSecond: 20,
			RequestsPerMinute: 1000,
			BurstSize:         50,
			CleanupInterval:   "5m",
			ClientTimeout:     "15m",
			Tiers:             make(map[string]TierLimitConfig),
		}, nil
	}

	var fullConfig FullServerYAMLConfig
	if err := yaml.Unmarshal(data, &fullConfig); err != nil {
		return KeyBasedRateLimitingConfig{}, fmt.Errorf("failed to parse server.yaml: %w", err)
	}

	return fullConfig.KeyBasedRateLimiting, nil
}

// ParseDuration parses a duration string
func ParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		// Default to 5 minutes if parsing fails
		return 5 * time.Minute
	}
	return d
}
