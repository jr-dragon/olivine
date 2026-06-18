package main

import (
	"olivine/internal/data"
	"olivine/internal/repo"
	"olivine/internal/server"
	"olivine/internal/service"
	"olivine/internal/service/cmd"

	"github.com/mazrean/kessoku"
)

//go:generate go tool kessoku $GOFILE

var _ = kessoku.Inject[*App](
	"NewApp",
	// repositories
	kessoku.Bind[repo.Storage](kessoku.Provide(repo.NewStorage)),

	// services
	kessoku.Bind[service.AOF](kessoku.Provide(func(cfg *data.Config) (service.AOF, error) {
		if !cfg.AOFEnabled {
			return nil, nil
		}

		return service.NewAOF(cfg, AOFPath)
	})),
	kessoku.Provide(cmd.NewCommands),

	// servers
	kessoku.Bind[server.Handler](kessoku.Provide(func(cfg *data.Config, aof service.AOF, cmds []cmd.Command) server.Handler {
		middlewares := []server.Middleware{}
		if cfg.AOFEnabled {
			middlewares = append(middlewares, server.NewAOFMiddleware(aof))
		}

		return server.NewHandler(cmds, middlewares...)
	})),
	kessoku.Bind[server.Restorer](kessoku.Provide(func(cfg *data.Config, aof service.AOF, handler server.Handler) server.Restorer {
		if !cfg.AOFEnabled {
			return nil
		}

		return server.NewRestorer(aof, handler)
	})),
	kessoku.Bind[server.Server](kessoku.Provide(server.NewServer)),

	// application
	kessoku.Provide(func(cfg *data.Config, aof service.AOF, srv server.Server) *App {
		return &App{
			cfg: cfg,

			aof: aof,
			srv: srv,
		}
	}),
)
