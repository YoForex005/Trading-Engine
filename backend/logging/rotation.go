package logging

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileLock is implemented in platform-specific files:
// - filelock_windows.go for Windows (LockFileEx)
// - filelock_unix.go for Unix/Linux (flock)

// RotatingFileWriter provides automatic log rotation based on size and time
type RotatingFileWriter struct {
	mu                 sync.Mutex
	filename           string
	file               *os.File
	maxSize            int64         // Maximum size in bytes before rotation
	maxAge             time.Duration // Maximum age before rotation
	maxBackups         int           // Maximum number of backup files to keep
	currentSize        int64
	createdAt          time.Time
	compressionEnabled bool
}

// RotationConfig configures log rotation
type RotationConfig struct {
	Filename           string
	MaxSizeMB          int           // Maximum size in MB
	MaxAge             time.Duration // Maximum age
	MaxBackups         int           // Maximum number of backups
	CompressionEnabled bool          // Enable gzip compression for rotated logs
}

// NewRotatingFileWriter creates a new rotating file writer
func NewRotatingFileWriter(config RotationConfig) (*RotatingFileWriter, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(config.Filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(config.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	maxSize := int64(config.MaxSizeMB) * 1024 * 1024
	if maxSize == 0 {
		maxSize = 100 * 1024 * 1024 // Default 100MB
	}

	maxAge := config.MaxAge
	if maxAge == 0 {
		maxAge = 7 * 24 * time.Hour // Default 7 days
	}

	maxBackups := config.MaxBackups
	if maxBackups == 0 {
		maxBackups = 10 // Default 10 backups
	}

	rfw := &RotatingFileWriter{
		filename:           config.Filename,
		file:               file,
		maxSize:            maxSize,
		maxAge:             maxAge,
		maxBackups:         maxBackups,
		currentSize:        stat.Size(),
		createdAt:          stat.ModTime(),
		compressionEnabled: config.CompressionEnabled,
	}

	// Start cleanup goroutine
	go rfw.cleanupOldLogs()

	return rfw, nil
}

// Write implements io.Writer
func (rfw *RotatingFileWriter) Write(p []byte) (n int, err error) {
	rfw.mu.Lock()
	defer rfw.mu.Unlock()

	// Check if rotation is needed
	if rfw.shouldRotate() {
		if err := rfw.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = rfw.file.Write(p)
	rfw.currentSize += int64(n)

	return n, err
}

// Close closes the file writer
func (rfw *RotatingFileWriter) Close() error {
	rfw.mu.Lock()
	defer rfw.mu.Unlock()

	if rfw.file != nil {
		return rfw.file.Close()
	}
	return nil
}

// shouldRotate checks if log rotation is needed
func (rfw *RotatingFileWriter) shouldRotate() bool {
	// Check size
	if rfw.currentSize >= rfw.maxSize {
		return true
	}

	// Check age
	if time.Since(rfw.createdAt) >= rfw.maxAge {
		return true
	}

	return false
}

// rotate performs the log rotation
func (rfw *RotatingFileWriter) rotate() error {
	// SECURITY: Acquire exclusive lock to prevent concurrent rotation
	// This prevents race conditions where multiple goroutines try to rotate simultaneously
	lock, err := NewFileLock(rfw.filename)
	if err != nil {
		return fmt.Errorf("failed to create rotation lock: %w", err)
	}

	if err := lock.Lock(); err != nil {
		return fmt.Errorf("failed to acquire rotation lock: %w", err)
	}
	defer lock.Unlock() // SAFETY: Always release lock, even on panic

	// Check if file still needs rotation (another goroutine may have rotated)
	info, err := os.Stat(rfw.filename)
	if err != nil {
		// File doesn't exist - already rotated or deleted
		if os.IsNotExist(err) {
			// Reopen the file
			file, reopenErr := os.OpenFile(rfw.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if reopenErr != nil {
				return reopenErr
			}
			rfw.file = file
			rfw.currentSize = 0
			rfw.createdAt = time.Now()
			return nil
		}
		return err
	}

	// Another goroutine already rotated - no action needed
	if info.Size() < rfw.maxSize {
		return nil
	}

	// Close current file
	if err := rfw.file.Close(); err != nil {
		return err
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s.%s", rfw.filename, timestamp)

	// Rename current file to backup
	if err := os.Rename(rfw.filename, backupName); err != nil {
		return err
	}

	// Compress backup if enabled (in background)
	if rfw.compressionEnabled {
		go compressFile(backupName)
	}

	// Open new file
	file, err := os.OpenFile(rfw.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	rfw.file = file
	rfw.currentSize = 0
	rfw.createdAt = time.Now()

	// Clean up old backups
	go rfw.cleanupOldBackups()

	return nil
}

// cleanupOldLogs removes logs older than maxAge
func (rfw *RotatingFileWriter) cleanupOldLogs() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		rfw.cleanupOldBackups()
	}
}

// cleanupOldBackups removes old backup files
func (rfw *RotatingFileWriter) cleanupOldBackups() {
	dir := filepath.Dir(rfw.filename)
	base := filepath.Base(rfw.filename)

	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// Collect backup files
	var backups []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > len(base) && file.Name()[:len(base)] == base {
			backups = append(backups, file)
		}
	}

	// Sort by modification time (oldest first)
	// Simple bubble sort for small lists
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			infoI, _ := backups[i].Info()
			infoJ, _ := backups[j].Info()
			if infoI != nil && infoJ != nil && infoI.ModTime().After(infoJ.ModTime()) {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	// Remove excess backups
	if len(backups) > rfw.maxBackups {
		for i := 0; i < len(backups)-rfw.maxBackups; i++ {
			os.Remove(filepath.Join(dir, backups[i].Name()))
		}
	}

	// Remove old backups based on age
	cutoff := time.Now().Add(-rfw.maxAge)
	for _, backup := range backups {
		info, err := backup.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(dir, backup.Name()))
		}
	}
}

// compressFile compresses a file using gzip asynchronously
// PERFORMANCE FIX #4: Async file compression
// Compression runs in background goroutine to prevent blocking log writes
func compressFile(filename string) {
	go func() {
		// PERFORMANCE: Non-blocking compression with panic recovery
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[LogRotation] PANIC during compression of %s: %v", filename, r)
			}
		}()

		// Actual compression with error handling
		if err := gzipCompressFile(filename); err != nil {
			log.Printf("[LogRotation] ERROR: Compression failed for %s: %v", filename, err)
			return
		}

		log.Printf("[LogRotation] INFO: Successfully compressed %s", filename)
	}()
}

