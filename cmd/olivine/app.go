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
	"olivine/internal/server"
	"olivine/internal/service"

	"golang.org/x/sync/errgroup"
)

const (
	AOFPath = "database.aof"
)

type App struct {
	cfg *data.Config

	aof service.AOF
	srv server.Server
}

func (app *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("restoring data from disk")
	if err := app.srv.RestoreFromDisk(); err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	if app.cfg.AOFEnabled && app.cfg.AOFFsync == data.AOFFsyncEverySec {
		g.Go(func() error {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					if err := app.aof.Sync(); err != nil {
						return err
					}
				}
			}
		})
	}

	g.Go(func() error {
		slog.Info("starting olivine server")
		if err := app.srv.ListenAndServe(); err != nil && !errors.Is(err, server.ErrServerClosed) {
			return err
		}
		return nil
	})

	errch := make(chan error, 1)
	go func() { errch <- g.Wait() }()

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
