package api

import (
	"net/http"

	"github.com/shuldan/framework/httpserver"

	"github.com/shuldan/skeleton/internal/module/payment/application/interactor"
	"github.com/shuldan/skeleton/internal/module/payment/domain/model"
)

// --- Output ---

type invoiceItem struct {
	ID      string `json:"id"`
	TaskID  string `json:"task_id"`
	Amount  int    `json:"amount"`
	Status  string `json:"status"`
	Version int    `json:"version"`
}

func (o *invoiceItem) SetID(v string) model.InvoicePresenter {
	o.ID = v
	return o
}

func (o *invoiceItem) SetTaskID(v string) model.InvoicePresenter {
	o.TaskID = v
	return o
}

func (o *invoiceItem) SetAmount(v int) model.InvoicePresenter {
	o.Amount = v
	return o
}

func (o *invoiceItem) SetStatus(v string) model.InvoicePresenter {
	o.Status = v
	return o
}

func (o *invoiceItem) SetVersion(v int) model.InvoicePresenter {
	o.Version = v
	return o
}

// ListInvoicesOutput собирает список счетов.
type ListInvoicesOutput struct {
	Invoices []*invoiceItem `json:"invoices"`
}

// AddInvoice добавляет элемент и возвращает presenter.
func (o *ListInvoicesOutput) AddInvoice() model.InvoicePresenter {
	item := &invoiceItem{}
	o.Invoices = append(o.Invoices, item)

	return item
}

// --- Handler ---

type listInvoicesHandler struct {
	interactor *interactor.ListInvoicesInteractor
}

// NewListInvoicesHandler возвращает обработчик списка счетов.
func NewListInvoicesHandler(
	inter *interactor.ListInvoicesInteractor,
) http.HandlerFunc {
	h := &listInvoicesHandler{interactor: inter}

	return httpserver.Wrap(h.handle)
}

func (h *listInvoicesHandler) handle(
	w http.ResponseWriter, r *http.Request,
) error {
	output := &ListInvoicesOutput{
		Invoices: make([]*invoiceItem, 0),
	}

	if err := h.interactor.Handle(r.Context(), output); err != nil {
		return err
	}

	httpserver.OK(w, output)

	return nil
}
