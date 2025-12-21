package asentric

type EngineConfig struct {
	Chain  ChainConfig
	Stream StreamConfig
}

type ChainConfig struct {
	RPC string
}

type StreamConfig struct {
	Driver string // redis, memory
}
