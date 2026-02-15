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
	completingOp     operation.CompletingOperation
	completedEmitter emitter.EventEmitter
}

// NewCompleteTaskInteractor создаёт интерактор.
func NewCompleteTaskInteractor(
	completingOp operation.CompletingOperation,
	completedEmitter emitter.EventEmitter,
) *CompleteTaskInteractor {
	return &CompleteTaskInteractor{
		completingOp:     completingOp,
		completedEmitter: completedEmitter,
	}
}

// Handle выполняет use-case.
func (i *CompleteTaskInteractor) Handle(
	ctx context.Context,
	input CompleteTaskInput,
	output CompleteTaskOutput,
) error {
	taskID := model.TaskID(uuid.MustParse(input.GetTaskID()))

	task, err := i.completingOp.Complete(ctx, taskID)
	if err != nil {
		return err
	}

	task.RepresentTo(output)
	i.completedEmitter.Emit(ctx, task)

	return nil
}
