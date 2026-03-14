package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shuldan/repository"

	"github.com/shuldan/skeleton/internal/module/payment/domain/model"
	domainpersistence "github.com/shuldan/skeleton/internal/module/payment/domain/persistence"
)

var invoiceTable = repository.Table{
	Name:          "invoices",
	PrimaryKey:    []string{"id"},
	Columns:       []string{"id", "task_id", "amount", "status", "version"},
	VersionColumn: "version",
	CreatedAt:     "created_at",
	UpdatedAt:     "updated_at",
}

type invoiceRepository struct {
	repo *repository.Repository[*model.Invoice]
	db   *sql.DB
}

// NewInvoiceRepository создаёт SQL-реализацию репозитория.
func NewInvoiceRepository(
	db *sql.DB,
) domainpersistence.InvoiceRepository {
	r := repository.New(
		db,
		repository.Postgres(),
		repository.Simple(repository.SimpleConfig[*model.Invoice]{
			Table:  invoiceTable,
			Scan:   scanInvoice,
			Values: invoiceValues,
		}),
	)

	return &invoiceRepository{repo: r, db: db}
}

func (r *invoiceRepository) FindByID(
	ctx context.Context, id model.InvoiceID,
) (*model.Invoice, error) {
	invoice, err := r.repo.Find(ctx, id.String())
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, model.ErrInvoiceNotFound.
				WithDetail("id", id.String())
		}

		return nil, err
	}

	return invoice, nil
}

func (r *invoiceRepository) FindByTaskID(
	ctx context.Context, taskID string,
) (*model.Invoice, error) {
	invoices, err := r.repo.FindBy(ctx, repository.Eq("task_id", taskID))
	if err != nil {
		return nil, err
	}

	if len(invoices) == 0 {
		return nil, nil
	}

	return invoices[0], nil
}

func (r *invoiceRepository) FindAll(
	ctx context.Context,
) ([]*model.Invoice, error) {
	return r.repo.FindBy(ctx, nil)
}

func (r *invoiceRepository) Save(
	ctx context.Context, invoice *model.Invoice,
) error {
	return r.repo.Save(ctx, invoice)
}

func scanInvoice(sc repository.Scanner) (*model.Invoice, error) {
	var s model.InvoiceSnapshot
	if err := sc.Scan(
		&s.ID, &s.TaskID, &s.Amount,
		&s.Status, &s.Version,
	); err != nil {
		return nil, err
	}

	return s.Restore(), nil
}

func invoiceValues(inv *model.Invoice) []any {
	s := inv.Snapshot()

	return []any{
		s.ID, s.TaskID, s.Amount,
		s.Status, s.Version,
	}
}
