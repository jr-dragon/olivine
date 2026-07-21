package main

import (
	"log/slog"
	"os"
	"runtime"
	"strconv"

	"olivine/internal/data"
)

func main() {
	if os.Getenv("PPROF_MUTEX_FRAC") != "" {
		if rate, err := strconv.Atoi(os.Getenv("PPROF_MUTEX_FRAC")); err != nil {
			panic("PPROF_MUTEX_FRAC must be integer: " + err.Error())
		} else {
			runtime.SetMutexProfileFraction(rate)
		}
	}

	if os.Getenv("PPROF_BLOCK_RATE") != "" {
		if rate, err := strconv.Atoi(os.Getenv("PPROF_BLOCK_RATE")); err != nil {
			panic("PPROF_BLOCK_RATE must be integer: " + err.Error())
		} else {
			runtime.SetBlockProfileRate(rate)
		}
	}

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
