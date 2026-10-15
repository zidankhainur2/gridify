package domain

import "errors"

var (
	ErrNotFound             = errors.New("resource not found")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrForbidden            = errors.New("forbidden: access denied")
	ErrConflict             = errors.New("conflict: resource already exists or is locked")
	ErrBadRequest           = errors.New("bad request")
	ErrInvalidCredentials   = errors.New("invalid email/username or password")
	ErrTournamentLocked     = errors.New("tournament is not in draft status")
	ErrDownstreamMatchPlayed = errors.New("downstream match has already been played")
	ErrInvalidInput         = errors.New("invalid input")
)

type FieldValidationError struct {
	Field   string
	Message string
}

type ValidationError struct {
	Errors []FieldValidationError
}

func (v *ValidationError) Error() string {
	return "validation failed"
}
