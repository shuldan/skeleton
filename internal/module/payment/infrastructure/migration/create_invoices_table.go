package migration

import "github.com/shuldan/migrator"

// CreateInvoicesTable возвращает миграцию создания таблицы invoices.
func CreateInvoicesTable() migrator.Migration {
	return migrator.CreateMigration(
		"20240102_001_create_invoices",
		"Create invoices table",
	).CreateTable("invoices",
		"id UUID PRIMARY KEY",
		"task_id VARCHAR(255) NOT NULL UNIQUE",
		"amount INTEGER NOT NULL",
		"status VARCHAR(50) NOT NULL DEFAULT 'created'",
		"version INTEGER NOT NULL DEFAULT 1",
		"created_at TIMESTAMP NOT NULL DEFAULT NOW()",
		"updated_at TIMESTAMP NOT NULL DEFAULT NOW()",
	).MustBuild()
}
