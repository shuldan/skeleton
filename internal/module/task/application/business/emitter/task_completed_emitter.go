package emitter

import (
	"context"
	"log/slog"

	"github.com/shuldan/events"

	domainemitter "github.com/shuldan/skeleton/internal/module/task/domain/business/emitter"
	"github.com/shuldan/skeleton/internal/module/task/domain/model"

	event "github.com/shuldan/skeleton/internal/event"
)

// TaskCompletedEmitter публикует событие TaskCompleted.
type TaskCompletedEmitter struct {
	dispatcher *events.Dispatcher
}

// NewTaskCompletedEmitter создаёт emitter.
func NewTaskCompletedEmitter(
	dispatcher *events.Dispatcher,
) domainemitter.EventEmitter {
	return &TaskCompletedEmitter{dispatcher: dispatcher}
}

func (e *TaskCompletedEmitter) Emit(
	ctx context.Context, task *model.Task,
) {
	ev := &taskEvent{}
	task.RepresentTo(ev)

	if err := e.dispatcher.Publish(
		context.WithoutCancel(ctx),
		event.TaskCompleted{
			BaseEvent: events.NewBaseEvent(
				"TaskCompleted", ev.taskID,
			),
			TaskID: ev.taskID,
		},
	); err != nil {
		slog.ErrorContext(ctx,
			"failed to emit TaskCompleted",
			"task_id", ev.taskID,
			"error", err,
		)
	}
}
