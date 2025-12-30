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

	// Dispatcher
	ErrDispatcherEngine         = errors.New("dispatcher : no engine provided")
	ErrDispatcherAlertSink      = errors.New("dispatcher : no alert sink provided")
	ErrDispatcherContextBuilder = errors.New("dispatcher : no context builder provided")
	ErrDispatcherABIRegistry    = errors.New("dispatcher : no ABI registry provided")
)
