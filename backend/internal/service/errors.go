package service

import "errors"

var (
	ErrInvalidInput          = errors.New("invalid input")
	ErrMissingToken          = errors.New("missing token")
	ErrInvalidAdminToken     = errors.New("invalid admin token")
	ErrForbidden             = errors.New("forbidden")
	ErrNotFound              = errors.New("not found")
	ErrConflict              = errors.New("conflict")
	ErrDependencyUnavailable = errors.New("dependency unavailable")
)
