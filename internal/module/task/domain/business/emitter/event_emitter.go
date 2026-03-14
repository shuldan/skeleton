package emitter

import (
	"context"

	"github.com/shuldan/skeleton/internal/event"
)

// EventEmitter — интерфейс публикации доменных событий.
type EventEmitter interface {
	Emit(ctx context.Context, events []event.Event)
}
