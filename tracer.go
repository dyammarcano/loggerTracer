package loggerTracer

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	Trace4U struct {
		childCtx context.Context
		span     trace.Span
	}
)

// NewTracer returns a new Trace4U.
func NewTracer(serviceName string) *Trace4U {
	ctxChild, childSpan := tracerProvider.Tracer(serviceName).Start(context.Background(), serviceName)
	ctxChild = context.WithValue(ctxChild, "serviceName", serviceName)

	return &Trace4U{
		childCtx: ctxChild,
		span:     childSpan,
	}
}

// End ends the span.
func (t *Trace4U) End() {
	t.span.End()
}

// Info logs an info message with context.
func (t *Trace4U) Info(msg string, fields ...Entry) {
	zapFields := getFields(t.childCtx, fields...)
	zap.L().Info(msg, zapFields...)
}

// Error logs an error message with context.
func (t *Trace4U) Error(msg string, fields ...Entry) {
	zapFields := getFields(t.childCtx, fields...)
	zap.L().Error(msg, zapFields...)
}

// Warn logs a warning message with context.
func (t *Trace4U) Warn(msg string, fields ...Entry) {
	zapFields := getFields(t.childCtx, fields...)
	zap.L().Warn(msg, zapFields...)
}

// Debug logs a debug message with context.
func (t *Trace4U) Debug(msg string, fields ...Entry) {
	zapFields := getFields(t.childCtx, fields...)
	zap.L().Debug(msg, zapFields...)
}

// Fatal logs a fatal message with context.
func (t *Trace4U) Fatal(msg string, fields ...Entry) {
	zapFields := getFields(t.childCtx, fields...)
	zap.L().Fatal(msg, zapFields...)
}

// Panic logs a panic message with context.
func (t *Trace4U) Panic(msg string, fields ...Entry) {
	zapFields := getFields(t.childCtx, fields...)
	zap.L().Panic(msg, zapFields...)
}
