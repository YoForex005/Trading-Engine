package compression

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// RetentionConfig holds the complete retention configuration
type RetentionConfig struct {
	Compression CompressionConfig `yaml:"compression"`
	Retention   RetentionPolicy   `yaml:"retention"`
	Paths       PathsConfig       `yaml:"paths"`
	Operations  OperationsConfig  `yaml:"operations"`
	Logging     LoggingConfig     `yaml:"logging"`
}

// CompressionConfig holds compression-specific settings
type CompressionConfig struct {
	Enabled        bool   `yaml:"enabled"`
	Schedule       string `yaml:"schedule"`
	MaxAgeSeconds  int64  `yaml:"max_age_seconds"`
	MaxConcurrency int    `yaml:"max_concurrency"`
}

// RetentionPolicy defines retention thresholds
type RetentionPolicy struct {
	ArchiveThresholdDays  int  `yaml:"archive_threshold_days"`
	DeletionThresholdDays int  `yaml:"deletion_threshold_days"`
	BackupBeforeDelete    bool `yaml:"backup_before_delete"`
}

// PathsConfig defines directory paths
type PathsConfig struct {
	TicksDirectory   string `yaml:"ticks_directory"`
	ArchiveDirectory string `yaml:"archive_directory"`
	LogsDirectory    string `yaml:"logs_directory"`
	LogFile          string `yaml:"log_file"`
}

// OperationsConfig defines which operations to enable
type OperationsConfig struct {
	EnableArchival   bool `yaml:"enable_archival"`
	EnableDeletion   bool `yaml:"enable_deletion"`
	CompressArchives bool `yaml:"compress_archives"`
	CreateBackup     bool `yaml:"create_backup"`
	DryRun           bool `yaml:"dry_run"`
}

// LoggingConfig defines logging behavior
type LoggingConfig struct {
	Level             string `yaml:"level"`
	MaxLogFiles       int    `yaml:"max_log_files"`
	MaxLogSizeMB      int    `yaml:"max_log_size_mb"`
	VerboseTimestamps bool   `yaml:"verbose_timestamps"`
}

// LoadRetentionConfig loads compression/retention configuration from YAML file
func LoadRetentionConfig(configPath string) (*RetentionConfig, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("config file not found at %s: %w", configPath, err)
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config RetentionConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Set defaults if not specified
	if config.Compression.MaxAgeSeconds == 0 {
		config.Compression.MaxAgeSeconds = 7 * 24 * 3600 // 7 days
	}
	if config.Compression.MaxConcurrency == 0 {
		config.Compression.MaxConcurrency = 4
	}
	if config.Compression.Schedule == "" {
		config.Compression.Schedule = "168h" // 1 week
	}

	// Make paths absolute relative to config directory
	if !filepath.IsAbs(config.Paths.TicksDirectory) {
		configDir := filepath.Dir(configPath)
		config.Paths.TicksDirectory = filepath.Join(configDir, "..", config.Paths.TicksDirectory)
	}
	if !filepath.IsAbs(config.Paths.ArchiveDirectory) {
		configDir := filepath.Dir(configPath)
		config.Paths.ArchiveDirectory = filepath.Join(configDir, "..", config.Paths.ArchiveDirectory)
	}

	return &config, nil
}

// ToCompressorConfig converts RetentionConfig to Compressor Config
func (rc *RetentionConfig) ToCompressorConfig() Config {
	return Config{
		Enabled:        rc.Compression.Enabled,
		DataDir:        rc.Paths.TicksDirectory,
		MaxAgeSeconds:  rc.Compression.MaxAgeSeconds,
		Schedule:       rc.Compression.Schedule,
		MaxConcurrency: rc.Compression.MaxConcurrency,
	}
}
