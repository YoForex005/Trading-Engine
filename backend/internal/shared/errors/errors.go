package errors

import "fmt"

// NotFoundError indicates a resource was not found
type NotFoundError struct {
	Resource string // "account", "position", "order"
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

func NewNotFound(resource, id string) error {
	return &NotFoundError{Resource: resource, ID: id}
}

// ValidationError indicates invalid input
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

func NewValidation(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

// InsufficientFundsError indicates account lacks balance
type InsufficientFundsError struct {
	AccountID string
	Required  string // decimal as string
	Available string
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("insufficient funds in account %s: required %s, available %s",
		e.AccountID, e.Required, e.Available)
}

func NewInsufficientFunds(accountID, required, available string) error {
	return &InsufficientFundsError{
		AccountID: accountID,
		Required:  required,
		Available: available,
	}
}
