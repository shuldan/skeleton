package operation

import (
	"context"

	"github.com/shuldan/skeleton/internal/module/task/domain/model"
)

// CreatingOperation — интерфейс операции создания задачи.
type CreatingOperation interface {
	Create(ctx context.Context, title model.Title, description string) (*model.Task, error)
}
