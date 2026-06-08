package cmd

import (
	"context"
	"olivine/pkg/resp"
)

type Command interface {
	Command() string
	Exec(context.Context, *resp.Command) (resp.Value, error)
}

func NewCommands() []Command {
	return []Command{
		&Ping{},
	}
}
