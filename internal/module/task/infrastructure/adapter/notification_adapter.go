package adapter

import (
	"context"

	"github.com/shuldan/skeleton/internal/module/task/application/port"
)

// LoggingNotificationAdapter — заглушка, пишет в лог.
type LoggingNotificationAdapter struct {
	log func(msg string, args ...any)
}

// NewLoggingNotificationAdapter создаёт адаптер.
func NewLoggingNotificationAdapter(
	log func(msg string, args ...any),
) port.NotificationPort {
	return &LoggingNotificationAdapter{log: log}
}

func (a *LoggingNotificationAdapter) Send(
	_ context.Context, taskID, message string,
) error {
	a.log("notification sent",
		"task_id", taskID,
		"message", message,
	)

	return nil
}
