package asentric

import "errors"

var (
	ErrInvalidRule   = errors.New("invalid rule")
	ErrDuplicateRule = errors.New("duplicate rule id")
	ErrInvalidEvent  = errors.New("invalid event")
)
