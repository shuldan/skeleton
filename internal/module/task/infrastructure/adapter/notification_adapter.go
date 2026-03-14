package adapter

import (
	"context"

	"github.com/shuldan/skeleton/internal/logger"
	"github.com/shuldan/skeleton/internal/module/task/application/port"
)

// LoggingNotificationAdapter — заглушка, пишет в лог.
type LoggingNotificationAdapter struct {
	log logger.Logger
}

// NewLoggingNotificationAdapter создаёт адаптер.
func NewLoggingNotificationAdapter(
	log logger.Logger,
) port.NotificationPort {
	return &LoggingNotificationAdapter{log: log}
}

func (a *LoggingNotificationAdapter) Send(
	_ context.Context, taskID, message string,
) error {
	a.log.Info("notification sent",
		"task_id", taskID,
		"message", message,
	)

	return nil
}
