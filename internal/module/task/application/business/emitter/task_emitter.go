package emitter

import (
	"context"

	"github.com/shuldan/events"

	appevent "github.com/shuldan/skeleton/internal/event"
	"github.com/shuldan/skeleton/internal/logger"
	domainemitter "github.com/shuldan/skeleton/internal/module/task/domain/business/emitter"
)

// TaskEmitter публикует доменные события задачи.
type TaskEmitter struct {
	dispatcher *events.Dispatcher
	log        logger.Logger
}

// NewTaskEmitter создаёт emitter.
func NewTaskEmitter(
	dispatcher *events.Dispatcher,
	log logger.Logger,
) domainemitter.EventEmitter {
	return &TaskEmitter{
		dispatcher: dispatcher,
		log:        log,
	}
}

func (e *TaskEmitter) Emit(
	ctx context.Context, domainEvents []appevent.Event,
) {
	publishCtx := context.WithoutCancel(ctx)

	for _, ev := range domainEvents {
		dispatchable, ok := ev.(events.Event)
		if !ok {
			e.log.Error("event does not implement events.Event",
				"event_name", ev.EventName(),
			)
			continue
		}

		if err := e.dispatcher.Publish(publishCtx, dispatchable); err != nil {
			e.log.Error("failed to emit event",
				"event_name", ev.EventName(),
				"error", err,
			)
		}
	}
}
