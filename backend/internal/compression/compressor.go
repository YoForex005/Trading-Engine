package compression

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics tracks compression statistics
type Metrics struct {
	FilesCompressed int64
	BytesOriginal   int64
	BytesCompressed int64
	ErrorCount      int64
	LastCompression time.Time
	LastError       string
	mu              sync.RWMutex
}

// Config holds compression configuration
type Config struct {
	Enabled        bool
	DataDir        string
	MaxAgeSeconds  int64  // Files older than this will be compressed
	Schedule       string // Cron-like schedule or duration string
	MaxConcurrency int
}

// Compressor handles tick data compression
type Compressor struct {
	config  Config
	metrics *Metrics
	done    chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
}

// NewCompressor creates a new compressor instance
func NewCompressor(config Config) *Compressor {
	if config.MaxAgeSeconds == 0 {
		config.MaxAgeSeconds = 7 * 24 * 3600 // Default 7 days
	}
	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = 4
	}
	if config.DataDir == "" {
		config.DataDir = "backend/data/ticks"
	}

	return &Compressor{
		config:  config,
		metrics: &Metrics{},
		done:    make(chan struct{}),
	}
}

// GetMetrics returns current compression metrics (copy without mutex)
func (c *Compressor) GetMetrics() Metrics {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()
	return Metrics{
		FilesCompressed: c.metrics.FilesCompressed,
		BytesOriginal:   c.metrics.BytesOriginal,
		BytesCompressed: c.metrics.BytesCompressed,
		ErrorCount:      c.metrics.ErrorCount,
		LastCompression: c.metrics.LastCompression,
		LastError:       c.metrics.LastError,
	}
}

// IsEnabled returns whether compression is enabled
func (c *Compressor) IsEnabled() bool {
	return c.config.Enabled
}

// GetConfig returns a copy of the compression configuration
func (c *Compressor) GetConfig() Config {
	return c.config
}

// recordError records an error in metrics
func (c *Compressor) recordError(err error) {
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()
	atomic.AddInt64(&c.metrics.ErrorCount, 1)
	c.metrics.LastError = err.Error()
}

// Start begins the compression scheduler
func (c *Compressor) Start() {
	if !c.config.Enabled {
		log.Println("[Compressor] Compression disabled by configuration")
		return
	}

	c.wg.Add(1)
	go c.run()
	log.Println("[Compressor] Compression scheduler started")
}

// Stop gracefully stops the compressor
func (c *Compressor) Stop() {
	log.Println("[Compressor] Stopping compression scheduler")
	close(c.done)
	c.wg.Wait()
	log.Println("[Compressor] Compression scheduler stopped")
}

// TriggerCompression manually triggers a compression scan
func (c *Compressor) TriggerCompression() {
	c.compressOldFiles()
}

// run executes the compression scheduler loop
func (c *Compressor) run() {
	defer c.wg.Done()

	// Parse schedule duration
	scheduleDuration := 7 * 24 * time.Hour // Default weekly
	if c.config.Schedule != "" {
		parsed, err := time.ParseDuration(c.config.Schedule)
		if err != nil {
			log.Printf("[Compressor] Invalid schedule duration '%s', using default weekly: %v", c.config.Schedule, err)
		} else {
			scheduleDuration = parsed
		}
	}

	ticker := time.NewTicker(scheduleDuration)
	defer ticker.Stop()

	// Run first compression immediately
	log.Println("[Compressor] Running initial compression scan")
	c.compressOldFiles()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			log.Println("[Compressor] Running scheduled compression scan")
			c.compressOldFiles()
		}
	}
}

// compressOldFiles scans and compresses files older than threshold
func (c *Compressor) compressOldFiles() {
	start := time.Now()
	log.Printf("[Compressor] Starting compression scan (threshold: %d seconds)", c.config.MaxAgeSeconds)

	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, c.config.MaxConcurrency)
	var wg sync.WaitGroup

	// Walk through all .json files in data directory
	err := filepath.Walk(c.config.DataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("[Compressor] Error accessing path %s: %v", path, err)
			return nil // Continue walking
		}

		// Skip directories and non-json files
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		// Skip already compressed files
		if _, err := os.Stat(path + ".gz"); err == nil {
			return nil // File already compressed
		}

		// Check if file is old enough
		age := time.Since(info.ModTime()).Seconds()
		if age < float64(c.config.MaxAgeSeconds) {
			return nil // File is too new
		}

		// Queue compression task
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			c.compressFile(filePath)
		}(path)

		return nil
	})

	if err != nil {
		log.Printf("[Compressor] Error walking directory: %v", err)
		c.recordError(fmt.Errorf("directory walk failed: %w", err))
	}

	// Wait for all compressions to complete
	wg.Wait()

	// Update metrics
	c.metrics.mu.Lock()
	c.metrics.LastCompression = time.Now()
	c.metrics.mu.Unlock()

	duration := time.Since(start)
	metrics := c.GetMetrics()
	log.Printf("[Compressor] Compression scan completed in %v: %d files (%.2f MB → %.2f MB saved: %.1f%%)",
		duration,
		metrics.FilesCompressed,
		float64(metrics.BytesOriginal)/(1024*1024),
		float64(metrics.BytesCompressed)/(1024*1024),
		100*float64(metrics.BytesOriginal-metrics.BytesCompressed)/float64(metrics.BytesOriginal+1), // +1 to avoid division by zero
	)
}

