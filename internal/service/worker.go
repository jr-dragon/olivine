package service

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"olivine/internal/data"
)

type Worker interface {
	Start(context.Context) error
}

func NewWorker(cfg *data.Config, aof AOF) Worker {
	return &worker{
		cfg: cfg,
		aof: aof,
	}
}

type worker struct {
	cfg *data.Config

	aof AOF
}

func (w *worker) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	if w.cfg.AOFEnabled && w.cfg.AOFFsync == data.AOFFsyncAlways {
		g.Go(w.aofSyncer(ctx))
	}

	return g.Wait()
}

func (w *worker) aofSyncer(ctx context.Context) func() error {
	return func() error {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := w.aof.Sync(); err != nil {
					return err
				}
			}
		}
	}
}
