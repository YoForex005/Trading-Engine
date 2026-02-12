package auth

import "errors"

var (
	// ErrTOTPNotEnabled is returned when 2FA is not enabled for a user
	ErrTOTPNotEnabled = errors.New("2FA is not enabled for this user")

	// ErrInvalidTOTPCode is returned when the TOTP code is invalid
	ErrInvalidTOTPCode = errors.New("invalid 2FA code")

	// ErrTOTPAlreadyEnabled is returned when trying to enable 2FA that's already enabled
	ErrTOTPAlreadyEnabled = errors.New("2FA is already enabled for this user")

	// Err2FARequired is returned when 2FA verification is required
	Err2FARequired = errors.New("2FA verification required")
)
