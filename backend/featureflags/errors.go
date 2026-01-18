package featureflags

import "errors"

// Common errors
var (
	ErrFlagNotFound          = errors.New("flag not found")
	ErrFlagExists            = errors.New("flag already exists")
	ErrExperimentNotFound    = errors.New("experiment not found")
	ErrExperimentExists      = errors.New("experiment already exists")
	ErrRolloutNotFound       = errors.New("rollout not found")
	ErrRolloutExists         = errors.New("rollout already exists")
	ErrInvalidRolloutStatus  = errors.New("invalid rollout status")
	ErrFunnelExists          = errors.New("funnel already exists")
	ErrSegmentExists         = errors.New("segment already exists")
)

// ErrInvalidFlag creates an invalid flag error
func ErrInvalidFlag(msg string) error {
	return errors.New("invalid flag: " + msg)
}

// ErrInvalidExperiment creates an invalid experiment error
func ErrInvalidExperiment(msg string) error {
	return errors.New("invalid experiment: " + msg)
}

// ErrInvalidRollout creates an invalid rollout error
func ErrInvalidRollout(msg string) error {
	return errors.New("invalid rollout: " + msg)
}
