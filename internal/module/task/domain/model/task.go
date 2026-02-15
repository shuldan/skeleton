package model

// Task — агрегат задачи.
type Task struct {
	id          TaskID
	title       Title
	description string
	status      Status
	version     int
}

// NewTask создаёт задачу в статусе draft.
func NewTask(title Title, description string) *Task {
	return &Task{
		id:          NewTaskID(),
		title:       title,
		description: description,
		status:      StatusDraft,
		version:     1,
	}
}

func (t *Task) ID() TaskID          { return t.id }
func (t *Task) Title() Title        { return t.title }
func (t *Task) Description() string { return t.description }
func (t *Task) Status() Status      { return t.status }
func (t *Task) Version() int        { return t.version }

// Complete переводит задачу в статус done.
func (t *Task) Complete() error {
	newStatus, err := t.status.TransitionTo(StatusDone)
	if err != nil {
		return err
	}

	t.status = newStatus

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
