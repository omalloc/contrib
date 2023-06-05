package tracing

import (
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type TracingConf interface {
	GetEndpoint() string
	GetCustomName() string
}

type traceConfig struct {
	endpoint    string
	serviceName string
	fraction    float64
}

type Option func(*traceConfig)

func InitTracer(opts ...Option) {
	host, _ := os.Hostname()
	c := &traceConfig{
		fraction: 1.0, // rate based on the parent span to 100%
	}
	for _, opt := range opts {
		opt(c)
	}

	if c.endpoint == "" {
		return
	}

	if c.serviceName == "" {
		c.serviceName = host
	}

	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(c.endpoint)))
	if err != nil {
		return
	}
	tp := tracesdk.NewTracerProvider(
		// Set the sampling rate based on the `tracingConfig`
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(c.fraction))),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(c.serviceName),
			attribute.String("host", host),
		)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)
}

// WithEndpoint ... 自定义 jaeger 入口
func WithEndpoint(endpoint string) Option {
	return func(tr *traceConfig) {
		tr.endpoint = endpoint
	}
}

// WithServiceName ...自定义上报服务名
func WithServiceName(name string) Option {
	return func(tr *traceConfig) {
		tr.serviceName = name
	}
}

// WithRatioBased ... 设置采样比率
func WithRatioBased(fraction float64) Option {
	return func(tr *traceConfig) {
		tr.fraction = fraction
	}
}
