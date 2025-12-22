package asentric

import "errors"

var (
	ErrInvalidContext = errors.New("invalid context")
	ErrInvalidRule    = errors.New("invalid rule")
	ErrDuplicateRule  = errors.New("duplicate rule id")
	ErrInvalidEvent   = errors.New("invalid event")
	ErrInvalidConfig  = errors.New("invalid configuration")
	ErrNoDispatcher   = errors.New("dispatcher is not set")
	ErrAlreadyRunning = errors.New("runtime is already running")
	ErrRulePanic      = errors.New("rule panic")
)
