package asentric

type AlertSink interface {
	Emit(ctx Context, alert *Alert) error
}
