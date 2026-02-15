package interactor

import (
	"context"

	"github.com/google/uuid"

	"github.com/shuldan/skeleton/internal/module/task/domain/model"
	"github.com/shuldan/skeleton/internal/module/task/domain/persistence"
)

// GetTaskInput — входные данные получения задачи.
type GetTaskInput interface {
	GetTaskID() string
}

// GetTaskOutput — выходные данные получения задачи.
type GetTaskOutput interface {
	model.TaskPresenter
}

// GetTaskInteractor обрабатывает запрос одной задачи.
type GetTaskInteractor struct {
	repo persistence.TaskRepository
}

// NewGetTaskInteractor создаёт интерактор.
func NewGetTaskInteractor(
	repo persistence.TaskRepository,
) *GetTaskInteractor {
	return &GetTaskInteractor{repo: repo}
}

// Handle выполняет use-case.
func (i *GetTaskInteractor) Handle(
	ctx context.Context,
	input GetTaskInput,
	output GetTaskOutput,
) error {
	taskID := model.TaskID(uuid.MustParse(input.GetTaskID()))

	task, err := i.repo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}

	task.RepresentTo(output)

	return nil
}
