package operation

import (
	"context"

	domainop "github.com/shuldan/skeleton/internal/module/task/domain/business/operation"
	"github.com/shuldan/skeleton/internal/module/task/domain/model"
	"github.com/shuldan/skeleton/internal/module/task/domain/persistence"
)

// CreatingOperation — реализация операции создания задачи.
type CreatingOperation struct {
	repo persistence.TaskRepository
}

// NewCreatingOperation создаёт операцию.
func NewCreatingOperation(
	repo persistence.TaskRepository,
) domainop.CreatingOperation {
	return &CreatingOperation{repo: repo}
}

func (o *CreatingOperation) Create(
	ctx context.Context,
	title model.Title,
	description string,
) (*model.Task, error) {
	task := model.NewTask(title, description)

	if err := o.repo.Save(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}
