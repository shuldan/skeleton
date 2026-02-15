package model

// TaskPresenter — интерфейс представления задачи.
// Реализуется output-структурами в presentation layer.
type TaskPresenter interface {
	SetID(id string) TaskPresenter
	SetTitle(title string) TaskPresenter
	SetDescription(description string) TaskPresenter
	SetStatus(status string) TaskPresenter
	SetVersion(version int) TaskPresenter
}
