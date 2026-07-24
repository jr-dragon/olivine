package cmd

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"olivine/internal/repo"
	"olivine/pkg/resp"
)

type Command interface {
	Command() string
	Exec(context.Context, *resp.Command) (resp.Value, error)
}

func NewCommands(storage repo.Storage, tracer trace.Tracer) []Command {
	return []Command{
		NewPing(tracer),

		NewSet(storage, tracer),
		NewGet(storage, tracer),

		NewTTL(storage, tracer),
	}
}
