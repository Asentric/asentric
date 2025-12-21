package asentric

type Rule interface {
	Name() string
	Description() string

	Match(ctx Context) bool
	Execute(ctx Context) (*Alert, error)
}
