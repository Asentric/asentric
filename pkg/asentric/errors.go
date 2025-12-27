package asentric

import "errors"

var (
	// Runtime errors
	ErrSourceRequired     = errors.New("EventSource is required")
	ErrSinkRequired       = errors.New("AlertSink is required")
	ErrEngineRequired     = errors.New("Engine is required")
	ErrDispatcherRequired = errors.New("Dispatcher is required")

	// Config errors
	ErrInvalidConfig          = errors.New("invalid configuration")
	ErrInvalidChainConfig     = errors.New("invalid chain configuration")
	ErrInvalidRedisConfig     = errors.New("invalid redis configuration")
	ErrInvalidWebhookConfig   = errors.New("invalid webhook configuration")
	ErrInvalidEngineConfig    = errors.New("invalid engine configuration")
	ErrInvalidWebSocketConfig = errors.New("invalid web socket configuration")

	// Websocket errors
	ErrFailedToConnectWebSocket = errors.New("failed to connect to websocket")

	// Engine errors
	ErrInvalidContext = errors.New("invalid context")
	ErrInvalidRule    = errors.New("invalid rule")
	ErrDuplicateRule  = errors.New("duplicate rule id")
	ErrInvalidEvent   = errors.New("invalid event")
	ErrNoDispatcher   = errors.New("dispatcher is not set")
	ErrAlreadyRunning = errors.New("runtime is already running")
	ErrRulePanic      = errors.New("rule panic")
)
