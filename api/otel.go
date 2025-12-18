package api

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

type ExporterType string

const (
	StdoutExporter   ExporterType = "stdout"
	OtlpGrpcExporter ExporterType = "otlp_grpc"
)

type OtelConfig struct {
	TraceExporter   ExporterType
	TraceEnabled    bool
	MetricsExporter ExporterType
	MetricsEnabled  bool
	LoggerEnabled   bool
}

func SetupOtel(ctx context.Context, cfg OtelConfig) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error
	traceExporter := cfg.TraceExporter
	metricsExporter := cfg.MetricsExporter

	if traceExporter == "" {
		traceExporter = StdoutExporter
	}

	if metricsExporter == "" {
		metricsExporter = StdoutExporter
	}

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
	if cfg.TraceEnabled {
		tracerProvider, err := newTracerProvider(ctx, traceExporter)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)
	}

	// Set up meter provider.
	if cfg.MetricsEnabled {
		meterProvider, err := newMeterProvider(ctx, metricsExporter)
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)
	}

	// Set up logger provider.
	// loggerProvider, err := newLoggerProvider()
	// if err != nil {
	// 	handleErr(err)
	// 	return shutdown, err
	// }
	// shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	// global.SetLoggerProvider(loggerProvider)

	return shutdown, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context, exporter ExporterType) (*trace.TracerProvider, error) {
	var tracerExporter trace.SpanExporter

	if exporter == OtlpGrpcExporter {
		grpcExporter, err := otlptracegrpc.New(ctx)

		if err != nil {
			return nil, err
		}
		tracerExporter = grpcExporter
	} else {
		stdoutExporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, err
		}

		tracerExporter = stdoutExporter
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(tracerExporter,
			trace.WithBatchTimeout(time.Second*5)),
	)
	return tracerProvider, nil
}

func newMeterProvider(ctx context.Context, exporter ExporterType) (*metric.MeterProvider, error) {
	var metricExporter metric.Exporter

	if exporter == OtlpGrpcExporter {
		grpcExporter, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return nil, err
		}
		metricExporter = grpcExporter
	} else {
		stdoutExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		if err != nil {
			return nil, err
		}

		metricExporter = stdoutExporter
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
	)
	return meterProvider, nil
}

func newLoggerProvider() (*log.LoggerProvider, error) {

	logExporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)
	return loggerProvider, nil
}
