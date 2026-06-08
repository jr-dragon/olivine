package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"olivine/internal/service/cmd"
	"olivine/pkg/resp"
)

var (
	ErrClient = errors.New("client error")
	ErrServer = errors.New("server error")
)

type Handler interface {
	ServeRESP(context.Context, *resp.Reader) (resp.Value, error)
}

type HandlerFunc func(context.Context, *resp.Command) (resp.Value, error)

func NewHandler(cmds []cmd.Command) Handler {
	factory := make(map[string]HandlerFunc)
	for _, cmd := range cmds {
		factory[cmd.Command()] = cmd.Exec
	}

	return &simpleHandler{
		factory: factory,
	}
}

type simpleHandler struct {
	factory map[string]HandlerFunc
}

func (h *simpleHandler) ServeRESP(ctx context.Context, rd *resp.Reader) (resp.Value, error) {
	cmd, err := resp.ReadCommand(rd)
	if err != nil {
		if errors.Is(err, resp.ErrProtocol) {
			return resp.NewSimpleError(err), fmt.Errorf("%w: %w", ErrClient, err)
		} else {
			return nil, fmt.Errorf("%w: %w", ErrServer, err)
		}
	}

	slog.Info("Read RESP command:", slog.Any("command", cmd))

	return h.exec(ctx, cmd)
}

func (h *simpleHandler) exec(ctx context.Context, cmd *resp.Command) (resp.Value, error) {
	f, ok := h.factory[cmd.Command()]
	if !ok {
		err := fmt.Errorf("%w: unknown command: %s", ErrClient, cmd.Command())
		return resp.NewSimpleError(err), err
	}

	return f(ctx, cmd)
}
