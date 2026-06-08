package main

import (
	"log/slog"

	"olivine/internal/repo"
	"olivine/internal/server"
	"olivine/internal/service"
	"olivine/internal/service/cmd"
)

type App struct {
	srv server.Server
}

func NewApp() (*App, error) {
	aof, err := service.NewAOF("database.aof")
	if err != nil {
		return nil, err
	}

	return &App{
		srv: server.NewServer(server.NewHandler(cmd.NewCommands(repo.NewStorage()), server.NewAOFMiddleware(aof))),
	}, nil
}

func (app *App) Run() error {
	slog.Info("starting olivine server...")
	if err := app.srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
