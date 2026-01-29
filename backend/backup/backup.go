package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// BackupType represents the type of backup
type BackupType string

const (
	BackupTypeFull        BackupType = "full"
	BackupTypeIncremental BackupType = "incremental"
	BackupTypeSnapshot    BackupType = "snapshot"
)

// BackupConfig holds backup configuration
type BackupConfig struct {
	// Database configuration
	DBHost     string
	DBPort     int
	DBName     string
	DBUser     string
	DBPassword string

	// Redis configuration
	RedisHost string
	RedisPort int
	RedisDB   int

	// Backup settings
	BackupRoot       string
	EnableEncryption bool
	GPGRecipient     string
	EnableS3         bool
	S3Bucket         string
	EnableOffsite    bool
	OffsiteHost      string

	// Retention policy
	RetentionDaily   int
	RetentionWeekly  int
	RetentionMonthly int
}

// BackupMetadata contains backup metadata
type BackupMetadata struct {
	BackupID        string     `json:"backup_id"`
	BackupType      BackupType `json:"backup_type"`
	Timestamp       time.Time  `json:"timestamp"`
	Hostname        string     `json:"hostname"`
	Components      []string   `json:"components"`
	Size            int64      `json:"size"`
	Compressed      bool       `json:"compressed"`
	Encrypted       bool       `json:"encrypted"`
	BaseBackupID    string     `json:"base_backup_id,omitempty"`
	PostgresVersion string     `json:"postgres_version,omitempty"`
	RedisVersion    string     `json:"redis_version,omitempty"`
}

// BackupManager manages backup operations
type BackupManager struct {
	config *BackupConfig
	logger Logger
}

// Logger interface for backup logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// NewBackupManager creates a new backup manager
func NewBackupManager(config *BackupConfig, logger Logger) *BackupManager {
	return &BackupManager{
		config: config,
		logger: logger,
	}
}

// CreateFullBackup creates a full system backup
func (bm *BackupManager) CreateFullBackup(ctx context.Context) (*BackupMetadata, error) {
	backupID := fmt.Sprintf("%s", time.Now().Format("20060102-150405"))
	backupDir := filepath.Join(bm.config.BackupRoot, fmt.Sprintf("full-%s", backupID))

	bm.logger.Info("Starting full backup", "backup_id", backupID)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	metadata := &BackupMetadata{
		BackupID:   backupID,
		BackupType: BackupTypeFull,
		Timestamp:  time.Now(),
		Components: []string{},
		Compressed: true,
		Encrypted:  bm.config.EnableEncryption,
	}

	// Get hostname
	hostname, _ := os.Hostname()
	metadata.Hostname = hostname

	// 1. Backup PostgreSQL
	if err := bm.backupPostgreSQL(ctx, backupDir, metadata); err != nil {
		return nil, fmt.Errorf("postgresql backup failed: %w", err)
	}

	// 2. Backup Redis
	if err := bm.backupRedis(ctx, backupDir, metadata); err != nil {
		bm.logger.Warn("Redis backup failed", "error", err)
	}

	// 3. Backup configuration files
	if err := bm.backupConfig(ctx, backupDir, metadata); err != nil {
		bm.logger.Warn("Config backup failed", "error", err)
	}

	// 4. Backup application state
	if err := bm.backupApplicationState(ctx, backupDir, metadata); err != nil {
		bm.logger.Warn("Application state backup failed", "error", err)
	}

	// 5. Create metadata file
	if err := bm.writeMetadata(backupDir, metadata); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	// 6. Compress backup
	compressedFile, err := bm.compressBackup(ctx, backupDir, backupID)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	// Get final size
	fileInfo, err := os.Stat(compressedFile)
	if err == nil {
		metadata.Size = fileInfo.Size()
	}

	// 7. Encrypt if enabled
	finalFile := compressedFile
	if bm.config.EnableEncryption {
		finalFile, err = bm.encryptBackup(ctx, compressedFile)
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
		os.Remove(compressedFile) // Remove unencrypted file
	}

	// 8. Upload to remote locations
	if bm.config.EnableS3 {
		if err := bm.uploadToS3(ctx, finalFile, metadata); err != nil {
			bm.logger.Error("S3 upload failed", "error", err)
		}
	}

	if bm.config.EnableOffsite {
		if err := bm.uploadToOffsite(ctx, finalFile, metadata); err != nil {
			bm.logger.Error("Offsite upload failed", "error", err)
		}
	}

	bm.logger.Info("Full backup completed", "backup_id", backupID, "size", metadata.Size)

	return metadata, nil
}

// backupPostgreSQL backs up the PostgreSQL database
func (bm *BackupManager) backupPostgreSQL(ctx context.Context, backupDir string, metadata *BackupMetadata) error {
	bm.logger.Info("Backing up PostgreSQL database")

	dumpFile := filepath.Join(backupDir, fmt.Sprintf("postgres-%s.dump", metadata.BackupID))

	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", bm.config.DBHost,
		"-p", fmt.Sprintf("%d", bm.config.DBPort),
		"-U", bm.config.DBUser,
		"-d", bm.config.DBName,
		"--format=custom",
		"--compress=9",
		fmt.Sprintf("--file=%s", dumpFile),
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", bm.config.DBPassword))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_dump failed: %w, output: %s", err, string(output))
	}

	metadata.Components = append(metadata.Components, "postgresql")

	// Get PostgreSQL version
	versionCmd := exec.CommandContext(ctx, "psql",
		"-h", bm.config.DBHost,
		"-p", fmt.Sprintf("%d", bm.config.DBPort),
		"-U", bm.config.DBUser,
		"-d", bm.config.DBName,
		"-t", "-c", "SELECT version();",
	)
	versionCmd.Env = cmd.Env

	if versionOutput, err := versionCmd.Output(); err == nil {
		metadata.PostgresVersion = string(versionOutput)
	}

	fileInfo, _ := os.Stat(dumpFile)
	bm.logger.Info("PostgreSQL backup completed", "size", fileInfo.Size())

	return nil
}

