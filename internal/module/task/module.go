package task

import (
	"database/sql"
	"encoding/json"

	"github.com/shuldan/events"
	"github.com/shuldan/queue"

	"github.com/shuldan/framework/commandbus"
	"github.com/shuldan/framework/eventbus"
	"github.com/shuldan/framework/httpserver"
	"github.com/shuldan/framework/migration"
	"github.com/shuldan/framework/queueworker"

	"github.com/shuldan/skeleton/internal/logger"
	"github.com/shuldan/skeleton/internal/module/task/application/business/emitter"
	"github.com/shuldan/skeleton/internal/module/task/application/business/operation"
	"github.com/shuldan/skeleton/internal/module/task/application/interactor"
	"github.com/shuldan/skeleton/internal/module/task/application/port"
	"github.com/shuldan/skeleton/internal/module/task/infrastructure/adapter"
	taskmigration "github.com/shuldan/skeleton/internal/module/task/infrastructure/migration"
	"github.com/shuldan/skeleton/internal/module/task/infrastructure/persistence"
	"github.com/shuldan/skeleton/internal/module/task/presentation/api"
	"github.com/shuldan/skeleton/internal/module/task/presentation/job"
	"github.com/shuldan/skeleton/internal/module/task/presentation/listener"
)

// Module — фасад DDD-модуля Task.
type Module struct {
	createInteractor   *interactor.CreateTaskInteractor
	completeInteractor *interactor.CompleteTaskInteractor
	getInteractor      *interactor.GetTaskInteractor
	listInteractor     *interactor.ListTasksInteractor
	notifier           port.NotificationPort
	log                logger.Logger
}

// NewModule собирает внутренний граф зависимостей.
func NewModule(
	db *sql.DB,
	dispatcher *events.Dispatcher,
	log logger.Logger,
) *Module {
	repo := persistence.NewTaskRepository(db)
	taskEmitter := emitter.NewTaskEmitter(dispatcher, log)
	creatingOp := operation.NewCreatingOperation(repo)
	completingOp := operation.NewCompletingOperation(repo)

	createInter := interactor.NewCreateTaskInteractor(
		creatingOp, taskEmitter,
	)
	completeInter := interactor.NewCompleteTaskInteractor(
		completingOp, taskEmitter,
	)
	getInter := interactor.NewGetTaskInteractor(repo)
	listInter := interactor.NewListTasksInteractor(repo)

	notifier := adapter.NewLoggingNotificationAdapter(log)

	return &Module{
		createInteractor:   createInter,
		completeInteractor: completeInter,
		getInteractor:      getInter,
		listInteractor:     listInter,
		notifier:           notifier,
		log:                log,
	}
}

// Routes регистрирует HTTP-маршруты.
func (m *Module) Routes(router *httpserver.Router) {
	group := router.Group("/api/v1/tasks")

	group.POST("", api.NewCreateTaskHandler(m.createInteractor))
	group.GET("", api.NewListTasksHandler(m.listInteractor))
	group.GET("/{id}", api.NewGetTaskHandler(m.getInteractor))
	group.POST("/{id}/complete",
		api.NewCompleteTaskHandler(m.completeInteractor),
	)
}

// Listeners регистрирует подписки на события (in-process).
func (m *Module) Listeners(d *events.Dispatcher) {
	l := listener.NewTaskCompletedListener(m.log)
	events.Subscribe(d, l)
}

// CommandSenders регистрирует отправку команд при событиях.
func (m *Module) CommandSenders(
	d *events.Dispatcher, sender *commandbus.CommandSender,
) {
	l := listener.NewTaskCompletedCommandSender(sender, m.log)
	events.Subscribe(d, l)
}

// ReplyHandlers регистрирует обработчики ответов на команды.
func (m *Module) ReplyHandlers(
	rl *commandbus.ReplyListener,
) {
	rl.OnResult(
		"CreateInvoice",
		listener.DeserializeInvoiceCreated,
		listener.NewInvoiceCreatedReplyHandler(m.log),
	)
}

// Relays настраивает пересылку событий в очередь.
func (m *Module) Relays(relay *eventbus.OutboundRelay) {
	relay.Forward("TaskCompleted", "task.completed",
		eventbus.WithTransform(
			func(e events.Event) ([]byte, error) {
				return json.Marshal(map[string]string{
					"task_id": e.AggregateID(),
					"event":   e.EventName(),
				})
			},
		),
	)
}

// Consumers регистрирует consumer-ов очередей.
func (m *Module) Consumers(
	qw *queueworker.Module, broker queue.Broker,
) {
	j := job.NewSendNotificationJob(broker, m.notifier)

	qw.Register(queueworker.Registration{
		Name: "task-notification",
		Run:  j.Run,
	})
}

// Migrations регистрирует миграции модуля.
func (m *Module) Migrations(runner *migration.Runner) {
	runner.Register("default", taskmigration.CreateTasksTable())
}
