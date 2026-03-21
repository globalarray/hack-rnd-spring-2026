package domain

import "errors"

var (
	ErrNotFound          = errors.New("data not found")
	ErrConflict          = errors.New("conflict")
	ErrTimeLimitExceeded = errors.New("time limit exceeded")
)
