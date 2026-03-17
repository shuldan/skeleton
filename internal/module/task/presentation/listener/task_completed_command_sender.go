package listener

import (
	"context"

	"github.com/shuldan/commands"

	appcommand "github.com/shuldan/skeleton/internal/command"
	"github.com/shuldan/skeleton/internal/event"
	"github.com/shuldan/skeleton/internal/logger"
)

const defaultInvoiceAmount = 100

// TaskCompletedCommandSender отправляет CreateInvoice при завершении задачи.
type TaskCompletedCommandSender struct {
	client *commands.CommandClient
	log    logger.Logger
}

// NewTaskCompletedCommandSender создаёт listener.
func NewTaskCompletedCommandSender(
	client *commands.CommandClient,
	log logger.Logger,
) *TaskCompletedCommandSender {
	return &TaskCompletedCommandSender{
		client: client,
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

	future, err := commands.Send[*appcommand.InvoiceCreated](
		ctx, l.client, cmd,
	)
	if err != nil {
		l.log.Error("failed to send CreateInvoice command",
			"task_id", e.TaskID,
			"error", err,
		)
		return err
	}

	// Обрабатываем результат асинхронно.
	go func() {
		result, awaitErr := future.Await(context.Background())
		if awaitErr != nil {
			l.log.Error("CreateInvoice command failed",
				"task_id", e.TaskID,
				"error", awaitErr,
			)
			return
		}

		l.log.Info("invoice created via command bus",
			"invoice_id", result.InvoiceID,
			"task_id", result.TaskID,
			"amount", result.Amount,
			"status", result.Status,
		)
	}()

	l.log.Info("CreateInvoice command sent",
		"task_id", e.TaskID,
	)

	return nil
}
