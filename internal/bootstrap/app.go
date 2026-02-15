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
	"github.com/shuldan/framework/database"
	"github.com/shuldan/framework/eventbus"
	"github.com/shuldan/framework/httpserver"
	"github.com/shuldan/framework/logger"
	"github.com/shuldan/framework/migration"
	"github.com/shuldan/framework/queueworker"

	_ "github.com/lib/pq"

	"github.com/shuldan/skeleton/internal/module/task"
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

	registerCommands(k, cfg, log, dbm, bus, broker, taskMod)
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
) {
	router := buildRouter(log, taskMod)

	server := httpserver.NewModule(router, httpserver.Config{
		Host: cfg.GetString("server.host", "0.0.0.0"),
		Port: cfg.GetInt("server.port", 8080),
	})

	taskMod.Listeners(bus.Dispatcher())

	relay := eventbus.NewRelay(bus.Dispatcher(), broker, log)
	taskMod.Relays(relay)

	qw := queueworker.NewModule(log)
	taskMod.Consumers(qw, broker)

	runner := migration.NewRunner(
		dbm, log, migration.WithAdvisoryLock(),
	)
	taskMod.Migrations(runner)

	appName := cfg.GetString("app.name", "app")
	timeout := 15 * time.Second

	k.Command(
		command.Serve(appName, log, timeout, dbm, bus, server, qw),
		command.QueueWork(appName, log, timeout, dbm, bus, qw),
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
