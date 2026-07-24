package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"olivine/internal/data"
	"olivine/internal/server"
	"olivine/internal/service"
)

const (
	AOFPath = "database.aof"
)

type App struct {
	cfg *data.Config

	worker  service.Worker
	server  server.Server
	httpsrv *http.Server
}

func (app *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	otelShutdown, err := service.SetupOTelSDK(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := otelShutdown(context.Background()); err != nil {
			slog.Error("failed to shutdown otel", slog.Any("error", err))
		}
	}()

	slog.Info("restoring data from disk")
	if err := app.server.RestoreFromDisk(); err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		slog.Info("starting olivine workers")
		if err := app.worker.Start(ctx); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		slog.Info("starting olivine server on :16379")
		if err := app.server.ListenAndServe(); err != nil && !errors.Is(err, server.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		slog.Info("starting web server on 127.0.0.1:6060")
		if err := app.httpsrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

		if err := app.server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		if err := app.httpsrv.Shutdown(shutdownCtx); err != nil {
			return err
		}

		return <-errch
	case err := <-errch:
		return err
	}
}