// backupRedis backs up Redis data
func (bm *BackupManager) backupRedis(ctx context.Context, backupDir string, metadata *BackupMetadata) error {
	bm.logger.Info("Backing up Redis data")

	// Trigger Redis BGSAVE
	cmd := exec.CommandContext(ctx, "redis-cli",
		"-h", bm.config.RedisHost,
		"-p", fmt.Sprintf("%d", bm.config.RedisPort),
		"BGSAVE",
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("redis BGSAVE failed: %w", err)
	}

	// Wait for BGSAVE to complete (simplified - production should poll LASTSAVE)
	time.Sleep(2 * time.Second)

	// Copy RDB file
	rdbFile := "/var/lib/redis/dump.rdb" // Default path
	destFile := filepath.Join(backupDir, fmt.Sprintf("redis-%s.rdb", metadata.BackupID))

	if err := copyFile(rdbFile, destFile); err != nil {
		return fmt.Errorf("failed to copy Redis RDB: %w", err)
	}

	metadata.Components = append(metadata.Components, "redis")

	// Get Redis version
	versionCmd := exec.CommandContext(ctx, "redis-cli",
		"-h", bm.config.RedisHost,
		"-p", fmt.Sprintf("%d", bm.config.RedisPort),
		"INFO", "server",
	)

	if versionOutput, err := versionCmd.Output(); err == nil {
		metadata.RedisVersion = string(versionOutput)
	}

	bm.logger.Info("Redis backup completed")

	return nil
}

// backupConfig backs up configuration files
func (bm *BackupManager) backupConfig(ctx context.Context, backupDir string, metadata *BackupMetadata) error {
	bm.logger.Info("Backing up configuration files")

	configDir := filepath.Join(backupDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Define config paths to backup
	configPaths := []string{
		"/etc/trading-engine",
		".env",
		"config.yaml",
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			if err := copyDir(path, filepath.Join(configDir, filepath.Base(path))); err != nil {
				bm.logger.Warn("Failed to backup config", "path", path, "error", err)
			}
		}
	}

	metadata.Components = append(metadata.Components, "configuration")

	return nil
}

// backupApplicationState backs up application state
func (bm *BackupManager) backupApplicationState(ctx context.Context, backupDir string, metadata *BackupMetadata) error {
	bm.logger.Info("Backing up application state")

	stateDir := filepath.Join(backupDir, "state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return err
	}

	// Backup data directory
	dataPath := "./data"
	if _, err := os.Stat(dataPath); err == nil {
		if err := copyDir(dataPath, filepath.Join(stateDir, "data")); err != nil {
			return err
		}
	}

	metadata.Components = append(metadata.Components, "application_state")

	return nil
}

// writeMetadata writes backup metadata to file
func (bm *BackupManager) writeMetadata(backupDir string, metadata *BackupMetadata) error {
	metadataFile := filepath.Join(backupDir, "backup-metadata.json")

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataFile, data, 0644)
}

// compressBackup compresses the backup directory
func (bm *BackupManager) compressBackup(ctx context.Context, backupDir, backupID string) (string, error) {
	bm.logger.Info("Compressing backup")

	outputFile := filepath.Join(bm.config.BackupRoot, fmt.Sprintf("backup-%s.tar.gz", backupID))

	cmd := exec.CommandContext(ctx, "tar", "-czf", outputFile, "-C", backupDir, ".")
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tar compression failed: %w", err)
	}

	// Remove original directory to save space
	os.RemoveAll(backupDir)

	return outputFile, nil
}

// encryptBackup encrypts a backup file using GPG
func (bm *BackupManager) encryptBackup(ctx context.Context, inputFile string) (string, error) {
	bm.logger.Info("Encrypting backup")

	outputFile := inputFile + ".gpg"

	cmd := exec.CommandContext(ctx, "gpg",
		"--encrypt",
		"--recipient", bm.config.GPGRecipient,
		"--output", outputFile,
		inputFile,
	)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gpg encryption failed: %w", err)
	}

	return outputFile, nil
}

// uploadToS3 uploads backup to S3
func (bm *BackupManager) uploadToS3(ctx context.Context, file string, metadata *BackupMetadata) error {
	bm.logger.Info("Uploading to S3")

	s3Path := fmt.Sprintf("s3://%s/backups/%s/backup-%s.tar.gz.gpg",
		bm.config.S3Bucket, metadata.BackupType, metadata.BackupID)

	cmd := exec.CommandContext(ctx, "aws", "s3", "cp", file, s3Path,
		"--storage-class", "STANDARD_IA",
		"--metadata", fmt.Sprintf("backup-type=%s,backup-id=%s", metadata.BackupType, metadata.BackupID),
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("s3 upload failed: %w", err)
	}

	return nil
}

// uploadToOffsite uploads backup to offsite location via rsync
func (bm *BackupManager) uploadToOffsite(ctx context.Context, file string, metadata *BackupMetadata) error {
	bm.logger.Info("Uploading to offsite location")

	remotePath := fmt.Sprintf("%s:/backups/trading-engine/%s/", bm.config.OffsiteHost, metadata.BackupType)

	cmd := exec.CommandContext(ctx, "rsync", "-avz", "--progress", file, remotePath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsync to offsite failed: %w", err)
	}

	return nil
}

// Utility functions

// copyFile copies a single file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcInfo.IsDir() {
		return copyFile(src, dst)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
