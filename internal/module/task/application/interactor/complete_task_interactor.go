package interactor

import (
	"context"

	"github.com/google/uuid"

	"github.com/shuldan/skeleton/internal/module/task/domain/business/emitter"
	"github.com/shuldan/skeleton/internal/module/task/domain/business/operation"
	"github.com/shuldan/skeleton/internal/module/task/domain/model"
)

// CompleteTaskInput — входные данные завершения задачи.
type CompleteTaskInput interface {
	GetTaskID() string
}

// CompleteTaskOutput — выходные данные завершения задачи.
type CompleteTaskOutput interface {
	model.TaskPresenter
}

// CompleteTaskInteractor оркестрирует завершение задачи.
type CompleteTaskInteractor struct {
	completingOp operation.CompletingOperation
	emitter      emitter.EventEmitter
}

// NewCompleteTaskInteractor создаёт интерактор.
func NewCompleteTaskInteractor(
	completingOp operation.CompletingOperation,
	emitter emitter.EventEmitter,
) *CompleteTaskInteractor {
	return &CompleteTaskInteractor{
		completingOp: completingOp,
		emitter:      emitter,
	}
}

// Handle выполняет use-case.
func (i *CompleteTaskInteractor) Handle(
	ctx context.Context,
	input CompleteTaskInput,
	output CompleteTaskOutput,
) error {
	taskID := model.NewTaskID(uuid.MustParse(input.GetTaskID()).String())

	task, err := i.completingOp.Complete(ctx, taskID)
	if err != nil {
		return err
	}

	i.emitter.Emit(ctx, task.ReleaseEvents())
	task.RepresentTo(output)

	return nil
}
