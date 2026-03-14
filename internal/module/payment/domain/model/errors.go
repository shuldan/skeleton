package model

import "github.com/shuldan/errors"

var invoiceCode = errors.WithPrefix("INVOICE")

var (
	ErrInvoiceNotFound = invoiceCode("NOT_FOUND").
				Kind(errors.NotFound).
				New("invoice {{.id}} not found")

	ErrInvoiceAlreadyExists = invoiceCode("ALREADY_EXISTS").
				Kind(errors.Conflict).
				New("invoice for task {{.task_id}} already exists")
)
