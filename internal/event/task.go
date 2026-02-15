package event

import "github.com/shuldan/events"

// TaskCreated публикуется при создании задачи.
type TaskCreated struct {
	events.BaseEvent
	TaskID string `json:"task_id"`
	Title  string `json:"title"`
}

// TaskCompleted публикуется при завершении задачи.
type TaskCompleted struct {
	events.BaseEvent
	TaskID string `json:"task_id"`
}
