package api

import (
	"net/http"

	"github.com/shuldan/framework/httpserver"

	"github.com/shuldan/skeleton/internal/module/task/application/interactor"
	"github.com/shuldan/skeleton/internal/module/task/domain/model"
)

type getTaskInput struct {
	taskID string
}

func (i *getTaskInput) GetTaskID() string { return i.taskID }

// GetTaskOutput реализует model.TaskPresenter.
type GetTaskOutput struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Version     int    `json:"version"`
}

func (o *GetTaskOutput) SetID(v string) model.TaskPresenter          { o.ID = v; return o }
func (o *GetTaskOutput) SetTitle(v string) model.TaskPresenter       { o.Title = v; return o }
func (o *GetTaskOutput) SetDescription(v string) model.TaskPresenter { o.Description = v; return o }
func (o *GetTaskOutput) SetStatus(v string) model.TaskPresenter      { o.Status = v; return o }
func (o *GetTaskOutput) SetVersion(v int) model.TaskPresenter        { o.Version = v; return o }

type getTaskHandler struct {
	interactor *interactor.GetTaskInteractor
}

// NewGetTaskHandler возвращает обработчик получения задачи.
func NewGetTaskHandler(
	inter *interactor.GetTaskInteractor,
) http.HandlerFunc {
	h := &getTaskHandler{interactor: inter}

	return httpserver.Wrap(h.handle)
}

func (h *getTaskHandler) handle(
	w http.ResponseWriter, r *http.Request,
) error {
	id := httpserver.PathParam(r, "id")
	output := &GetTaskOutput{}

	if err := h.interactor.Handle(
		r.Context(), &getTaskInput{taskID: id}, output,
	); err != nil {
		return err
	}

	httpserver.OK(w, output)

	return nil
}
