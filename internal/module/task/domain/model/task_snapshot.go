package model

// TaskSnapshot — плоское представление Task для персистентности.
type TaskSnapshot struct {
	ID          string
	Title       string
	Description string
	Status      string
	Version     int
}

// Snapshot возвращает снимок агрегата.
func (t *Task) Snapshot() TaskSnapshot {
	return TaskSnapshot{
		ID:          t.id.String(),
		Title:       t.title.String(),
		Description: t.description,
		Status:      t.status.String(),
		Version:     t.version,
	}
}

// Restore восстанавливает агрегат из снимка без повторной валидации.
func (s *TaskSnapshot) Restore() (*Task, error) {
	t, err := NewTitle(s.Title)
	if err != nil {
		return nil, err
	}

	return &Task{
		id:          NewTaskID(s.ID),
		title:       t,
		description: s.Description,
		status:      statusFromString(s.Status),
		version:     s.Version,
		events:      make([]any, 0),
	}, nil
}
