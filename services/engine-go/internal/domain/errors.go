package domain

import "errors"

var (
	ErrNotFound          = errors.New("data not found")
	ErrConflict          = errors.New("conflict")
	ErrShareLinkUsed     = errors.New("share link already used")
	ErrTimeLimitExceeded = errors.New("time limit exceeded")
)
