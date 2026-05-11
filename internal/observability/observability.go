// Package observability bootstraps the OpenTelemetry SDK.
//
// A single call to Init wires global TracerProvider, MeterProvider and
// LoggerProvider that ship traces, metrics and logs over OTLP/gRPC to the
// Collector configured via OTEL_EXPORTER_OTLP_ENDPOINT. The returned Shutdown
// function should be invoked at process exit so pending batches are flushed.
package observability

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	otelmetric "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	oteltrace "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	logglobal "go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Options groups the inputs required to wire OTel.
type Options struct {
	Endpoint       string
	Insecure       bool
	ServiceName    string
	ServiceVersion string
	Environment    string
}

// Shutdown gracefully flushes all OTel providers.
type Shutdown func(context.Context) error

// Init initialises the OTel SDK and returns a Shutdown to be deferred from
// cmd/main.go. Returning an error from Init means the application could not
// be observed and should not start.
func Init(ctx context.Context, opts Options) (Shutdown, error) {
	res, err := buildResource(ctx, opts)
	if err != nil {
		return nil, err
	}

	tracerShutdown, err := initTracer(ctx, opts, res)
	if err != nil {
		return nil, fmt.Errorf("observability: tracer: %w", err)
	}
	meterShutdown, err := initMeter(ctx, opts, res)
	if err != nil {
		_ = tracerShutdown(ctx)
		return nil, fmt.Errorf("observability: meter: %w", err)
	}
	loggerShutdown, err := initLogger(ctx, opts, res)
	if err != nil {
		_ = meterShutdown(ctx)
		_ = tracerShutdown(ctx)
		return nil, fmt.Errorf("observability: logger: %w", err)
	}

	return func(ctx context.Context) error {
		return errors.Join(
			loggerShutdown(ctx),
			meterShutdown(ctx),
			tracerShutdown(ctx),
		)
	}, nil
}

func buildResource(ctx context.Context, opts Options) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(opts.ServiceName),
		semconv.ServiceVersion(opts.ServiceVersion),
	}
	if opts.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironment(opts.Environment))
	}
	res, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("observability: resource: %w", err)
	}
	return res, nil
}

func initTracer(ctx context.Context, opts Options, res *resource.Resource) (Shutdown, error) {
	exp, err := oteltrace.New(ctx, traceOptions(opts)...)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

func initMeter(ctx context.Context, opts Options, res *resource.Resource) (Shutdown, error) {
	exp, err := otelmetric.New(ctx, metricOptions(opts)...)
	if err != nil {
		return nil, err
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(15*time.Second))),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	return mp.Shutdown, nil
}

func initLogger(ctx context.Context, opts Options, res *resource.Resource) (Shutdown, error) {
	exp, err := otellog.New(ctx, logOptions(opts)...)
	if err != nil {
		return nil, err
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)),
		sdklog.WithResource(res),
	)
	logglobal.SetLoggerProvider(lp)
	return lp.Shutdown, nil
}

func traceOptions(opts Options) []oteltrace.Option {
	out := []oteltrace.Option{oteltrace.WithEndpoint(opts.Endpoint)}
	if opts.Insecure {
		out = append(out, oteltrace.WithInsecure())
	}
	return out
}

func metricOptions(opts Options) []otelmetric.Option {
	out := []otelmetric.Option{otelmetric.WithEndpoint(opts.Endpoint)}
	if opts.Insecure {
		out = append(out, otelmetric.WithInsecure())
	}
	return out
}

func logOptions(opts Options) []otellog.Option {
	out := []otellog.Option{otellog.WithEndpoint(opts.Endpoint)}
	if opts.Insecure {
		out = append(out, otellog.WithInsecure())
	}
	return out
}
