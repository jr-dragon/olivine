package cmd

import (
	"context"
	"fmt"

	"olivine/internal/repo"
	"olivine/internal/repo/object"
	"olivine/pkg/resp"
)

type Set struct {
	storage repo.Storage
}

func NewSet(storage repo.Storage) *Set {
	return &Set{storage: storage}
}

func (c *Set) Command() string {
	return "SET"
}

func (c *Set) Exec(ctx context.Context, cmd *resp.Command) (resp.Value, error) {
	args := cmd.Args()
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: argument count mismatch: expect '%d' got '%d'", ErrValidation, len(args), 2)
	}

	k, v := args[0], args[1]
	if err := c.storage.Set(ctx, object.NewString(k.String(), v.String(), nil)); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrStorage, err)
	}

	return resp.SimpleString("OK"), nil
}
