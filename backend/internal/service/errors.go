package service

import "errors"

var (
	ErrInvalidInput          = errors.New("invalid input")
	ErrMissingToken          = errors.New("missing token")
	ErrInvalidAdminToken     = errors.New("invalid admin token")
	ErrForbidden             = errors.New("forbidden")
	ErrIPBanned              = errors.New("ip banned")
	ErrNotFound              = errors.New("not found")
	ErrConflict              = errors.New("conflict")
	ErrDependencyUnavailable = errors.New("dependency unavailable")
)

// UserError wraps a domain error with a message that is safe to return to API clients.
type UserError struct {
	Err     error
	Message string
}

func (e UserError) Error() string {
	if e.Err == nil {
		return e.Message
	}
	if e.Message == "" {
		return e.Err.Error()
	}
	return e.Err.Error() + ": " + e.Message
}

func (e UserError) Unwrap() error {
	return e.Err
}

func WithUserMessage(err error, message string) error {
	return UserError{Err: err, Message: message}
}

func UserMessage(err error, fallback string) string {
	var userErr UserError
	if errors.As(err, &userErr) && userErr.Message != "" {
		return userErr.Message
	}
	return fallback
}
