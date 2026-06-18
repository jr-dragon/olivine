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
	kessoku.Bind[service.AOF](kessoku.Provide(func(cfg *data.Config) (service.AOF, error) {
		if !cfg.AOFEnabled {
			return nil, nil
		}

		return service.NewAOF(cfg, AOFPath)
	})),
	kessoku.Bind[server.Handler](kessoku.Provide(func(cfg *data.Config, aof service.AOF) server.Handler {
		middlewares := []server.Middleware{}
		if cfg.AOFEnabled {
			middlewares = append(middlewares, server.NewAOFMiddleware(aof))
		}

		return server.NewHandler(cmd.NewCommands(repo.NewStorage()), middlewares...)
	})),
	kessoku.Bind[server.Restorer](kessoku.Provide(func(cfg *data.Config, aof service.AOF, handler server.Handler) server.Restorer {
		if !cfg.AOFEnabled {
			return nil
		}

		return server.NewRestorer(aof, handler)
	})),
	kessoku.Bind[server.Server](kessoku.Provide(server.NewServer)),
	kessoku.Provide(func(cfg *data.Config, aof service.AOF, srv server.Server) *App {
		return &App{
			cfg: cfg,

			aof: aof,
			srv: srv,
		}
	}),
)
