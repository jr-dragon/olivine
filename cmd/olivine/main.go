package main

import "log/slog"

func main() {
	app, err := NewApp()
	if err != nil {
		slog.Error("failed to init app", slog.Any("error", err))
		return
	}

	if err := app.Run(); err != nil {
		slog.Error("failed to run app", slog.Any("error", err))
		return
	}
}
