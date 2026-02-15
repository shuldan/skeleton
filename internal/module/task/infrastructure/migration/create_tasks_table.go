package migration

import "github.com/shuldan/migrator"

// CreateTasksTable возвращает миграцию создания таблицы tasks.
func CreateTasksTable() migrator.Migration {
	return migrator.CreateMigration(
		"20240101_001_create_tasks",
		"Create tasks table",
	).CreateTable("tasks",
		"id UUID PRIMARY KEY",
		"title VARCHAR(255) NOT NULL",
		"description TEXT NOT NULL DEFAULT ''",
		"status VARCHAR(50) NOT NULL DEFAULT 'draft'",
		"version INTEGER NOT NULL DEFAULT 1",
		"created_at TIMESTAMP NOT NULL DEFAULT NOW()",
		"updated_at TIMESTAMP NOT NULL DEFAULT NOW()",
	).MustBuild()
}
