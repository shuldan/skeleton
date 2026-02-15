package operation

import (
	"context"

	domainop "github.com/shuldan/skeleton/internal/module/task/domain/business/operation"
	"github.com/shuldan/skeleton/internal/module/task/domain/model"
	"github.com/shuldan/skeleton/internal/module/task/domain/persistence"
)

// CompletingOperation — реализация операции завершения задачи.
type CompletingOperation struct {
	repo persistence.TaskRepository
}

// NewCompletingOperation создаёт операцию.
func NewCompletingOperation(
	repo persistence.TaskRepository,
) domainop.CompletingOperation {
	return &CompletingOperation{repo: repo}
}

func (o *CompletingOperation) Complete(
	ctx context.Context, id model.TaskID,
) (*model.Task, error) {
	task, err := o.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := task.Complete(); err != nil {
		return nil, err
	}

	if err := o.repo.Save(ctx, task); err != nil {
		return nil, err
	}

	return task, nil
}
