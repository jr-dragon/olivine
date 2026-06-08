package main

import (
	"log/slog"

	"olivine/internal/server"
	"olivine/internal/service/cmd"
)

type App struct {
	srv server.Server
}

func NewApp() *App {
	return &App{
		srv: server.NewServer(server.NewHandler(cmd.NewCommands())),
	}
}

func (app *App) Run() error {
	slog.Info("starting olivine server...")
	if err := app.srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
