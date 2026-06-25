package server

import (
	"context"
	"fmt"

	"olivine/internal/service"
	"olivine/pkg/resp"
)

func NewAOFMiddleware(aof service.AOF) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, cmd *resp.Command) (resp.Value, error) {
			ret, err := next(ctx, cmd)
			if err != nil {
				return ret, err
			}
			if err := aof.Write(cmd); err != nil {
				return nil, fmt.Errorf("%w: failed to write AOF: %w", ErrServer, err)
			}

			return ret, err
		}
	}
}
