package backup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// RestoreOptions contains options for restore operations
type RestoreOptions struct {
	BackupFile        string
	TargetTime        *time.Time // For point-in-time recovery
	Force             bool
	VerifyBeforeApply bool
	CreateSafetyBackup bool
}

// RestoreResult contains the result of a restore operation
type RestoreResult struct {
	Success         bool
	RestoredBackupID string
	TablesRestored  int
	Duration        time.Duration
	DataLossMinutes int
	SafetyBackupPath string
	Errors          []error
}

// RestoreManager handles database restoration
type RestoreManager struct {
	config *BackupConfig
	logger Logger
}

// NewRestoreManager creates a new restore manager
func NewRestoreManager(config *BackupConfig, logger Logger) *RestoreManager {
	return &RestoreManager{
		config: config,
		logger: logger,
	}
}

// RestoreFull performs a full system restore
func (rm *RestoreManager) RestoreFull(ctx context.Context, opts RestoreOptions) (*RestoreResult, error) {
	rm.logger.Info("Starting full system restore", "backup_file", opts.BackupFile)

	startTime := time.Now()
	result := &RestoreResult{
		Errors: []error{},
	}

	if !opts.Force {
		return nil, fmt.Errorf("restore requires Force=true for safety")
	}

	// 1. Verify backup file exists
	if _, err := os.Stat(opts.BackupFile); err != nil {
		return nil, fmt.Errorf("backup file not found: %w", err)
	}

	// 2. Create safety backup of current database
	if opts.CreateSafetyBackup {
		safetyBackup, err := rm.createSafetyBackup(ctx)
		if err != nil {
			rm.logger.Warn("Failed to create safety backup", "error", err)
			result.Errors = append(result.Errors, fmt.Errorf("safety backup failed: %w", err))
		} else {
			result.SafetyBackupPath = safetyBackup
			rm.logger.Info("Safety backup created", "path", safetyBackup)
		}
	}

	// 3. Extract and decrypt backup
	extractDir, metadata, err := rm.extractBackup(ctx, opts.BackupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to extract backup: %w", err)
	}
	defer os.RemoveAll(extractDir)

	result.RestoredBackupID = metadata.BackupID

	// 4. Stop application services
	if err := rm.stopServices(ctx); err != nil {
		rm.logger.Warn("Failed to stop services", "error", err)
	}

	// 5. Restore PostgreSQL
	if err := rm.restorePostgreSQL(ctx, extractDir, metadata); err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Errorf("postgresql restore failed: %w", err))
		return result, err
	}

	// 6. Restore Redis
	if err := rm.restoreRedis(ctx, extractDir, metadata); err != nil {
		rm.logger.Warn("Redis restore failed", "error", err)
		result.Errors = append(result.Errors, err)
	}

	// 7. Restore configuration
	if err := rm.restoreConfig(ctx, extractDir); err != nil {
		rm.logger.Warn("Config restore failed", "error", err)
		result.Errors = append(result.Errors, err)
	}

	// 8. Restore application state
	if err := rm.restoreApplicationState(ctx, extractDir); err != nil {
		rm.logger.Warn("Application state restore failed", "error", err)
		result.Errors = append(result.Errors, err)
	}

	// 9. Verify restoration
	tableCount, err := rm.verifyRestore(ctx)
	if err != nil {
		result.Errors = append(result.Errors, err)
	}
	result.TablesRestored = tableCount

	// 10. Start application services
	if err := rm.startServices(ctx); err != nil {
		rm.logger.Warn("Failed to start services", "error", err)
		result.Errors = append(result.Errors, err)
	}

	result.Duration = time.Since(startTime)
	result.Success = len(result.Errors) == 0

	rm.logger.Info("Restore completed",
		"success", result.Success,
		"duration", result.Duration,
		"tables", result.TablesRestored,
	)

	return result, nil
}

