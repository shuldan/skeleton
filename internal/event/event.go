package event

import "time"

// Event — интерфейс доменного события.
type Event interface {
	EventName() string
	OccurredAt() time.Time
	AggregateID() string
}
