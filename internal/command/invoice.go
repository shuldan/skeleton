package command

// CreateInvoice — команда создания счёта при завершении задачи.
type CreateInvoice struct {
	TaskID string `json:"task_id"`
	Amount int    `json:"amount"`
}

func (c *CreateInvoice) CommandName() string { return "CreateInvoice" }

// InvoiceCreated — результат успешного создания счёта.
type InvoiceCreated struct {
	InvoiceID string `json:"invoice_id"`
	TaskID    string `json:"task_id"`
	Amount    int    `json:"amount"`
	Status    string `json:"status"`
}

func (r *InvoiceCreated) ResultName() string { return "InvoiceCreated" }