// RestorePointInTime performs point-in-time recovery using WAL files
func (rm *RestoreManager) RestorePointInTime(ctx context.Context, opts RestoreOptions) (*RestoreResult, error) {
	if opts.TargetTime == nil {
		return nil, fmt.Errorf("target time required for point-in-time recovery")
	}

	rm.logger.Info("Starting point-in-time recovery", "target_time", opts.TargetTime)

	startTime := time.Now()
	result := &RestoreResult{}

	// 1. Find appropriate base backup
	baseBackup, err := rm.findBaseBackup(ctx, *opts.TargetTime)
	if err != nil {
		return nil, fmt.Errorf("failed to find base backup: %w", err)
	}

	// 2. Extract base backup
	extractDir, metadata, err := rm.extractBackup(ctx, baseBackup)
	if err != nil {
		return nil, fmt.Errorf("failed to extract backup: %w", err)
	}
	defer os.RemoveAll(extractDir)

	result.RestoredBackupID = metadata.BackupID

	// 3. Create recovery configuration
	recoveryConf := rm.createRecoveryConfig(*opts.TargetTime, extractDir)

	// 4. Restore base backup
	if err := rm.restorePostgreSQL(ctx, extractDir, metadata); err != nil {
		return nil, fmt.Errorf("base backup restore failed: %w", err)
	}

	// 5. Apply recovery configuration
	pgDataDir := "/var/lib/postgresql/data" // Should be configurable
	recoveryFile := filepath.Join(pgDataDir, "recovery.conf")
	if err := os.WriteFile(recoveryFile, []byte(recoveryConf), 0600); err != nil {
		return nil, fmt.Errorf("failed to write recovery.conf: %w", err)
	}

	// 6. Restart PostgreSQL to trigger recovery
	if err := rm.restartPostgreSQL(ctx); err != nil {
		return nil, fmt.Errorf("failed to restart postgresql: %w", err)
	}

	// 7. Wait for recovery to complete
	if err := rm.waitForRecovery(ctx); err != nil {
		return nil, fmt.Errorf("recovery failed: %w", err)
	}

	result.Duration = time.Since(startTime)
	result.Success = true

	// Calculate data loss
	dataLoss := time.Since(*opts.TargetTime)
	result.DataLossMinutes = int(dataLoss.Minutes())

	rm.logger.Info("Point-in-time recovery completed",
		"target_time", opts.TargetTime,
		"data_loss_minutes", result.DataLossMinutes,
		"duration", result.Duration,
	)

	return result, nil
}

// createSafetyBackup creates a safety backup before restore
func (rm *RestoreManager) createSafetyBackup(ctx context.Context) (string, error) {
	safetyFile := fmt.Sprintf("/tmp/safety-backup-%d.dump", time.Now().Unix())

	cmd := exec.CommandContext(ctx, "pg_dump",
		"-h", rm.config.DBHost,
		"-p", fmt.Sprintf("%d", rm.config.DBPort),
		"-U", rm.config.DBUser,
		"-d", rm.config.DBName,
		"--format=custom",
		fmt.Sprintf("--file=%s", safetyFile),
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", rm.config.DBPassword))

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return safetyFile, nil
}

// extractBackup extracts and decrypts a backup file
func (rm *RestoreManager) extractBackup(ctx context.Context, backupFile string) (string, *BackupMetadata, error) {
	extractDir := filepath.Join("/tmp", fmt.Sprintf("restore-%d", time.Now().Unix()))
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return "", nil, err
	}

	// Decrypt if encrypted
	archiveFile := backupFile
	if filepath.Ext(backupFile) == ".gpg" {
		rm.logger.Info("Decrypting backup")

		decryptedFile := filepath.Join(extractDir, "backup.tar.gz")
		cmd := exec.CommandContext(ctx, "gpg", "--decrypt", "--output", decryptedFile, backupFile)

		if err := cmd.Run(); err != nil {
			os.RemoveAll(extractDir)
			return "", nil, fmt.Errorf("decryption failed: %w", err)
		}

		archiveFile = decryptedFile
	}

	// Extract archive
	rm.logger.Info("Extracting backup archive")

	cmd := exec.CommandContext(ctx, "tar", "-xzf", archiveFile, "-C", extractDir)
	if err := cmd.Run(); err != nil {
		os.RemoveAll(extractDir)
		return "", nil, fmt.Errorf("extraction failed: %w", err)
	}

	// Read metadata
	metadataFile := filepath.Join(extractDir, "backup-metadata.json")
	metadataData, err := os.ReadFile(metadataFile)
	if err != nil {
		return extractDir, nil, fmt.Errorf("metadata not found: %w", err)
	}

	var metadata BackupMetadata
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		return extractDir, nil, fmt.Errorf("invalid metadata: %w", err)
	}

	return extractDir, &metadata, nil
}

// restorePostgreSQL restores the PostgreSQL database
func (rm *RestoreManager) restorePostgreSQL(ctx context.Context, extractDir string, metadata *BackupMetadata) error {
	rm.logger.Info("Restoring PostgreSQL database")

	// Find dump file
	dumpFile := filepath.Join(extractDir, fmt.Sprintf("postgres-%s.dump", metadata.BackupID))
	if _, err := os.Stat(dumpFile); err != nil {
		return fmt.Errorf("postgres dump not found: %w", err)
	}

	// Drop and recreate database
	if err := rm.recreateDatabase(ctx); err != nil {
		return fmt.Errorf("failed to recreate database: %w", err)
	}

	// Restore dump
	cmd := exec.CommandContext(ctx, "pg_restore",
		"-h", rm.config.DBHost,
		"-p", fmt.Sprintf("%d", rm.config.DBPort),
		"-U", rm.config.DBUser,
		"-d", rm.config.DBName,
		"--verbose",
		"--no-owner",
		"--no-acl",
		dumpFile,
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", rm.config.DBPassword))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_restore failed: %w, output: %s", err, string(output))
	}

	rm.logger.Info("PostgreSQL restore completed")

	return nil
}

