package payment

import (
	"database/sql"

	"github.com/shuldan/framework/commandbus"
	"github.com/shuldan/framework/httpserver"
	"github.com/shuldan/framework/migration"

	"github.com/shuldan/skeleton/internal/logger"
	"github.com/shuldan/skeleton/internal/module/payment/application/interactor"
	invoicemigration "github.com/shuldan/skeleton/internal/module/payment/infrastructure/migration"
	"github.com/shuldan/skeleton/internal/module/payment/infrastructure/persistence"
	"github.com/shuldan/skeleton/internal/module/payment/presentation/api"
	"github.com/shuldan/skeleton/internal/module/payment/presentation/commandhandler"
)

// Module — фасад DDD-модуля Payment.
type Module struct {
	createInteractor *interactor.CreateInvoiceInteractor
	listInteractor   *interactor.ListInvoicesInteractor
	log              logger.Logger
}

// NewModule собирает внутренний граф зависимостей.
func NewModule(db *sql.DB, log logger.Logger) *Module {
	repo := persistence.NewInvoiceRepository(db)

	return &Module{
		createInteractor: interactor.NewCreateInvoiceInteractor(repo),
		listInteractor:   interactor.NewListInvoicesInteractor(repo),
		log:              log,
	}
}

// Routes регистрирует HTTP-маршруты.
func (m *Module) Routes(router *httpserver.Router) {
	group := router.Group("/api/v1/invoices")

	group.GET("", api.NewListInvoicesHandler(m.listInteractor))
}

// CommandHandlers регистрирует обработчики команд.
func (m *Module) CommandHandlers(receiver *commandbus.CommandReceiver) {
	if err := receiver.Handle(
		"CreateInvoice",
		commandhandler.DeserializeCreateInvoice,
		commandhandler.HandleCreateInvoice(m.createInteractor),
	); err != nil {
		m.log.Error("failed to register command handler",
			"command", "CreateInvoice",
			"error", err,
		)
	}
}

// Migrations регистрирует миграции модуля.
func (m *Module) Migrations(runner *migration.Runner) {
	runner.Register("default", invoicemigration.CreateInvoicesTable())
}
