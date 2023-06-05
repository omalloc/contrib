package orm

import (
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	dbName         string
	tracerProvider trace.TracerProvider
	alwaysOmitVars bool
}

// TraceOption allows for managing options for the tracing plugin.
type TraceOption interface {
	apply(*config)
}

type traceOptionFunc func(*config)

func (o traceOptionFunc) apply(c *config) {
	o(c)
}

// WithTracerProvider sets the tracer provider to use for opentelemetry.
//
// If none is specified, the global provider is used.
func WithTracerProvider(provider trace.TracerProvider) TraceOption {
	return traceOptionFunc(func(c *config) {
		c.tracerProvider = provider
	})
}

// WithDatabaseName specified the database name to be used in span names
//
// since its not possible to extract this information from gorm
func WithDatabaseName(dbName string) TraceOption {
	return traceOptionFunc(func(c *config) {
		c.dbName = dbName
	})
}

// WithAlwaysOmitVariables will omit variables from the span attributes
func WithAlwaysOmitVariables() TraceOption {
	return traceOptionFunc(func(c *config) {
		c.alwaysOmitVars = true
	})
}
