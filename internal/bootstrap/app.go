package bootstrap

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shuldan/commands"
	codecjson "github.com/shuldan/commands/codec/json"
	memorytransport "github.com/shuldan/commands/transport/memory"
	"github.com/shuldan/config"
	"github.com/shuldan/events"
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

	// Command bus: transport + codec.
	transport := memorytransport.New()
	codec := codecjson.New()

	cmdClient, err := commands.NewCommandClient(transport, codec,
		commands.WithTimeout(10*time.Second),
	)
	if err != nil {
		return fmt.Errorf("command client: %w", err)
	}

	cmdServer, err := commands.NewCommandServer(transport, codec)
	if err != nil {
		return fmt.Errorf("command server: %w", err)
	}

	cmdModule := commandbus.NewModule(
		commandbus.WithClient(cmdClient),
		commandbus.WithServer(cmdServer),
	)

	taskMod := task.NewModule(
		dbm.Default(), bus.Dispatcher(), log,
	)
	paymentMod := payment.NewModule(dbm.Default(), log)

	// Register command handlers on server.
	paymentMod.CommandHandlers(cmdServer)

	registerCommands(
		k, cfg, log, dbm, bus, cmdModule, broker,
		taskMod, paymentMod, cmdClient,
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
	opts := make([]events.Option, 0)

	if cfg.GetBool("events.async", true) {
		opts = append(opts, events.WithAsyncMode())
	}

	if workers := cfg.GetInt("events.workers", 4); workers > 0 {
		opts = append(opts, events.WithWorkerPool(workers))
	}

	dispatcher := events.New(opts...)

	return eventbus.NewModule(dispatcher)
}

func registerCommands(
	k *framework.Kernel,
	cfg *config.Config,
	log *logger.Logger,
	dbm *database.Manager,
	bus *eventbus.Module,
	cmdModule *commandbus.Module,
	broker queue.Broker,
	taskMod *task.Module,
	paymentMod *payment.Module,
	cmdClient *commands.CommandClient,
) {
	router := buildRouter(log, taskMod, paymentMod)

	server := httpserver.NewModule(router, httpserver.Config{
		Host: cfg.GetString("server.host", "0.0.0.0"),
		Port: cfg.GetInt("server.port", 8080),
	})

	// Events: listeners.
	taskMod.Listeners(bus.Dispatcher())

	// Events: command senders (task → payment via command bus).
	taskMod.CommandSenders(bus.Dispatcher(), cmdClient)

	// Queue workers.
	qw := queueworker.NewModule(log)
	taskMod.Consumers(qw, broker)

	// Migrations.
	runner := migration.NewRunner(
		dbm, log, migration.WithAdvisoryLock(),
	)
	taskMod.Migrations(runner)
	paymentMod.Migrations(runner)

	appName := cfg.GetString("app.name", "app")

	k.Command(
		command.Serve(appName, log, gracefulTimeout,
			dbm, bus, cmdModule, server, qw,
		),
		command.QueueWork(appName, log, gracefulTimeout,
			dbm, bus, cmdModule, qw,
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
