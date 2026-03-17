package emitter

import "context"

// EventEmitter — интерфейс публикации доменных событий.
type EventEmitter interface {
	Emit(ctx context.Context, events []any)
}
