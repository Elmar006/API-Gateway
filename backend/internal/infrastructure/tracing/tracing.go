// Package tracing wires the OpenTelemetry SDK with an OTLP/HTTP exporter so
// gateway traces can be shipped to any OTLP-compatible backend (Tempo, Jaeger,
// Honeycomb, …). Configuration follows the standard OpenTelemetry env-var
// scheme so no separate config knobs are needed.
package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// noopShutdown is returned when tracing is disabled and there is nothing to
// flush; callers can ignore it safely.
var noopShutdown = func(context.Context) error { return nil }

// Init initialises a global tracer with the named service. When `enabled` is
// false, a no-op tracer is installed (so callers can `tracer.Start` safely
// regardless of configuration).
//
// Returns a shutdown function the caller should defer to flush spans on exit.
func Init(ctx context.Context, serviceName string, enabled bool) (func(context.Context) error, error) {
	if !enabled {
		// Install a no-op so calls to otel.Tracer don't blow up when tracing
		// is disabled.
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
		return noopShutdown, nil
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return noopShutdown, fmt.Errorf("build resource: %w", err)
	}

	exporter, err := otlptrace.New(ctx, otlptracehttp.NewClient())
	if err != nil {
		return noopShutdown, fmt.Errorf("init otlp exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	return tp.Shutdown, nil
}

// SpanContextID returns the hex trace id of the span on the given context, or
// the empty string when none is recording.
func SpanContextID(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}
