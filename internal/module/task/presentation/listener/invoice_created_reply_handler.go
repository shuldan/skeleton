package listener

import (
	"context"
	"encoding/json"

	"github.com/shuldan/commands"

	appcommand "github.com/shuldan/skeleton/internal/command"
	"github.com/shuldan/skeleton/internal/logger"

	"github.com/shuldan/framework/commandbus"
)

// DeserializeInvoiceCreated десериализует результат команды CreateInvoice.
func DeserializeInvoiceCreated(
	payload []byte, _ *commandbus.ResultEnvelope,
) (commands.Result, error) {
	var result appcommand.InvoiceCreated
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// NewInvoiceCreatedReplyHandler возвращает callback
// для обработки результата CreateInvoice.
func NewInvoiceCreatedReplyHandler(
	log logger.Logger,
) commandbus.ResultCallbackFunc {
	return func(
		_ context.Context,
		result commands.Result,
		err error,
	) error {
		if err != nil {
			log.Error("CreateInvoice command failed",
				"error", err,
			)

			return nil
		}

		invoiceResult, ok := result.(*appcommand.InvoiceCreated)
		if !ok {
			log.Error("unexpected result type for CreateInvoice")
			return nil
		}

		log.Info("invoice created via command bus",
			"invoice_id", invoiceResult.InvoiceID,
			"task_id", invoiceResult.TaskID,
			"amount", invoiceResult.Amount,
			"status", invoiceResult.Status,
		)

		return nil
	}
}
