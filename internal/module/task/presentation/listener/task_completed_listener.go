package listener

import (
	"context"

	event "github.com/shuldan/skeleton/internal/event"
)

// TaskCompletedListener обрабатывает TaskCompleted in-process.
type TaskCompletedListener struct {
	log func(msg string, args ...any)
}

// NewTaskCompletedListener создаёт listener.
func NewTaskCompletedListener(
	log func(msg string, args ...any),
) *TaskCompletedListener {
	return &TaskCompletedListener{log: log}
}

// Handle вызывается Dispatcher-ом при получении TaskCompleted.
func (l *TaskCompletedListener) Handle(
	_ context.Context, e event.TaskCompleted,
) error {
	l.log("task completed event received",
		"task_id", e.TaskID,
	)

	return nil
}
