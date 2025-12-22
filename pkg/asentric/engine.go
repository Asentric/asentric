package asentric

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Engine mengevaluasi Context terhadap RuleRegistry.
// Engine bersifat stateless dan single-call safe.
type Engine struct {
	registry *RuleRegistry
	config   *Config
}

func NewEngine(registry *RuleRegistry, cfg *Config) *Engine {
	return &Engine{
		registry: registry,
		config:   cfg,
	}
}

// Evaluate mengeksekusi semua rule terhadap Context.
// - Tidak memodifikasi Context
// - Tidak mengelola concurrency
// - Tidak melakukan side-effect selain menghasilkan Alert
func (e *Engine) Evaluate(ctx *Context) ([]*Alert, error) {
	if ctx == nil || ctx.Event == nil {
		return nil, ErrInvalidEvent
	}

	var alerts []*Alert

	for _, rule := range e.registry.List() {
		if rule.Match == nil {
			continue
		}

		if rule.Match(ctx) {
			alerts = append(alerts, &Alert{
				ID:        generateID(),
				RuleID:    rule.ID,
				Severity:  rule.Severity,
				Summary:   rule.Name,
				Evidence:  ctx.Event,
				Timestamp: time.Now(),
			})

			if e.config.FailFast {
				break
			}
		}
	}

	return alerts, nil
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
