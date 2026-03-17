package emitter

import (
	"context"

	"github.com/shuldan/events"

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
	ctx context.Context, domainEvents []any,
) {
	publishCtx := context.WithoutCancel(ctx)

	for _, ev := range domainEvents {
		if err := e.dispatcher.Publish(publishCtx, ev); err != nil {
			e.log.Error("failed to emit event",
				"event", ev,
				"error", err,
			)
		}
	}
}
