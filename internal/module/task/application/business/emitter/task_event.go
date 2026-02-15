package emitter

import "github.com/shuldan/skeleton/internal/module/task/domain/model"

// taskEvent — приватный presenter для извлечения данных агрегата.
type taskEvent struct {
	taskID      string
	title       string
	description string
	status      string
	version     int
}

func (e *taskEvent) SetID(id string) model.TaskPresenter {
	e.taskID = id
	return e
}

func (e *taskEvent) SetTitle(title string) model.TaskPresenter {
	e.title = title
	return e
}

func (e *taskEvent) SetDescription(desc string) model.TaskPresenter {
	e.description = desc
	return e
}

func (e *taskEvent) SetStatus(status string) model.TaskPresenter {
	e.status = status
	return e
}

func (e *taskEvent) SetVersion(version int) model.TaskPresenter {
	e.version = version
	return e
}
