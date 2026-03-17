package model

// Invoice — агрегат счёта на оплату.
type Invoice struct {
	id      InvoiceID
	taskID  string
	amount  int
	status  status
	version int
}

// NewInvoice создаёт счёт в статусе created.
func NewInvoice(
	id InvoiceID, taskID string, amount int,
) *Invoice {
	return &Invoice{
		id:      id,
		taskID:  taskID,
		amount:  amount,
		status:  statusCreated,
		version: 1,
	}
}

func (i *Invoice) ID() InvoiceID  { return i.id }
func (i *Invoice) TaskID() string { return i.taskID }
func (i *Invoice) Amount() int    { return i.amount }

// RepresentTo передаёт данные агрегата через presenter.
func (i *Invoice) RepresentTo(p InvoicePresenter) {
	p.SetID(i.id.String()).
		SetTaskID(i.taskID).
		SetAmount(i.amount).
		SetStatus(i.status.String()).
		SetVersion(i.version)
}

// Snapshot возвращает снимок агрегата.
func (i *Invoice) Snapshot() InvoiceSnapshot {
	return InvoiceSnapshot{
		ID:      i.id.String(),
		TaskID:  i.taskID,
		Amount:  i.amount,
		Status:  i.status.String(),
		Version: i.version,
	}
}
