package cmd

import (
	"context"
	"fmt"

	"olivine/pkg/resp"
)

type Ping struct{}

func (c *Ping) Command() string {
	return "PING"
}

func (c *Ping) Exec(_ context.Context, cmd *resp.Command) (resp.Value, error) {
	if err := c.parse(cmd); err != nil {
		return nil, err
	}

	args := cmd.Args()
	if len(args) == 0 {
		return resp.SimpleString("PONG"), nil
	}

	return args[0], nil
}

func (c *Ping) parse(cmd *resp.Command) error {
	const (
		awaitingMessage = iota
		messageReceived
		tooManyMessages
	)

	state := awaitingMessage
	for range cmd.Args() {
		switch state {
		case awaitingMessage:
			state = messageReceived
		case messageReceived:
			state = tooManyMessages
		}
	}

	if state == tooManyMessages {
		return fmt.Errorf("%w: argument count mismatch: expect at most %d got %d", ErrSyntax, 1, len(cmd.Args()))
	}

	return nil
}
