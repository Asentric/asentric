package asentric

type Config struct {
	FailFast bool
}

func DefaultConfig() *Config {
	return &Config{
		FailFast: false,
	}
}
