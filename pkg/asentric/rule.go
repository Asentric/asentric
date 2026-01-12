package asentric

// Rule defines the interface for security detection rules.
//
// Rules are evaluated by the Engine against a Context and may produce
// an Alert if conditions are met.
//
// Priority/ordering is determined by Severity:
// CRITICAL > HIGH > MEDIUM > LOW > INFO
type Rule interface {
	// Name returns the unique identifier for this rule
	Name() string

	// Severity returns the severity level of alerts produced by this rule.
	// This also serves as the priority indicator for rule evaluation order.
	Severity() Severity

	// Evaluate processes the given context and returns an alert if triggered.
	// Returns nil if rule conditions are not met.
	Evaluate(ctx Context) (*Alert, error)
}

