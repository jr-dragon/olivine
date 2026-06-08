package cmd

import (
	"context"
	"olivine/internal/repo"
	"olivine/pkg/resp"
)

type Command interface {
	Command() string
	Exec(context.Context, *resp.Command) (resp.Value, error)
}

func NewCommands(storage repo.Storage) []Command {
	return []Command{
		&Ping{},

		NewSet(storage),
		NewGet(storage),
	}
}
