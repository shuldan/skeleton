package interactor

import (
	"context"

	"github.com/shuldan/skeleton/internal/module/task/domain/business/emitter"
	"github.com/shuldan/skeleton/internal/module/task/domain/business/operation"
	"github.com/shuldan/skeleton/internal/module/task/domain/model"
)

// CreateTaskInput — входные данные создания задачи.
type CreateTaskInput interface {
	GetTitle() string
	GetDescription() string
}

// CreateTaskOutput — выходные данные создания задачи.
type CreateTaskOutput interface {
	model.TaskPresenter
}

// CreateTaskInteractor оркестрирует создание задачи.
type CreateTaskInteractor struct {
	creatingOp operation.CreatingOperation
	emitter    emitter.EventEmitter
}

// NewCreateTaskInteractor создаёт интерактор.
func NewCreateTaskInteractor(
	creatingOp operation.CreatingOperation,
	emitter emitter.EventEmitter,
) *CreateTaskInteractor {
	return &CreateTaskInteractor{
		creatingOp: creatingOp,
		emitter:    emitter,
	}
}

// Handle выполняет use-case.
func (i *CreateTaskInteractor) Handle(
	ctx context.Context,
	input CreateTaskInput,
	output CreateTaskOutput,
) error {
	title, err := model.NewTitle(input.GetTitle())
	if err != nil {
		return err
	}

	task, err := i.creatingOp.Create(
		ctx, title, input.GetDescription(),
	)
	if err != nil {
		return err
	}

	i.emitter.Emit(ctx, task.ReleaseEvents())
	task.RepresentTo(output)

	return nil
}
