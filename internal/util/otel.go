package util

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"olivine/pkg/resp"
)

const tracerName = "olivine"

func SelectTracer(tracers []trace.Tracer) trace.Tracer {
	if len(tracers) != 0 {
		return tracerOrDefault(tracers[0])
	}
	return otel.Tracer(tracerName)
}

func tracerOrDefault(tracer trace.Tracer) trace.Tracer {
	if tracer != nil {
		return tracer
	}
	return otel.Tracer(tracerName)
}

func StartCommandSpan(
	ctx context.Context,
	tracer trace.Tracer,
	spanName string,
	command *resp.Command,
) (context.Context, trace.Span) {
	return tracerOrDefault(tracer).Start(
		ctx,
		spanName,
		trace.WithAttributes(
			attribute.String("olivine.command.name", command.Command()),
			attribute.Int("olivine.command.argument_count", len(command.Args())),
		),
	)
}

// EndSpan records err, when present, and ends span.
func EndSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