// gzipCompressFile performs the actual gzip compression
func gzipCompressFile(srcPath string) error {
	// Open source file
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer src.Close()

	// Create compressed output
	dstPath := srcPath + ".gz"
	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create output: %w", err)
	}
	defer dst.Close()

	// Compress
	gzWriter := gzip.NewWriter(dst)
	defer gzWriter.Close()

	if _, err := io.Copy(gzWriter, src); err != nil {
		os.Remove(dstPath) // Clean up partial file on error
		return fmt.Errorf("compression failed: %w", err)
	}

	// Close gzip writer to flush
	if err := gzWriter.Close(); err != nil {
		os.Remove(dstPath)
		return fmt.Errorf("failed to flush compressed data: %w", err)
	}

	// Close destination file
	if err := dst.Close(); err != nil {
		os.Remove(dstPath)
		return fmt.Errorf("failed to close output file: %w", err)
	}

	// Remove original after successful compression
	if err := os.Remove(srcPath); err != nil {
		log.Printf("[LogRotation] WARNING: Failed to remove original file %s: %v", srcPath, err)
	}

	return nil
}

// MultiWriter writes to multiple writers simultaneously
type MultiWriter struct {
	writers []io.Writer
}

// NewMultiWriter creates a new multi-writer
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

// Write implements io.Writer
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}
	return len(p), nil
}

// Close closes all writers that implement io.Closer
func (mw *MultiWriter) Close() error {
	for _, w := range mw.writers {
		if closer, ok := w.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}
