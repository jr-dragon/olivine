package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"olivine/internal/repo"
	"olivine/pkg/resp"
)

type TTL struct {
	storage repo.Storage
}

func NewTTL(storage repo.Storage) *TTL {
	return &TTL{
		storage: storage,
	}
}

func (c *TTL) Command() string {
	return "TTL"
}

func (c *TTL) Exec(ctx context.Context, cmd *resp.Command) (resp.Value, error) {
	if err := c.parse(cmd); err != nil {
		return nil, err
	}

	args := cmd.Args()

	k := args[0]
	v, err := c.storage.Get(ctx, k.String())
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return resp.Integer(-2), nil
		}
		return nil, fmt.Errorf("%w: %w", ErrStorage, err)
	}
	if v.ExpiresAt() == nil {
		return resp.Integer(-1), nil
	}

	return resp.Integer(time.Until(*v.ExpiresAt()) / time.Second), nil
}

func (c *TTL) parse(cmd *resp.Command) error {
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
