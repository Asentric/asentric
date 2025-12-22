package asentric

import "errors"

var (
	ErrInvalidRule    = errors.New("invalid rule")
	ErrDuplicateRule  = errors.New("duplicate rule id")
	ErrInvalidEvent   = errors.New("invalid event")
	ErrNoDispatcher   = errors.New("dispatcher is not set")
	ErrAlreadyRunning = errors.New("runtime is already running")
)
