package asentric

import "context"

type EventSource interface {
	Start(ctx context.Context) (<-chan Event, error)
}
