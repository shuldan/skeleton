package listener

import (
	"context"

	"github.com/shuldan/framework/commandbus"

	appcommand "github.com/shuldan/skeleton/internal/command"
	"github.com/shuldan/skeleton/internal/event"
	"github.com/shuldan/skeleton/internal/logger"
)

const defaultInvoiceAmount = 100

// TaskCompletedCommandSender отправляет CreateInvoice при заверш��нии задачи.
type TaskCompletedCommandSender struct {
	sender *commandbus.CommandSender
	log    logger.Logger
}

// NewTaskCompletedCommandSender создаёт listener.
func NewTaskCompletedCommandSender(
	sender *commandbus.CommandSender,
	log logger.Logger,
) *TaskCompletedCommandSender {
	return &TaskCompletedCommandSender{
		sender: sender,
		log:    log,
	}
}

// Handle вызывается Dispatcher-ом при получении TaskCompleted.
func (l *TaskCompletedCommandSender) Handle(
	ctx context.Context, e *event.TaskCompleted,
) error {
	cmd := &appcommand.CreateInvoice{
		TaskID: e.TaskID,
		Amount: defaultInvoiceAmount,
	}

	if err := l.sender.Send(ctx, cmd); err != nil {
		l.log.Error("failed to send CreateInvoice command",
			"task_id", e.TaskID,
			"error", err,
		)

		return err
	}

	l.log.Info("CreateInvoice command sent",
		"task_id", e.TaskID,
	)

	return nil
}
