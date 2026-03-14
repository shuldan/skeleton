package interactor

import (
	"context"

	"github.com/shuldan/skeleton/internal/module/payment/domain/model"
	"github.com/shuldan/skeleton/internal/module/payment/domain/persistence"
)

// ListInvoicesOutput — выходные данные списка счетов.
type ListInvoicesOutput interface {
	AddInvoice() model.InvoicePresenter
}

// ListInvoicesInteractor обрабатывает запрос списка счетов.
type ListInvoicesInteractor struct {
	repo persistence.InvoiceRepository
}

// NewListInvoicesInteractor создаёт интерактор.
func NewListInvoicesInteractor(
	repo persistence.InvoiceRepository,
) *ListInvoicesInteractor {
	return &ListInvoicesInteractor{repo: repo}
}

// Handle выполняет use-case.
func (i *ListInvoicesInteractor) Handle(
	ctx context.Context,
	output ListInvoicesOutput,
) error {
	invoices, err := i.repo.FindAll(ctx)
	if err != nil {
		return err
	}

	for _, inv := range invoices {
		inv.RepresentTo(output.AddInvoice())
	}

	return nil
}
