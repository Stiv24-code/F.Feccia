package telemetry

import (
	"context"
	"os"
	"strings"
	"time"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
)

type Shutdown func(context.Context) error

type TelemetryProviders struct {
	TraceProvider *sdktrace.TracerProvider
	MeterProvider *sdkmetric.MeterProvider
	Propagator    propagation.TextMapPropagator
}

// Init sets up OpenTelemetry tracing and metrics with OTLP gRPC exporters.
// Env vars:
// - OTEL_EXPORTER_OTLP_ENDPOINT (default: localhost:4317)
// - OTEL_EXPORTER_OTLP_INSECURE (default: true)
// - OTEL_SERVICE_NAME (default: fratelli-feccia)
func Init(ctx context.Context) (TelemetryProviders, func(context.Context) error, error) {
	endpoint := normalizeOTLPEndpoint(firstNonEmpty(
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		os.Getenv("OTLP_TRACE_ENDPOINT"),   // legacy fallback
		os.Getenv("OTLP_METRICS_ENDPOINT"), // legacy fallback
		"localhost:4317",
	))

	insecureConn := parseBoolWithDefault(firstNonEmpty(
		os.Getenv("OTEL_EXPORTER_OTLP_INSECURE"),
		os.Getenv("OTLP_INSECURE"), // legacy fallback
	), true)
	serviceName := firstNonEmpty(os.Getenv("OTEL_SERVICE_NAME"), "fratelli-feccia")

	dialOpts := []grpc.DialOption{}
	if insecureConn {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return TelemetryProviders{}, nil, err
	}

	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithDialOption(dialOpts...),
	)
	if err != nil {
		return TelemetryProviders{}, nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExp,
			sdktrace.WithMaxExportBatchSize(512),
			sdktrace.WithBatchTimeout(2*time.Second),
		),
	)

	metricExp, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithDialOption(dialOpts...),
	)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return TelemetryProviders{}, nil, err
	}
	reader := sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(5*time.Second))
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	prop := propagation.TraceContext{}
	otel.SetTextMapPropagator(prop)

	shutdown := func(c context.Context) error {
		var firstErr error
		if err := mp.Shutdown(c); err != nil {
			firstErr = err
		}
		if err := tp.Shutdown(c); err != nil && firstErr == nil {
			firstErr = err
		}
		return firstErr
	}

	slog.Info("OpenTelemetry initialized", "endpoint", endpoint, "insecure", insecureConn)
	return TelemetryProviders{TraceProvider: tp, MeterProvider: mp, Propagator: prop}, shutdown, nil
}

// NewCounter helper to increment custom metrics
func NewCounter(name string) metric.Int64Counter {
	m := otel.GetMeterProvider().Meter("fratelli-feccia")
	c, _ := m.Int64Counter(name)
	return c
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseBoolWithDefault(value string, defaultValue bool) bool {
	if value == "" {
		return defaultValue
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return defaultValue
	}
}

// normalizeOTLPEndpoint converts values like:
// - http://localhost:4317
// - https://collector:4317
// to host:port expected by OTLP gRPC exporters.
func normalizeOTLPEndpoint(endpoint string) string {
	e := strings.TrimSpace(endpoint)
	e = strings.TrimPrefix(e, "http://")
	e = strings.TrimPrefix(e, "https://")
	e = strings.TrimSuffix(e, "/")
	return e
}

// NewFiberMiddleware returns Fiber middleware preconfigured with the given providers
func NewFiberMiddleware(p TelemetryProviders) fiber.Handler {
	return otelfiber.Middleware(
		otelfiber.WithTracerProvider(p.TraceProvider),
		otelfiber.WithPropagators(p.Propagator),
		otelfiber.WithSpanNameFormatter(func(c *fiber.Ctx) string {
			path := c.Path()
			if r := c.Route(); r != nil && r.Path != "" {
				path = r.Path
			}
			return c.Method() + " " + path
		}),
	)
}

// NewFiberMetricsMiddleware records basic HTTP metrics for each request.
// Metrics:
// - http_server_requests_total (counter)
// - http_server_request_duration_ms (histogram)
func NewFiberMetricsMiddleware() fiber.Handler {
	meter := otel.GetMeterProvider().Meter("fratelli-feccia/http")
	requestsCounter, _ := meter.Int64Counter("http_server_requests_total")
	latencyHistogram, _ := meter.Float64Histogram("http_server_request_duration_ms")

	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		durationMs := float64(time.Since(start).Milliseconds())

		route := c.Path()
		if r := c.Route(); r != nil && r.Path != "" {
			route = r.Path
		}
		attrs := metric.WithAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.route", route),
			attribute.Int("http.status_code", c.Response().StatusCode()),
		)

		requestsCounter.Add(c.UserContext(), 1, attrs)
		latencyHistogram.Record(c.UserContext(), durationMs, attrs)
		return err
	}
}
