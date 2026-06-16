package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

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
	ctx, stop := signal.NotifyContext(context.TODO(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("restoring data from disk")
	if err := app.srv.RestoreFromDisk(); err != nil {
		return err
	}

	errch := make(chan error, 1)
	go func() {
		slog.Info("starting olivine server")
		errch <- app.srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.Info("closing olivine server")
		if err := app.srv.Close(); err != nil {
			return err
		}

		return <-errch
	case err := <-errch:
		return err
	}
}
