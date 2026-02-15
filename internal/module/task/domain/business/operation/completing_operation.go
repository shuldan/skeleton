package operation

import (
	"context"

	"github.com/shuldan/skeleton/internal/module/task/domain/model"
)

// CompletingOperation — интерфейс операции завершения задачи.
type CompletingOperation interface {
	Complete(ctx context.Context, id model.TaskID) (*model.Task, error)
}
