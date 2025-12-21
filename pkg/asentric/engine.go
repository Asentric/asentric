package asentric

import "context"

type Engine interface {
	RegisterRule(rule Rule) error
	Start(ctx context.Context) error
	Stop() error
}
