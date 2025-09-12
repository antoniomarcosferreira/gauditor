package gauditor

import "errors"

// ErrInvalidEvent is returned when required fields are missing.
var ErrInvalidEvent = errors.New("invalid event: missing required fields")
