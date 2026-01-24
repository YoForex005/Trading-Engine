//go:build windows
// +build windows

package logging

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = kernel32.NewProc("LockFileEx")
	procUnlockFileEx = kernel32.NewProc("UnlockFileEx")
)

const (
	LOCKFILE_EXCLUSIVE_LOCK   = 0x00000002
	LOCKFILE_FAIL_IMMEDIATELY = 0x00000001
)

// FileLock provides Windows file locking
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
	// Windows LockFileEx parameters
	var overlapped syscall.Overlapped

	// LockFileEx(handle, flags, reserved, nNumberOfBytesToLockLow, nNumberOfBytesToLockHigh, lpOverlapped)
	r1, _, err := procLockFileEx.Call(
		uintptr(fl.file.Fd()),
		uintptr(LOCKFILE_EXCLUSIVE_LOCK), // Exclusive, blocking
		uintptr(0),                       // Reserved
		uintptr(1),                       // Lock 1 byte
		uintptr(0),                       // High-order 32 bits
		uintptr(unsafe.Pointer(&overlapped)),
	)

	if r1 == 0 {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	return nil
}

// Unlock releases lock and removes lock file
func (fl *FileLock) Unlock() error {
	var overlapped syscall.Overlapped

	// UnlockFileEx(handle, reserved, nNumberOfBytesToUnlockLow, nNumberOfBytesToUnlockHigh, lpOverlapped)
	r1, _, err := procUnlockFileEx.Call(
		uintptr(fl.file.Fd()),
		uintptr(0), // Reserved
		uintptr(1), // Unlock 1 byte
		uintptr(0), // High-order 32 bits
		uintptr(unsafe.Pointer(&overlapped)),
	)

	fl.file.Close()
	os.Remove(fl.path) // Clean up lock file

	if r1 == 0 {
		return fmt.Errorf("failed to unlock: %w", err)
	}
	return nil
}