// recreateDatabase drops and recreates the database
func (rm *RestoreManager) recreateDatabase(ctx context.Context) error {
	// Terminate connections
	terminateCmd := fmt.Sprintf(
		"SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '%s' AND pid <> pg_backend_pid();",
		rm.config.DBName,
	)

	cmd := exec.CommandContext(ctx, "psql",
		"-h", rm.config.DBHost,
		"-p", fmt.Sprintf("%d", rm.config.DBPort),
		"-U", rm.config.DBUser,
		"-d", "postgres",
		"-c", terminateCmd,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", rm.config.DBPassword))
	cmd.Run() // Ignore errors

	// Drop database
	dropCmd := exec.CommandContext(ctx, "dropdb",
		"-h", rm.config.DBHost,
		"-p", fmt.Sprintf("%d", rm.config.DBPort),
		"-U", rm.config.DBUser,
		rm.config.DBName,
	)
	dropCmd.Env = cmd.Env
	dropCmd.Run() // Ignore errors

	// Create database
	createCmd := exec.CommandContext(ctx, "createdb",
		"-h", rm.config.DBHost,
		"-p", fmt.Sprintf("%d", rm.config.DBPort),
		"-U", rm.config.DBUser,
		rm.config.DBName,
	)
	createCmd.Env = cmd.Env

	return createCmd.Run()
}

// restoreRedis restores Redis data
func (rm *RestoreManager) restoreRedis(ctx context.Context, extractDir string, metadata *BackupMetadata) error {
	rm.logger.Info("Restoring Redis data")

	rdbFile := filepath.Join(extractDir, fmt.Sprintf("redis-%s.rdb", metadata.BackupID))
	if _, err := os.Stat(rdbFile); err != nil {
		return fmt.Errorf("redis dump not found: %w", err)
	}

	// Stop Redis
	exec.Command("systemctl", "stop", "redis").Run()

	// Replace RDB file
	destRDB := "/var/lib/redis/dump.rdb"
	if err := copyFile(rdbFile, destRDB); err != nil {
		return err
	}

	// Start Redis
	return exec.Command("systemctl", "start", "redis").Run()
}

// restoreConfig restores configuration files
func (rm *RestoreManager) restoreConfig(ctx context.Context, extractDir string) error {
	configDir := filepath.Join(extractDir, "config")
	if _, err := os.Stat(configDir); err != nil {
		return nil // No config to restore
	}

	return copyDir(configDir, "/etc/trading-engine")
}

// restoreApplicationState restores application state
func (rm *RestoreManager) restoreApplicationState(ctx context.Context, extractDir string) error {
	stateDir := filepath.Join(extractDir, "state", "data")
	if _, err := os.Stat(stateDir); err != nil {
		return nil // No state to restore
	}

	return copyDir(stateDir, "./data")
}

// Helper functions

func (rm *RestoreManager) stopServices(ctx context.Context) error {
	return exec.CommandContext(ctx, "systemctl", "stop", "trading-engine").Run()
}

func (rm *RestoreManager) startServices(ctx context.Context) error {
	return exec.CommandContext(ctx, "systemctl", "start", "trading-engine").Run()
}

func (rm *RestoreManager) restartPostgreSQL(ctx context.Context) error {
	return exec.CommandContext(ctx, "systemctl", "restart", "postgresql").Run()
}

func (rm *RestoreManager) verifyRestore(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"

	cmd := exec.CommandContext(ctx, "psql",
		"-h", rm.config.DBHost,
		"-p", fmt.Sprintf("%d", rm.config.DBPort),
		"-U", rm.config.DBUser,
		"-d", rm.config.DBName,
		"-t", "-c", query,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", rm.config.DBPassword))

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var count int
	fmt.Sscanf(string(output), "%d", &count)

	return count, nil
}

func (rm *RestoreManager) findBaseBackup(ctx context.Context, targetTime time.Time) (string, error) {
	// This should search for the latest full backup before targetTime
	// Simplified implementation
	return "", fmt.Errorf("not implemented")
}

func (rm *RestoreManager) createRecoveryConfig(targetTime time.Time, walDir string) string {
	return fmt.Sprintf(`restore_command = 'cp %s/%%f %%p'
recovery_target_time = '%s'
recovery_target_action = 'promote'
`, walDir, targetTime.Format("2006-01-02 15:04:05"))
}

func (rm *RestoreManager) waitForRecovery(ctx context.Context) error {
	// Poll PostgreSQL until recovery completes
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			cmd := exec.CommandContext(ctx, "psql",
				"-h", rm.config.DBHost,
				"-p", fmt.Sprintf("%d", rm.config.DBPort),
				"-U", rm.config.DBUser,
				"-d", "postgres",
				"-t", "-c", "SELECT pg_is_in_recovery();",
			)
			cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", rm.config.DBPassword))

			output, err := cmd.Output()
			if err != nil {
				continue
			}

			if string(output) == " f\n" {
				return nil // Recovery complete
			}
		}
	}
}
