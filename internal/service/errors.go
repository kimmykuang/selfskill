package service

import "errors"

// Sentinel errors so callers (handlers, CLI) can map to HTTP status / exit codes.
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrConflict      = errors.New("conflict")
)
