package model

// InvoiceID — идентификатор счёта.
type InvoiceID struct {
	value string
}

// NewInvoiceID создаёт InvoiceID из строки.
func NewInvoiceID(raw string) InvoiceID {
	return InvoiceID{value: raw}
}

func (id InvoiceID) String() string {
	return id.value
}
