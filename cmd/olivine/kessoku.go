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
	kessoku.Provide(func(cfg *data.Config) (*App, error) {
		var handler server.Handler
		var restorer server.Restorer
		var aof service.AOF
		if cfg.AOFEnabled {
			var err error
			aof, err = service.NewAOF(cfg, AOFPath)
			if err != nil {
				return nil, err
			}

			handler = server.NewHandler(cmd.NewCommands(repo.NewStorage()), server.NewAOFMiddleware(aof))
			restorer = server.NewRestorer(aof, handler)
		} else {
			handler = server.NewHandler(cmd.NewCommands(repo.NewStorage()))
		}

		return &App{
			cfg: cfg,

			aof: aof,
			srv: server.NewServer(handler, restorer),
		}, nil
	}),
)
