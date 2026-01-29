//go:build !windows
// +build !windows

package logging

import (
	"fmt"
	"os"
	"syscall"
)

// FileLock provides Unix/Linux file locking
type FileLock struct {
	path string
	file *os.File
}

// NewFileLock creates a lock file for exclusive access control
func NewFileLock(basePath string) (*FileLock, error) {
	lockPath := basePath + ".lock"

	// Create or open lock file
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create lock file: %w", err)
	}

	return &FileLock{path: lockPath, file: f}, nil
}

// Lock acquires exclusive lock (blocking)
func (fl *FileLock) Lock() error {
	err := syscall.Flock(int(fl.file.Fd()), syscall.LOCK_EX)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	return nil
}

// Unlock releases lock and removes lock file
func (fl *FileLock) Unlock() error {
	unlockErr := syscall.Flock(int(fl.file.Fd()), syscall.LOCK_UN)

	fl.file.Close()
	os.Remove(fl.path) // Clean up lock file

	return unlockErr
}
