package emitter

import (
	"context"
	"log/slog"

	"github.com/shuldan/events"

	domainemitter "github.com/shuldan/skeleton/internal/module/task/domain/business/emitter"
	"github.com/shuldan/skeleton/internal/module/task/domain/model"

	event "github.com/shuldan/skeleton/internal/event"
)

// TaskCreatedEmitter публикует событие TaskCreated.
type TaskCreatedEmitter struct {
	dispatcher *events.Dispatcher
}

// NewTaskCreatedEmitter создаёт emitter.
func NewTaskCreatedEmitter(
	dispatcher *events.Dispatcher,
) domainemitter.EventEmitter {
	return &TaskCreatedEmitter{dispatcher: dispatcher}
}

func (e *TaskCreatedEmitter) Emit(
	ctx context.Context, task *model.Task,
) {
	ev := &taskEvent{}
	task.RepresentTo(ev)

	if err := e.dispatcher.Publish(
		context.WithoutCancel(ctx),
		event.TaskCreated{
			BaseEvent: events.NewBaseEvent(
				"TaskCreated", ev.taskID,
			),
			TaskID: ev.taskID,
			Title:  ev.title,
		},
	); err != nil {
		slog.ErrorContext(ctx,
			"failed to emit TaskCreated",
			"task_id", ev.taskID,
			"error", err,
		)
	}
}
