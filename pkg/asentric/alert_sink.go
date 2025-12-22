package asentric

import "context"

type AlertSink interface {
	Emit(ctx context.Context, alert *Alert) error
}
