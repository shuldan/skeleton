package command

import "github.com/shuldan/commands"

// CreateInvoice — команда создания счёта при завершении задачи.
type CreateInvoice struct {
	TaskID string `json:"task_id"`
	Amount int    `json:"amount"`
}

func (c *CreateInvoice) CommandName() string    { return "CreateInvoice" }
func (c *CreateInvoice) IdempotencyKey() string { return "invoice-" + c.TaskID }

// Compile-time check.
var _ commands.Command = (*CreateInvoice)(nil)

// InvoiceCreated — результат успешного создания счёта.
type InvoiceCreated struct {
	InvoiceID string `json:"invoice_id"`
	TaskID    string `json:"task_id"`
	Amount    int    `json:"amount"`
	Status    string `json:"status"`
}

func (r *InvoiceCreated) ResultName() string { return "InvoiceCreated" }
