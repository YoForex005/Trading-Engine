package database

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// HandleQueryError wraps database query errors with appropriate context
// Returns nil if err is nil
// Returns formatted error with resource context if query fails
func HandleQueryError(err error, resource, id string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s not found: %s", resource, id)
	}

	return fmt.Errorf("failed to get %s %s: %w", resource, id, err)
}

// HandleInsertError wraps database insert errors with appropriate context
func HandleInsertError(err error, resource string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to create %s: %w", resource, err)
}

// HandleUpdateError wraps database update errors with appropriate context
func HandleUpdateError(err error, resource, id string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to update %s %s: %w", resource, id, err)
}

// HandleDeleteError wraps database delete errors with appropriate context
func HandleDeleteError(err error, resource, id string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to delete %s %s: %w", resource, id, err)
}

// ScanError wraps scan errors with appropriate context
func ScanError(err error, resource string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to scan %s: %w", resource, err)
}
