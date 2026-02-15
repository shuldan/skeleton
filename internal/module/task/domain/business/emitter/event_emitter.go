package emitter

import (
	"context"

	"github.com/shuldan/skeleton/internal/module/task/domain/model"
)

// EventEmitter — интерфейс публикации доменного события.
type EventEmitter interface {
	Emit(ctx context.Context, task *model.Task)
}
