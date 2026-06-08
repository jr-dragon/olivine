package main

import (
	"log/slog"

	"olivine/internal/repo"
	"olivine/internal/server"
	"olivine/internal/service"
	"olivine/internal/service/cmd"
)

const (
	AOFPath = "database.aof"
)

type App struct {
	srv server.Server
}

func NewApp() (*App, error) {
	aof, err := service.NewAOF(AOFPath)
	if err != nil {
		return nil, err
	}

	handler := server.NewHandler(cmd.NewCommands(repo.NewStorage()), server.NewAOFMiddleware(aof))

	return &App{
		srv: server.NewServer(
			handler,
			server.NewRestorer(aof, handler),
		),
	}, nil
}

func (app *App) Run() error {
	slog.Info("restore data from disk")
	if err := app.srv.RestoreFromDisk(); err != nil {
		return err
	}

	slog.Info("starting olivine server...")
	if err := app.srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
