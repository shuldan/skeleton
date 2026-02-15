package interactor

import (
	"context"

	"github.com/shuldan/skeleton/internal/module/task/domain/model"
	"github.com/shuldan/skeleton/internal/module/task/domain/persistence"
)

// ListTasksOutput — выходные данные списка задач.
type ListTasksOutput interface {
	AddTask() model.TaskPresenter
}

// ListTasksInteractor обрабатывает запрос списка задач.
type ListTasksInteractor struct {
	repo persistence.TaskRepository
}

// NewListTasksInteractor создаёт интерактор.
func NewListTasksInteractor(
	repo persistence.TaskRepository,
) *ListTasksInteractor {
	return &ListTasksInteractor{repo: repo}
}

// Handle выполняет use-case.
func (i *ListTasksInteractor) Handle(
	ctx context.Context,
	output ListTasksOutput,
) error {
	tasks, err := i.repo.FindAll(ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		task.RepresentTo(output.AddTask())
	}

	return nil
}
