package cmd

import (
	"context"
	"errors"
	"fmt"

	"olivine/internal/repo"
	"olivine/internal/repo/object"
	"olivine/pkg/resp"
)

type Get struct {
	storage repo.Storage
}

func NewGet(storage repo.Storage) *Get {
	return &Get{
		storage: storage,
	}
}

func (c *Get) Command() string {
	return "GET"
}

func (c *Get) Exec(ctx context.Context, cmd *resp.Command) (resp.Value, error) {
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

	var ret resp.Value
	if str, ok := v.(*object.String); ok {
		ret = resp.NewBulkString(str.String())
	} else {
		ret = resp.NewSimpleError(ErrWrongType)
	}

	return ret, nil
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
