package server

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"olivine/internal/service/cmd"
	"olivine/internal/util"
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

func NewHandler(tracer trace.Tracer, cmds []cmd.Command, middlewares ...Middleware) Handler {
	h := simpleHandler{
		tracer: tracer,
	}
	h.executor = make(map[string]HandlerFunc)
	for _, cmd := range cmds {
		h.executor[cmd.Command()] = cmd.Exec
	}

	h.serve = h.Exec
	for _, m := range slices.Backward(middlewares) {
		h.serve = m(h.serve)
	}

	return &h
}

type simpleHandler struct {
	tracer   trace.Tracer
	executor map[string]HandlerFunc
	serve    HandlerFunc
}

func (h *simpleHandler) ServeRESP(ctx context.Context, rd *resp.Reader) (ret resp.Value, err error) {
	ctx, span := h.tracer.Start(
		ctx,
		"server.(*simpleHandler).ServeRESP",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer func() {
		util.EndSpan(span, err)
	}()

	command, err := resp.ReadCommand(rd)
	if err != nil {
		if errors.Is(err, resp.ErrProtocol) {
			return resp.NewSimpleError(err), fmt.Errorf("%w: %w", ErrClient, err)
		} else {
			return nil, fmt.Errorf("%w: %w", ErrServer, err)
		}
	}

	span.SetAttributes(commandAttributes(command)...)
	return h.serve(ctx, command)
}

func (h *simpleHandler) Exec(ctx context.Context, command *resp.Command) (ret resp.Value, err error) {
	ctx, span := h.tracer.Start(
		ctx,
		"server.(*simpleHandler).Exec",
		trace.WithAttributes(commandAttributes(command)...),
	)
	defer func() { util.EndSpan(span, err) }()

	f, ok := h.executor[command.Command()]
	if !ok {
		err := fmt.Errorf("%w: unknown command: %s", ErrClient, command.Command())
		return resp.NewSimpleError(err), err
	}

	return f(ctx, command)
}

func commandAttributes(command *resp.Command) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("olivine.command.name", command.Command()),
		attribute.Int("olivine.command.argument_count", len(command.Args())),
	}
}
