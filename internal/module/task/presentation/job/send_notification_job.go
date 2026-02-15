package job

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shuldan/queue"

	"github.com/shuldan/skeleton/internal/module/task/application/port"
)

type taskCompletedPayload struct {
	TaskID string `json:"task_id"`
	Event  string `json:"event"`
}

// SendNotificationJob обрабатывает сообщения
// из topic "task.completed", отправляет уведомления.
type SendNotificationJob struct {
	broker   queue.Broker
	notifier port.NotificationPort
}

// NewSendNotificationJob создаёт job.
func NewSendNotificationJob(
	broker queue.Broker,
	notifier port.NotificationPort,
) *SendNotificationJob {
	return &SendNotificationJob{
		broker:   broker,
		notifier: notifier,
	}
}

// Run запускает consumer. Блокируется до отмены ctx.
func (j *SendNotificationJob) Run(
	ctx context.Context,
) error {
	return j.broker.Consume(
		ctx,
		"task.completed",
		func(data []byte) error {
			return j.process(ctx, data)
		},
	)
}

func (j *SendNotificationJob) process(
	ctx context.Context, data []byte,
) error {
	var payload taskCompletedPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return j.notifier.Send(
		ctx,
		payload.TaskID,
		"Task "+payload.TaskID+" has been completed",
	)
}
