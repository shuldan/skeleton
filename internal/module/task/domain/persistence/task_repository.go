package persistence

import (
	"context"

	"github.com/shuldan/skeleton/internal/module/task/domain/model"
)

// TaskRepository — интерфейс репозитория задач.
type TaskRepository interface {
	FindByID(ctx context.Context, id model.TaskID) (*model.Task, error)
	FindAll(ctx context.Context) ([]*model.Task, error)
	Save(ctx context.Context, task *model.Task) error
}
