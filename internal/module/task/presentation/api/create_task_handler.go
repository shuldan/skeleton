package api

import (
	"net/http"

	"github.com/shuldan/framework/httpserver"

	"github.com/shuldan/skeleton/internal/module/task/application/interactor"
	"github.com/shuldan/skeleton/internal/module/task/domain/model"
)

// --- Input / Output ---

type createTaskBody struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type createTaskInput struct {
	body createTaskBody
}

func (i *createTaskInput) GetTitle() string       { return i.body.Title }
func (i *createTaskInput) GetDescription() string { return i.body.Description }

// CreateTaskOutput реализует model.TaskPresenter
// и сериализуется в JSON для клиента.
type CreateTaskOutput struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Version     int    `json:"version"`
}

func (o *CreateTaskOutput) SetID(v string) model.TaskPresenter          { o.ID = v; return o }
func (o *CreateTaskOutput) SetTitle(v string) model.TaskPresenter       { o.Title = v; return o }
func (o *CreateTaskOutput) SetDescription(v string) model.TaskPresenter { o.Description = v; return o }
func (o *CreateTaskOutput) SetStatus(v string) model.TaskPresenter      { o.Status = v; return o }
func (o *CreateTaskOutput) SetVersion(v int) model.TaskPresenter        { o.Version = v; return o }

// --- Handler ---

type createTaskHandler struct {
	interactor *interactor.CreateTaskInteractor
}

// NewCreateTaskHandler возвращает обработчик создания задачи.
func NewCreateTaskHandler(
	inter *interactor.CreateTaskInteractor,
) http.HandlerFunc {
	h := &createTaskHandler{interactor: inter}

	return httpserver.Wrap(h.handle)
}

func (h *createTaskHandler) handle(
	w http.ResponseWriter, r *http.Request,
) error {
	var body createTaskBody
	if err := httpserver.Bind(r, &body); err != nil {
		return err
	}

	output := &CreateTaskOutput{}

	if err := h.interactor.Handle(
		r.Context(), &createTaskInput{body: body}, output,
	); err != nil {
		return err
	}

	httpserver.Created(w, output)

	return nil
}
