package main

import (
	"log/slog"
	"olivine/internal/data"
)

func main() {
	cfg, err := data.NewConfig("redis.conf")
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		return
	}

	app, err := NewApp(cfg)
	if err != nil {
		slog.Error("failed to init app", slog.Any("error", err))
		return
	}

	if err := app.Run(); err != nil {
		slog.Error("failed to run app", slog.Any("error", err))
		return
	}
}
