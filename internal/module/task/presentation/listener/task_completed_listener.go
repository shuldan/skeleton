package listener

import (
	"context"

	"github.com/shuldan/skeleton/internal/event"
	"github.com/shuldan/skeleton/internal/logger"
)

// TaskCompletedListener обрабатывает TaskCompleted in-process.
type TaskCompletedListener struct {
	log logger.Logger
}

// NewTaskCompletedListener создаёт listener.
func NewTaskCompletedListener(
	log logger.Logger,
) *TaskCompletedListener {
	return &TaskCompletedListener{log: log}
}

// Handle вызывается Dispatcher-ом при получении TaskCompleted.
func (l *TaskCompletedListener) Handle(
	_ context.Context, e *event.TaskCompleted,
) error {
	l.log.Info("task completed event received",
		"task_id", e.TaskID,
	)

	return nil
}
