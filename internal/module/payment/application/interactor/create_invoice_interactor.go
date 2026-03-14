package interactor

import (
	"context"

	"github.com/google/uuid"

	"github.com/shuldan/skeleton/internal/module/payment/domain/model"
	"github.com/shuldan/skeleton/internal/module/payment/domain/persistence"
)

// CreateInvoiceInput — входные данные создания счёта.
type CreateInvoiceInput interface {
	GetTaskID() string
	GetAmount() int
}

// CreateInvoiceOutput — выходные данные создания счёта.
type CreateInvoiceOutput interface {
	model.InvoicePresenter
}

// CreateInvoiceInteractor оркестрирует создание счёта.
type CreateInvoiceInteractor struct {
	repo persistence.InvoiceRepository
}

// NewCreateInvoiceInteractor создаёт интерактор.
func NewCreateInvoiceInteractor(
	repo persistence.InvoiceRepository,
) *CreateInvoiceInteractor {
	return &CreateInvoiceInteractor{repo: repo}
}

// Handle выполняет use-case.
func (i *CreateInvoiceInteractor) Handle(
	ctx context.Context,
	input CreateInvoiceInput,
	output CreateInvoiceOutput,
) error {
	existing, _ := i.repo.FindByTaskID(ctx, input.GetTaskID())
	if existing != nil {
		existing.RepresentTo(output)
		return nil
	}

	id := model.NewInvoiceID(uuid.New().String())
	invoice := model.NewInvoice(id, input.GetTaskID(), input.GetAmount())

	if err := i.repo.Save(ctx, invoice); err != nil {
		return err
	}

	invoice.RepresentTo(output)

	return nil
}
