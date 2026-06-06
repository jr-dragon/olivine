package main

import "log/slog"

func main() {
	app := NewApp()
	if err := app.Run(); err != nil {
		slog.Error("failed to fun app", slog.Any("error", err))
	}
}
