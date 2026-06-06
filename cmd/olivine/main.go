package main

import "log/slog"

func main() {
	app := &App{}
	if err := app.Run(); err != nil {
		slog.Error("failed to fun app", slog.Any("error", err))
	}
}
