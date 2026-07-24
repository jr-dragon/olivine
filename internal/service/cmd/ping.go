package cmd

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"

	"olivine/internal/util"
	"olivine/pkg/resp"
)

type Ping struct {
	tracer trace.Tracer
}

func NewPing(tracers ...trace.Tracer) *Ping {
	return &Ping{
		tracer: util.SelectTracer(tracers),
	}
}

func (c *Ping) Command() string {
	return "PING"
}

func (c *Ping) Exec(ctx context.Context, cmd *resp.Command) (ret resp.Value, err error) {
	_, span := util.StartCommandSpan(ctx, c.tracer, "cmd.(*Ping).Exec", cmd)
	defer func() { util.EndSpan(span, err) }()

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
