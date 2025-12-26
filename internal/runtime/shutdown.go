package runtime

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

type Shutdown struct {
	signals []os.Signal
}

func NewShutdown() *Shutdown {
	return &Shutdown{
		signals: []os.Signal{
			syscall.SIGINT,
			syscall.SIGTERM,
		},
	}
}

// Wait blocks until either:
// - the context is done (runtime already stopped), or
// - an external signal (SIGINT/SIGTERM) is received.
// Caller is responsible for invoking Stop() or cancelling the context.
func (s *Shutdown) Wait(ctx context.Context) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, s.signals...)
	defer signal.Stop(ch)

	select {
	case <-ctx.Done():
		return
	case <-ch:
		return
	}
}
