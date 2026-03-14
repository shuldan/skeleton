package commandhandler

import (
	"context"
	"encoding/json"

	"github.com/shuldan/commands"

	appcommand "github.com/shuldan/skeleton/internal/command"
	"github.com/shuldan/skeleton/internal/module/payment/application/interactor"
	"github.com/shuldan/skeleton/internal/module/payment/domain/model"

	"github.com/shuldan/framework/commandbus"
)

// createInvoiceInput адаптирует команду к интерфейсу интерактора.
type createInvoiceInput struct {
	cmd *appcommand.CreateInvoice
}

func (i *createInvoiceInput) GetTaskID() string { return i.cmd.TaskID }
func (i *createInvoiceInput) GetAmount() int    { return i.cmd.Amount }

// createInvoiceOutput реализует InvoicePresenter и собирает результат.
type createInvoiceOutput struct {
	ID      string `json:"invoice_id"`
	TaskID  string `json:"task_id"`
	Amount  int    `json:"amount"`
	Status  string `json:"status"`
	Version int    `json:"-"`
}

func (o *createInvoiceOutput) SetID(v string) model.InvoicePresenter {
	o.ID = v
	return o
}

func (o *createInvoiceOutput) SetTaskID(v string) model.InvoicePresenter {
	o.TaskID = v
	return o
}

func (o *createInvoiceOutput) SetAmount(v int) model.InvoicePresenter {
	o.Amount = v
	return o
}

func (o *createInvoiceOutput) SetStatus(v string) model.InvoicePresenter {
	o.Status = v
	return o
}

func (o *createInvoiceOutput) SetVersion(v int) model.InvoicePresenter {
	o.Version = v
	return o
}

// DeserializeCreateInvoice десериализует payload в команду.
func DeserializeCreateInvoice(
	payload []byte, _ *commandbus.CommandEnvelope,
) (commands.Command, error) {
	var cmd appcommand.CreateInvoice
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return nil, err
	}

	return &cmd, nil
}

// HandleCreateInvoice обрабатывает команду и возвращает результат.
func HandleCreateInvoice(
	inter *interactor.CreateInvoiceInteractor,
) commandbus.CommandHandlerFunc {
	return func(
		ctx context.Context, cmd commands.Command,
	) (commands.Result, error) {
		createCmd, ok := cmd.(*appcommand.CreateInvoice)
		if !ok {
			return nil, commands.ErrHandlerNotFound
		}

		output := &createInvoiceOutput{}
		input := &createInvoiceInput{cmd: createCmd}

		if err := inter.Handle(ctx, input, output); err != nil {
			return nil, err
		}

		return &appcommand.InvoiceCreated{
			InvoiceID: output.ID,
			TaskID:    output.TaskID,
			Amount:    output.Amount,
			Status:    output.Status,
		}, nil
	}
}
