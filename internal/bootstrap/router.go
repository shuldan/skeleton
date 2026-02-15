package bootstrap

import (
	"github.com/shuldan/framework/httpserver"
	"github.com/shuldan/framework/httpserver/middleware"
	"github.com/shuldan/framework/logger"

	"github.com/shuldan/skeleton/internal/module/task"
)

func buildRouter(
	log *logger.Logger, taskMod *task.Module,
) *httpserver.Router {
	router := httpserver.NewRouter()

	router.Use(
		middleware.Recovery(log.Error),
		middleware.RequestID(),
		middleware.Logging(log.Info),
	)

	taskMod.Routes(router)

	return router
}
