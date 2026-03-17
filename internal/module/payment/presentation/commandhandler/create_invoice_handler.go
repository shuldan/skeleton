package commandhandler

import (
	"context"

	"github.com/shuldan/commands"

	appcommand "github.com/shuldan/skeleton/internal/command"
	"github.com/shuldan/skeleton/internal/module/payment/application/interactor"
	"github.com/shuldan/skeleton/internal/module/payment/domain/model"
)

// createInvoiceInput адаптирует команду к интерфейсу интерактора.
type createInvoiceInput struct {
	cmd *appcommand.CreateInvoice
}

func (i *createInvoiceInput) GetTaskID() string { return i.cmd.TaskID }
func (i *createInvoiceInput) GetAmount() int    { return i.cmd.Amount }

// createInvoiceOutput реализует InvoicePresenter и собирает результат.
type createInvoiceOutput struct {
	ID      string
	TaskID  string
	Amount  int
	Status  string
	Version int
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

// CreateInvoiceHandler обрабатывает команду CreateInvoice.
type CreateInvoiceHandler struct {
	inter *interactor.CreateInvoiceInteractor
}

// NewCreateInvoiceHandler создаёт обработчик.
func NewCreateInvoiceHandler(
	inter *interactor.CreateInvoiceInteractor,
) *CreateInvoiceHandler {
	return &CreateInvoiceHandler{inter: inter}
}

// Handle обрабатывает команду и отправляет результат через reply.
func (h *CreateInvoiceHandler) Handle(
	ctx context.Context,
	cmd *appcommand.CreateInvoice,
	reply commands.ReplySender,
) error {
	output := &createInvoiceOutput{}
	input := &createInvoiceInput{cmd: cmd}

	if err := h.inter.Handle(ctx, input, output); err != nil {
		return reply.SendError(ctx, err)
	}

	return reply.Send(ctx, &appcommand.InvoiceCreated{
		InvoiceID: output.ID,
		TaskID:    output.TaskID,
		Amount:    output.Amount,
		Status:    output.Status,
	})
}
