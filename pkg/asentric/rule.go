package asentric

type Rule struct {
	ID          string
	Name        string
	Description string
	Severity    Severity

	// Matcher adalah fungsi pure
	Match func(ctx *Context) bool

	Tags    []string
	Version string
}
