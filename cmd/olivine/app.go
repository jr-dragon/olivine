package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"olivine/internal/data"
	"olivine/internal/repo"
	"olivine/internal/server"
	"olivine/internal/service"
	"olivine/internal/service/cmd"
)

const (
	AOFPath = "database.aof"
)

type App struct {
	cfg *data.Config
	srv server.Server
}

func NewApp(cfg *data.Config) (*App, error) {
	var handler server.Handler
	var restorer server.Restorer
	if cfg.AOFEnabled {
		aof, err := service.NewAOF(AOFPath)
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
		srv: server.NewServer(handler, restorer),
	}, nil
}

func (app *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("restoring data from disk")
	if err := app.srv.RestoreFromDisk(); err != nil {
		return err
	}

	errch := make(chan error, 1)
	go func() {
		slog.Info("starting olivine server")
		if err := app.srv.ListenAndServe(); err != nil && !errors.Is(err, server.ErrServerClosed) {
			errch <- err
			return
		}
		errch <- nil
	}()

	select {
	case <-ctx.Done():
		slog.Info("closing olivine server")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := app.srv.Shutdown(shutdownCtx); err != nil {
			return err
		}

		return <-errch
	case err := <-errch:
		return err
	}
}
