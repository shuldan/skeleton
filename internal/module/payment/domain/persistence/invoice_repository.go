package persistence

import (
	"context"

	"github.com/shuldan/skeleton/internal/module/payment/domain/model"
)

// InvoiceRepository — интерфейс репозитория счетов.
type InvoiceRepository interface {
	FindByID(ctx context.Context, id model.InvoiceID) (*model.Invoice, error)
	FindByTaskID(ctx context.Context, taskID string) (*model.Invoice, error)
	FindAll(ctx context.Context) ([]*model.Invoice, error)
	Save(ctx context.Context, invoice *model.Invoice) error
}
