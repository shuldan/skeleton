package model

import (
	"github.com/shuldan/events"
	"github.com/shuldan/skeleton/internal/event"
)

// Task — агрегат задачи.
type Task struct {
	id          TaskID
	title       Title
	description string
	status      status
	version     int
	events      []event.Event
}

// NewTask создаёт задачу в статусе draft.
func NewTask(id TaskID, title Title, description string) *Task {
	task := &Task{
		id:          id,
		title:       title,
		description: description,
		status:      statusDraft,
		version:     1,
		events:      make([]event.Event, 0),
	}

	task.record(&event.TaskCreated{
		BaseEvent: events.NewBaseEvent("TaskCreated", task.id.String()),
		TaskID:    task.id.String(),
		Title:     title.String(),
	})

	return task
}

func (t *Task) ID() TaskID { return t.id }

// Complete переводит задачу в статус done.
func (t *Task) Complete() error {
	newStatus, err := t.status.transitionTo(statusDone)
	if err != nil {
		return err
	}

	t.status = newStatus

	t.record(&event.TaskCompleted{
		BaseEvent: events.NewBaseEvent("TaskCompleted", t.id.String()),
		TaskID:    t.id.String(),
	})

	return nil
}

// RepresentTo передаёт данные агрегата через presenter.
func (t *Task) RepresentTo(p TaskPresenter) {
	p.SetID(t.id.String()).
		SetTitle(t.title.String()).
		SetDescription(t.description).
		SetStatus(t.status.String()).
		SetVersion(t.version)
}

// ReleaseEvents возвращает накопленные события и очищает очередь.
func (t *Task) ReleaseEvents() []event.Event {
	evts := t.events
	t.events = nil
	return evts
}

func (t *Task) record(e event.Event) {
	t.events = append(t.events, e)
}
