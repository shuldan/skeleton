package event

// TaskCreated публикуется при создании задачи.
type TaskCreated struct {
	TaskID string `json:"task_id"`
	Title  string `json:"title"`
}

// EventKey реализует events.KeyedEvent для упорядоченной обработки.
func (e *TaskCreated) EventKey() string { return e.TaskID }

// TaskCompleted публикуется при завершении задачи.
type TaskCompleted struct {
	TaskID string `json:"task_id"`
}

// EventKey реализует events.KeyedEvent для упорядоченной обработки.
func (e *TaskCompleted) EventKey() string { return e.TaskID }
