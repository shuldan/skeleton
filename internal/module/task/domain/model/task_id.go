package model

import "github.com/google/uuid"

// TaskID — идентификатор задачи.
type TaskID uuid.UUID

// NewTaskID генерирует новый идентификатор.
func NewTaskID() TaskID {
	return TaskID(uuid.New())
}

func (id TaskID) String() string {
	return uuid.UUID(id).String()
}
