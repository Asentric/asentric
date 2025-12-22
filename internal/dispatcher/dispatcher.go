package dispatcher

import (
	"context"

	"github.com/asentric/asentric/pkg/asentric"
)

type EngineDispatcher struct {
	Engine *asentric.Engine
	Sink   asentric.AlertSink
}

func NewEngineDispatcher(
	engine *asentric.Engine,
	sink asentric.AlertSink,
) *EngineDispatcher {
	return &EngineDispatcher{
		Engine: engine,
		Sink:   sink,
	}
}

func (d *EngineDispatcher) Dispatch(
	ctx context.Context,
	event asentric.Event,
) error {
	execCtx := &asentric.Context{
		Event: &event,
	}

	alerts, err := d.Engine.Evaluate(execCtx)
	if err != nil {
		return err
	}

	for _, alert := range alerts {
		if err := d.Sink.Emit(ctx, alert); err != nil {
			return err
		}
	}

	return nil
}
