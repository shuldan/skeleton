package model

import "github.com/google/uuid"

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

// Restore восстанавливает агрегат из снимка.
func (s *TaskSnapshot) Restore() (*Task, error) {
	title, err := NewTitle(s.Title)
	if err != nil {
		return nil, err
	}

	return &Task{
		id:          TaskID(uuid.MustParse(s.ID)),
		title:       title,
		description: s.Description,
		status:      Status(s.Status),
		version:     s.Version,
	}, nil
}
