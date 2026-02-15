package api

import (
	"net/http"

	"github.com/shuldan/framework/httpserver"

	"github.com/shuldan/skeleton/internal/module/task/application/interactor"
	"github.com/shuldan/skeleton/internal/module/task/domain/model"
)

type completeTaskInput struct {
	taskID string
}

func (i *completeTaskInput) GetTaskID() string { return i.taskID }

// CompleteTaskOutput реализует model.TaskPresenter.
type CompleteTaskOutput struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Version     int    `json:"version"`
}

func (o *CompleteTaskOutput) SetID(v string) model.TaskPresenter    { o.ID = v; return o }
func (o *CompleteTaskOutput) SetTitle(v string) model.TaskPresenter { o.Title = v; return o }
func (o *CompleteTaskOutput) SetDescription(v string) model.TaskPresenter {
	o.Description = v
	return o
}
func (o *CompleteTaskOutput) SetStatus(v string) model.TaskPresenter { o.Status = v; return o }
func (o *CompleteTaskOutput) SetVersion(v int) model.TaskPresenter   { o.Version = v; return o }

type completeTaskHandler struct {
	interactor *interactor.CompleteTaskInteractor
}

// NewCompleteTaskHandler возвращает обработчик завершения задачи.
func NewCompleteTaskHandler(
	inter *interactor.CompleteTaskInteractor,
) http.HandlerFunc {
	h := &completeTaskHandler{interactor: inter}

	return httpserver.Wrap(h.handle)
}

func (h *completeTaskHandler) handle(
	w http.ResponseWriter, r *http.Request,
) error {
	id := httpserver.PathParam(r, "id")
	output := &CompleteTaskOutput{}

	if err := h.interactor.Handle(
		r.Context(), &completeTaskInput{taskID: id}, output,
	); err != nil {
		return err
	}

	httpserver.OK(w, output)

	return nil
}
