package cmd

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/trace"

	"olivine/internal/repo"
	"olivine/internal/repo/object"
	"olivine/internal/util"
	"olivine/pkg/resp"
)

type Get struct {
	storage repo.Storage
	tracer  trace.Tracer
}

func NewGet(storage repo.Storage, tracers ...trace.Tracer) *Get {
	return &Get{
		storage: storage,
		tracer:  util.SelectTracer(tracers),
	}
}

func (c *Get) Command() string {
	return "GET"
}

func (c *Get) Exec(ctx context.Context, cmd *resp.Command) (ret resp.Value, err error) {
	ctx, span := util.StartCommandSpan(ctx, c.tracer, "cmd.(*Get).Exec", cmd)
	defer func() { util.EndSpan(span, err) }()

	if err := c.parse(cmd); err != nil {
		return nil, err
	}

	args := cmd.Args()

	k := args[0]
	v, err := c.storage.Get(ctx, k.String())
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return resp.NewNullBulkString(), nil
		}
		return nil, fmt.Errorf("%w: %w", ErrStorage, err)
	}

	var value resp.Value
	if str, ok := v.(*object.String); ok {
		value = resp.NewBulkString(str.String())
	} else {
		value = resp.NewSimpleError(ErrWrongType)
	}

	return value, nil
}

func (c *Get) parse(cmd *resp.Command) error {
	const (
		awaitingKey = iota
		keyReceived
		tooManyKeys
	)

	state := awaitingKey
	for range cmd.Args() {
		switch state {
		case awaitingKey:
			state = keyReceived
		case keyReceived:
			state = tooManyKeys
		}
	}

	if state != keyReceived {
		return fmt.Errorf("%w: argument count mismatch: expect %d got %d", ErrSyntax, 1, len(cmd.Args()))
	}

	return nil
}
