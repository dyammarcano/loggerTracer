package loggerTracing

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	Trace4U struct {
		childCtx context.Context
		span     trace.Span
		name     string
	}
)

// NewTracer returns a new Trace4U.
func NewTracer(serviceName string) *Trace4U {
	ctxChild, childSpan := globalProvider.Tracer(serviceName).Start(context.Background(), serviceName)

	return &Trace4U{
		childCtx: ctxChild,
		span:     childSpan,
		name:     serviceName,
	}
}

// End ends the span.
func (t *Trace4U) End() {
	t.span.End()
}

// getFields returns the fields with the context.
func (t *Trace4U) getFields(ctx context.Context, fields Fields) []Field {
	var zapFields []Field

	if fields == nil {
		return zapFields
	}

	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}

	_, newSpan := globalProvider.Tracer("logEntry").Start(ctx, "logEntry")

	zapFields = append(zapFields, zap.String("traceId", trace.SpanContextFromContext(ctx).TraceID().String()))
	zapFields = append(zapFields, zap.String("spanId", newSpan.SpanContext().SpanID().String()))
	zapFields = append(zapFields, zap.String("service", t.name))

	return zapFields
}

// Info logs an info message with context.
func (t *Trace4U) Info(msg string, fields Fields) {
	zapFields := t.getFields(t.childCtx, fields)
	zap.L().Info(msg, zapFields...)
}

// Error logs an error message with context.
func (t *Trace4U) Error(msg string, fields Fields) {
	zapFields := t.getFields(t.childCtx, fields)
	zap.L().Error(msg, zapFields...)
}

// Warn logs a warning message with context.
func (t *Trace4U) Warn(msg string, fields Fields) {
	zapFields := t.getFields(t.childCtx, fields)
	zap.L().Warn(msg, zapFields...)
}

// Debug logs a debug message with context.
func (t *Trace4U) Debug(msg string, fields Fields) {
	zapFields := t.getFields(t.childCtx, fields)
	zap.L().Debug(msg, zapFields...)
}

// Fatal logs a fatal message with context.
func (t *Trace4U) Fatal(msg string, fields Fields) {
	zapFields := t.getFields(t.childCtx, fields)
	zap.L().Fatal(msg, zapFields...)
}

// Panic logs a panic message with context.
func (t *Trace4U) Panic(msg string, fields Fields) {
	zapFields := t.getFields(t.childCtx, fields)
	zap.L().Panic(msg, zapFields...)
}
