package asentric

type Rule interface {
	// Name returns the unique identifier for this rule
	Name() string

	// Evaluate processes the given context and returns alerts if triggered
	Evaluate(ctx Context) (*Alert, error)
}
