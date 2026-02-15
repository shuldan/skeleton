package api

import (
	"net/http"

	"github.com/shuldan/framework/httpserver"

	"github.com/shuldan/skeleton/internal/module/task/application/interactor"
	"github.com/shuldan/skeleton/internal/module/task/domain/model"
)

// --- Output ---

type taskItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Version     int    `json:"version"`
}

func (o *taskItem) SetID(v string) model.TaskPresenter          { o.ID = v; return o }
func (o *taskItem) SetTitle(v string) model.TaskPresenter       { o.Title = v; return o }
func (o *taskItem) SetDescription(v string) model.TaskPresenter { o.Description = v; return o }
func (o *taskItem) SetStatus(v string) model.TaskPresenter      { o.Status = v; return o }
func (o *taskItem) SetVersion(v int) model.TaskPresenter        { o.Version = v; return o }

// ListTasksOutput собирает список задач.
type ListTasksOutput struct {
	Tasks []*taskItem `json:"tasks"`
}

// AddTask добавляет новый элемент и возвращает его как presenter.
func (o *ListTasksOutput) AddTask() model.TaskPresenter {
	item := &taskItem{}
	o.Tasks = append(o.Tasks, item)

	return item
}

// --- Handler ---

type listTasksHandler struct {
	interactor *interactor.ListTasksInteractor
}

// NewListTasksHandler возвращает обработчик списка задач.
func NewListTasksHandler(
	inter *interactor.ListTasksInteractor,
) http.HandlerFunc {
	h := &listTasksHandler{interactor: inter}

	return httpserver.Wrap(h.handle)
}

func (h *listTasksHandler) handle(
	w http.ResponseWriter, r *http.Request,
) error {
	output := &ListTasksOutput{
		Tasks: make([]*taskItem, 0),
	}

	if err := h.interactor.Handle(r.Context(), output); err != nil {
		return err
	}

	httpserver.OK(w, output)

	return nil
}
