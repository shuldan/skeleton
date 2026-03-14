package model

// InvoicePresenter — интерфейс представления счёта.
type InvoicePresenter interface {
	SetID(id string) InvoicePresenter
	SetTaskID(taskID string) InvoicePresenter
	SetAmount(amount int) InvoicePresenter
	SetStatus(status string) InvoicePresenter
	SetVersion(version int) InvoicePresenter
}
