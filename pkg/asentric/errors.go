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

	// Config Validation Details
	ErrSourceTypeRequired = errors.New("source.type is required")
	ErrSourceURLRequired  = errors.New("source.url is required for websocket type")
	ErrSinkTypeRequired   = errors.New("sink.type is required")
	ErrSinkURLRequired    = errors.New("sink.url is required for webhook type")
	ErrChainIDRequired    = errors.New("chain.id is required")
	ErrChainNameRequired  = errors.New("chain.name is required")
	ErrChainRPCWSRequired = errors.New("chain.rpcWs is required for websocket source")

	// Config Loading
	ErrConfigReadFailed  = errors.New("failed to read config file")
	ErrConfigParseFailed = errors.New("failed to parse config file")

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
