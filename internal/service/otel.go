package service

import (
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTracerProvider(ctx)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(ctx)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	// Bridge the standard slog logger to OpenTelemetry while preserving its
	// existing console output.
	defaultLogger := slog.Default()
	defaultLogWriter := log.Writer()
	defaultLogFlags := log.Flags()
	defaultLogPrefix := log.Prefix()
	slog.SetDefault(newSlogLogger(os.Stderr, loggerProvider))
	shutdownFuncs = append(shutdownFuncs, func(context.Context) error {
		slog.SetDefault(defaultLogger)
		log.SetOutput(defaultLogWriter)
		log.SetFlags(defaultLogFlags)
		log.SetPrefix(defaultLogPrefix)
		return nil
	})

	return shutdown, err
}

func newSlogLogger(consoleWriter io.Writer, loggerProvider otellog.LoggerProvider) *slog.Logger {
	consoleHandler := slog.NewTextHandler(consoleWriter, nil)
	otelHandler := otelslog.NewHandler(
		"olivine",
		otelslog.WithLoggerProvider(loggerProvider),
	)
	return slog.New(slog.NewMultiHandler(consoleHandler, otelHandler))
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context) (*sdktrace.TracerProvider, error) {
	spanExporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, err
	}
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(spanExporter),
	)
	return tracerProvider, nil
}

func newMeterProvider(ctx context.Context) (*sdkmetric.MeterProvider, error) {
	metricReader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(metricReader),
	)
	return meterProvider, nil
}

func newLoggerProvider(ctx context.Context) (*sdklog.LoggerProvider, error) {
	logExporter, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		return nil, err
	}

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}
