package asentric

type Client struct {
	engine   *Engine
	registry *RuleRegistry
}

func NewClient(cfg *Config) *Client {
	registry := NewRuleRegistry()
	engine := NewEngine(registry, cfg)

	return &Client{
		engine:   engine,
		registry: registry,
	}
}

func (c *Client) RegisterRule(rule *Rule) error {
	return c.registry.Register(rule)
}

func (c *Client) Process(ctx *Context) ([]*Alert, error) {
	return c.engine.Evaluate(ctx)
}
