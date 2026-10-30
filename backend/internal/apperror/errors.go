// Package apperror holds the sentinel errors shared by all backend layers.
package apperror

import "errors"

var (
	ErrNotFound       = errors.New("user not found")
	ErrDuplicateEmail = errors.New("a user with this email already exists")
	ErrInvalidID      = errors.New("invalid user id")
	ErrValidation     = errors.New("validation failed")
	ErrUnavailable    = errors.New("service unavailable")
)

// ValidationError carries per-field messages and matches ErrValidation.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return ErrValidation.Error() }

func (e *ValidationError) Is(target error) bool { return target == ErrValidation }
