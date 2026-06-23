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
	Exec(context.Context, *resp.Command) (resp.Value, error)
}

type HandlerFunc func(context.Context, *resp.Command) (resp.Value, error)
type Middleware func(HandlerFunc) HandlerFunc

func NewHandler(cmds []cmd.Command, middlewares ...Middleware) Handler {
	h := simpleHandler{}
	h.executor = make(map[string]HandlerFunc)
	h.validator = make(map[string]func(*resp.Command) error)
	for _, cmd := range cmds {
		h.executor[cmd.Command()] = cmd.Exec
		h.validator[cmd.Command()] = cmd.Validate
	}

	h.serve = h.Exec
	for i := len(middlewares) - 1; i >= 0; i-- {
		h.serve = middlewares[i](h.serve)
	}

	return &h
}

type simpleHandler struct {
	executor  map[string]HandlerFunc
	validator map[string]func(*resp.Command) error
	serve     HandlerFunc
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

	return h.serve(ctx, cmd)
}

func (h *simpleHandler) Exec(ctx context.Context, cmd *resp.Command) (resp.Value, error) {
	if v, ok := h.validator[cmd.Command()]; !ok {
		err := fmt.Errorf("%w: unknown command: %s", ErrClient, cmd.Command())
		return resp.NewSimpleError(err), err
	} else if err := v(cmd); err != nil {
		err := fmt.Errorf("%w: %w", ErrClient, err)
		return resp.NewSimpleError(err), err
	}

	return h.executor[cmd.Command()](ctx, cmd)
}