// compressFile compresses a single file atomically
// PERFORMANCE FIX #5: Compression error handling with atomic operations
func (c *Compressor) compressFile(sourceFile string) {
	// Get file info before compression
	srcInfo, err := os.Stat(sourceFile)
	if err != nil {
		log.Printf("[Compressor] Cannot stat file %s: %v", sourceFile, err)
		c.recordError(err)
		return
	}

	srcSize := srcInfo.Size()

	// Create temp file in same directory for atomicity
	tempFile := sourceFile + ".tmp.gz"
	destFile := sourceFile + ".gz"

	// SAFETY: Track error state to ensure cleanup on any error path
	var compressionErr error
	defer func() {
		// PERFORMANCE FIX #5: Clean up partial/corrupt files on failure
		if compressionErr != nil {
			if _, statErr := os.Stat(tempFile); statErr == nil {
				os.Remove(tempFile) // Remove partial file
				log.Printf("[Compressor] Cleaned up partial file %s after error", tempFile)
			}
		}
	}()

	// Open source file
	src, err := os.Open(sourceFile)
	if err != nil {
		compressionErr = err
		log.Printf("[Compressor] Cannot open source file %s: %v", sourceFile, err)
		c.recordError(err)
		return
	}
	defer src.Close()

	// Create temp destination file
	dst, err := os.Create(tempFile)
	if err != nil {
		compressionErr = err
		log.Printf("[Compressor] Cannot create temp file %s: %v", tempFile, err)
		c.recordError(err)
		return
	}
	defer dst.Close()

	// Create gzip writer with compression level 6 (balanced)
	gzipWriter, err := gzip.NewWriterLevel(dst, gzip.DefaultCompression)
	if err != nil {
		compressionErr = err
		log.Printf("[Compressor] Cannot create gzip writer: %v", err)
		c.recordError(err)
		return
	}

	// Copy and compress
	bytesWritten, err := io.Copy(gzipWriter, src)
	if err != nil {
		compressionErr = err
		log.Printf("[Compressor] Error during compression of %s: %v", sourceFile, err)
		c.recordError(err)
		gzipWriter.Close()
		return
	}
	log.Printf("[Compressor] Compressed %d bytes from %s", bytesWritten, sourceFile)

	// Close gzip writer to flush
	if err := gzipWriter.Close(); err != nil {
		compressionErr = err
		log.Printf("[Compressor] Error closing gzip writer: %v", err)
		c.recordError(err)
		return
	}

	if err := dst.Close(); err != nil {
		compressionErr = err
		log.Printf("[Compressor] Error closing destination file: %v", err)
		c.recordError(err)
		return
	}

	// Get compressed file size
	dstInfo, err := os.Stat(tempFile)
	if err != nil {
		compressionErr = err
		log.Printf("[Compressor] Cannot stat compressed file: %v", err)
		c.recordError(err)
		return
	}

	dstSize := dstInfo.Size()

	// PERFORMANCE FIX #5: Atomic rename only if compression succeeded
	if err := os.Rename(tempFile, destFile); err != nil {
		compressionErr = err
		log.Printf("[Compressor] Failed to finalize %s: %v", destFile, err)
		c.recordError(err)
		return
	}

	// Delete original file only after successful compression and rename
	if err := os.Remove(sourceFile); err != nil {
		log.Printf("[Compressor] Warning: Could not delete original file %s: %v", sourceFile, err)
		// Don't fail here, file is already compressed
	}

	// Update metrics (only on success)
	atomic.AddInt64(&c.metrics.FilesCompressed, 1)
	atomic.AddInt64(&c.metrics.BytesOriginal, srcSize)
	atomic.AddInt64(&c.metrics.BytesCompressed, dstSize)

	compression := 100 * (1 - float64(dstSize)/float64(srcSize))
	log.Printf("[Compressor] Compressed %s: %.2f KB → %.2f KB (%.1f%% reduction)",
		filepath.Base(sourceFile),
		float64(srcSize)/1024,
		float64(dstSize)/1024,
		compression,
	)
}

// CompressFile manually compresses a specific file
func (c *Compressor) CompressFile(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("cannot stat file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file")
	}

	if filepath.Ext(filePath) != ".json" {
		return fmt.Errorf("file must have .json extension")
	}

	// Check if already compressed
	if _, err := os.Stat(filePath + ".gz"); err == nil {
		return fmt.Errorf("file is already compressed")
	}

	c.compressFile(filePath)
	return nil
}
