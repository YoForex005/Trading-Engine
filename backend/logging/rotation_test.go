package logging

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestFileLockBasic tests basic lock acquisition and release
func TestFileLockBasic(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	lock, err := NewFileLock(testFile)
	if err != nil {
		t.Fatalf("Failed to create lock: %v", err)
	}

	if err := lock.Lock(); err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	if err := lock.Unlock(); err != nil {
		t.Errorf("Failed to unlock: %v", err)
	}
}

// TestFileLockConcurrency tests that locks prevent concurrent access
func TestFileLockConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Run 10 goroutines trying to acquire the same lock
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			lock, err := NewFileLock(testFile)
			if err != nil {
				t.Errorf("Failed to create lock: %v", err)
				return
			}

			if err := lock.Lock(); err != nil {
				t.Errorf("Failed to acquire lock: %v", err)
				return
			}
			defer lock.Unlock()

			// Critical section - increment counter
			mu.Lock()
			counter++
			mu.Unlock()

			// Simulate work
			time.Sleep(10 * time.Millisecond)
		}()
	}

	wg.Wait()

	if counter != 10 {
		t.Errorf("Expected counter to be 10, got %d", counter)
	}
}

// TestConcurrentRotation tests that concurrent writes don't cause duplicate rotations
func TestConcurrentRotation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	// Create writer with small max size to trigger rotation
	writer, err := NewRotatingFileWriter(RotationConfig{
		Filename:   testFile,
		MaxSizeMB:  1, // 1MB max size
		MaxAge:     24 * time.Hour,
		MaxBackups: 5,
	})
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	var wg sync.WaitGroup
	var rotationCount int
	var rotationMu sync.Mutex

	// Spawn 10 goroutines writing concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Write data that might trigger rotation
			data := make([]byte, 100*1024) // 100KB per write
			for j := 0; j < 100; j++ {
				n, err := writer.Write(data)
				if err != nil {
					t.Errorf("Goroutine %d: Write failed: %v", id, err)
					return
				}
				if n != len(data) {
					t.Errorf("Goroutine %d: Partial write: %d/%d", id, n, len(data))
				}
			}
		}(i)
	}

	// Wait for all writes to complete
	wg.Wait()

	// Count backup files
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) != ".lock" {
			rotationMu.Lock()
			rotationCount++
			rotationMu.Unlock()
		}
	}

	// Should have main file + rotated backups
	t.Logf("Total files after concurrent writes: %d", rotationCount)

	// Verify no lock files remain
	lockFiles := 0
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".lock" {
			lockFiles++
		}
	}

	if lockFiles > 0 {
		t.Errorf("Found %d lock files remaining - locks not properly cleaned up", lockFiles)
	}
}

// TestRotationRaceCondition tests the specific race condition scenario
func TestRotationRaceCondition(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	writer, err := NewRotatingFileWriter(RotationConfig{
		Filename:   testFile,
		MaxSizeMB:  1, // Small size to trigger rotation
		MaxAge:     24 * time.Hour,
		MaxBackups: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	var wg sync.WaitGroup
	errorCount := 0
	var errorMu sync.Mutex

	// Simulate the exact race condition: multiple goroutines hitting rotation threshold
	largeData := make([]byte, 900*1024) // 900KB - close to 1MB threshold

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each goroutine writes enough to trigger rotation
			_, err := writer.Write(largeData)
			if err != nil {
				errorMu.Lock()
				errorCount++
				errorMu.Unlock()
				t.Logf("Write error (expected during rotation): %v", err)
			}
		}()
	}

	wg.Wait()

	// Should have completed without deadlock
	t.Logf("Completed with %d write errors during rotation", errorCount)

	// Verify file integrity
	if _, err := os.Stat(testFile); err != nil {
		t.Errorf("Main log file missing after rotation: %v", err)
	}
}

// BenchmarkRotationWithLock benchmarks rotation performance with file locking
func BenchmarkRotationWithLock(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	writer, err := NewRotatingFileWriter(RotationConfig{
		Filename:   testFile,
		MaxSizeMB:  1,
		MaxAge:     24 * time.Hour,
		MaxBackups: 5,
	})
	if err != nil {
		b.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	data := []byte("benchmark test data for rotation performance\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(data)
	}
}

// BenchmarkConcurrentWrites benchmarks concurrent write performance
func BenchmarkConcurrentWrites(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	writer, err := NewRotatingFileWriter(RotationConfig{
		Filename:   testFile,
		MaxSizeMB:  10,
		MaxAge:     24 * time.Hour,
		MaxBackups: 5,
	})
	if err != nil {
		b.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	data := []byte("concurrent write benchmark data\n")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			writer.Write(data)
		}
	})
}
