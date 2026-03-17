package model

// InvoiceSnapshot — плоское представление Invoice для персистентности.
type InvoiceSnapshot struct {
	ID      string
	TaskID  string
	Amount  int
	Status  string
	Version int
}

// Restore восстанавливает агрегат из снимка.
func (s *InvoiceSnapshot) Restore() *Invoice {
	return &Invoice{
		id:      NewInvoiceID(s.ID),
		taskID:  s.TaskID,
		amount:  s.Amount,
		status:  statusFromString(s.Status),
		version: s.Version,
	}
}
