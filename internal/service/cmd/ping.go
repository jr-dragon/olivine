package cmd

import (
	"context"
	
	"olivine/pkg/resp"
)

type Ping struct{}

func (c *Ping) Command() string {
	return "PING"
}

func (c *Ping) Exec(_ context.Context, cmd *resp.Command) (resp.Value, error) {
	args := cmd.Args()
	if len(args) == 0 {
		return resp.SimpleString("PONG"), nil
	}

	return args[0], nil
}
