package port

import "context"

// NotificationPort — порт отправки уведомлений.
type NotificationPort interface {
	Send(ctx context.Context, taskID, message string) error
}
