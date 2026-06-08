package cmd

import (
	"context"
	"errors"
	"fmt"
	
	"olivine/internal/repo"
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
	args := cmd.Args()
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: argument count mismatch: expect '%d' got '%d'", ErrValidation, len(args), 1)
	}

	k := args[0]
	v, err := c.storage.Get(ctx, k.String())
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return resp.NewNullBulkString(), nil
		}
		return nil, fmt.Errorf("%w: %w", ErrStorage, err)
	}

	return resp.NewBulkString(v), nil
}
