package bootstrap

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shuldan/config"
	"github.com/shuldan/queue"
	memorymq "github.com/shuldan/queue/broker/memory"

	"github.com/shuldan/framework"
	"github.com/shuldan/framework/command"
	"github.com/shuldan/framework/commandbus"
	"github.com/shuldan/framework/database"
	"github.com/shuldan/framework/eventbus"
	"github.com/shuldan/framework/httpserver"
	"github.com/shuldan/framework/logger"
	"github.com/shuldan/framework/migration"
	"github.com/shuldan/framework/queueworker"

	_ "github.com/lib/pq"

	"github.com/shuldan/skeleton/internal/module/payment"
	"github.com/shuldan/skeleton/internal/module/task"
)

const (
	serviceName     = "task-service"
	gracefulTimeout = 15 * time.Second
)

// Run — точка входа приложения.
func Run(ctx context.Context) error {
	k, cfg, log, err := initKernel()
	if err != nil {
		return err
	}

	dbm, err := initDatabase(cfg, log)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	bus := initEventBus(cfg)
	broker := memorymq.New()

	taskMod := task.NewModule(
		dbm.Default(), bus.Dispatcher(), log,
	)
	paymentMod := payment.NewModule(dbm.Default(), log)

	registerCommands(
		k, cfg, log, dbm, bus, broker, taskMod, paymentMod,
	)
	registerShutdown(k, broker, dbm)

	return k.Run(ctx, os.Args[1:])
}

func initKernel() (
	*framework.Kernel, *config.Config, *logger.Logger, error,
) {
	k, err := framework.NewKernel(
		framework.WithConfigFile("config/config.yaml"),
		framework.WithEnvPrefix("APP_"),
		framework.WithProfileEnv("APP_ENV"),
	)
	if err != nil {
		return nil, nil, nil, err
	}

	return k, k.Config(), k.Logger(), nil
}

func initDatabase(
	cfg *config.Config, log *logger.Logger,
) (*database.Manager, error) {
	return database.NewManager(
		map[string]database.ConnectionConfig{
			"default": {
				Driver: cfg.GetString(
					"database.connections.default.driver", "postgres",
				),
				DSN: cfg.GetString(
					"database.connections.default.dsn",
				),
				MaxOpenConns: cfg.GetInt(
					"database.connections.default.max_open_conns", 25,
				),
				MaxIdleConns: cfg.GetInt(
					"database.connections.default.max_idle_conns", 5,
				),
				ConnMaxLifetime: cfg.GetDuration(
					"database.connections.default.conn_max_lifetime",
					5*time.Minute,
				),
			},
		},
		log,
	)
}

func initEventBus(cfg *config.Config) *eventbus.Module {
	return eventbus.NewModule(eventbus.Config{
		Async:      cfg.GetBool("events.async", true),
		Workers:    cfg.GetInt("events.workers", 4),
		BufferSize: cfg.GetInt("events.buffer_size", 128),
	})
}

func registerCommands(
	k *framework.Kernel,
	cfg *config.Config,
	log *logger.Logger,
	dbm *database.Manager,
	bus *eventbus.Module,
	broker queue.Broker,
	taskMod *task.Module,
	paymentMod *payment.Module,
) {
	router := buildRouter(log, taskMod, paymentMod)

	server := httpserver.NewModule(router, httpserver.Config{
		Host: cfg.GetString("server.host", "0.0.0.0"),
		Port: cfg.GetInt("server.port", 8080),
	})

	// Events: listeners
	taskMod.Listeners(bus.Dispatcher())

	// Events: outbound relay
	relay := eventbus.NewOutboundRelay(bus.Dispatcher(), broker, log)
	taskMod.Relays(relay)

	// Command bus: sender (task → payment)
	sender := commandbus.NewCommandSender(broker, log,
		commandbus.WithSender(serviceName),
		commandbus.WithReplyTo(serviceName),
	)
	sender.Forward("CreateInvoice")
	taskMod.CommandSenders(bus.Dispatcher(), sender)

	// Command bus: receiver (payment handles commands)
	receiver := commandbus.NewCommandReceiver(broker, log,
		commandbus.WithIdempotencyTTL(24*time.Hour),
	)
	paymentMod.CommandHandlers(receiver)

	// Command bus: reply listener (task receives results)
	replyListener := commandbus.NewReplyListener(broker, log,
		commandbus.WithListenerServiceName(serviceName),
	)
	taskMod.ReplyHandlers(replyListener)

	// Queue workers
	qw := queueworker.NewModule(log)
	taskMod.Consumers(qw, broker)

	// Command receiver consumers
	for _, reg := range receiver.Registrations() {
		qw.Register(reg)
	}

	// Reply listener consumer
	qw.Register(queueworker.Registration{
		Name: "reply-listener",
		Run:  replyListener.Run,
	})

	// Migrations
	runner := migration.NewRunner(
		dbm, log, migration.WithAdvisoryLock(),
	)
	taskMod.Migrations(runner)
	paymentMod.Migrations(runner)

	appName := cfg.GetString("app.name", "app")

	k.Command(
		command.Serve(appName, log, gracefulTimeout,
			dbm, bus, server, qw,
		),
		command.QueueWork(appName, log, gracefulTimeout,
			dbm, bus, qw,
		),
		command.MigrateUp(runner),
		command.MigrateDown(runner),
		command.MigrateStatus(runner),
		command.MigratePlan(runner),
		command.Health(dbm),
		command.ConfigDump(cfg),
	)
}

func registerShutdown(
	k *framework.Kernel,
	broker queue.Broker,
	dbm *database.Manager,
) {
	k.OnShutdown(func() { _ = broker.Close() })
	k.OnShutdown(func() { _ = dbm.Stop(context.Background()) })
}
