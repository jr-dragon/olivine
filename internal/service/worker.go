package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/errgroup"

	"olivine/internal/data"
	"olivine/internal/repo"
)

type Worker interface {
	Start(context.Context) error
}

func NewWorker(cfg *data.Config, aof AOF, storage repo.Storage) Worker {
	return &worker{
		cfg: cfg,

		aof:     aof,
		storage: storage,
	}
}

type worker struct {
	cfg *data.Config

	storage repo.Storage
	aof     AOF
}

func (w *worker) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	if w.cfg.AOFEnabled && w.cfg.AOFFsync == data.AOFFsyncEverySec {
		g.Go(w.aofSyncer(ctx))
	}

	g.Go(w.storagePruner(ctx))

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

func (w *worker) storagePruner(ctx context.Context) func() error {
	return func() error {
		ticker := time.NewTicker(time.Millisecond * 100)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := w.storage.Prune(ctx); err != nil {
					if ctx.Err() != nil && errors.Is(err, ctx.Err()) {
						return nil
					}
					return err
				}
			}
		}
	}
}
