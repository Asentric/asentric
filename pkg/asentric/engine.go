package asentric

// Engine orchestrates rule evaluation over an immutable Context.
// Engine is deterministic, single-threaded, and side-effect free.
//
// Engine is NOT safe for concurrent use.
// Engine MUST NOT perform I/O or manage concurrency.
//
// Engine is a CONCRETE TYPE, not an interface.
// Engine is NOT an extension point and should not be extended or replaced.
type Engine struct {
	// internal state (rule registry, config)
	rules []Rule
}

// NewEngine creates a new Engine instance.
// Engine maintains internal state (rule registry, config), but does not maintain
// per-event or cross-event execution state. Each Evaluate() call is independent.
func NewEngine() *Engine {
	return &Engine{
		rules: make([]Rule, 0),
	}
}

// RegisterRule registers a rule into the engine.
// Rules are executed in registration order (deterministic).
func (e *Engine) RegisterRule(rule Rule) error {
	if rule == nil {
		return ErrInvalidRule
	}
	e.rules = append(e.rules, rule)
	return nil
}

// Evaluate evaluates all registered rules against the given Context.
//
// Rules:
// - ctx MUST NOT be nil
// - Engine MUST NOT mutate Context
// - Engine MUST NOT perform I/O
// - Rules are executed sequentially
// - Rule execution order is deterministic (registration order)
//
// Returns:
// - zero or more Alerts
// - error on execution failure
func (e *Engine) Evaluate(ctx Context) ([]*Alert, error) {
	if ctx == nil {
		return nil, ErrInvalidContext
	}

	var alerts []*Alert

	for _, rule := range e.rules {
		alert, err := e.executeRule(rule, ctx)
		if err != nil {
			return nil, err
		}
		if alert != nil {
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// executeRule executes a single rule against the context.
// Recovers from panics and returns ErrRulePanic if panic occurs.
func (e *Engine) executeRule(rule Rule, ctx Context) (alert *Alert, err error) {
	defer func() {
		if recover() != nil {
			alert = nil
			err = ErrRulePanic
		}
	}()

	return rule.Evaluate(ctx)
}
