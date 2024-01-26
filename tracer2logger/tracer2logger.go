package tracer2logger

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type (
	ctxKey struct{}

	LoggerTracer struct {
		trace.Tracer
		*zap.Logger
		*Config
	}

	Config struct {
		logDir       string
		serviceName  string
		maxFileSize  int
		maxAge       int
		maxBackups   int
		localTime    bool
		compress     bool
		stdout       bool
		rotateByDate bool
		tracing      bool
		structured   bool
		filename     string
		instance     string
		logger       *lumberjack.Logger
	}

	Tracing struct {
		Span
		trace.Tracer
		context.Context
	}

	Span struct {
		trace.Span
		context.Context
	}
)

// NewLoggerTracer returns a new default config.
func NewLoggerTracer(cfg *Config) (*LoggerTracer, error) {
	writeSyncer := zapcore.AddSync(os.Stdout)

	if cfg.logger != nil {
		writeSyncer = zapcore.AddSync(cfg.logger)
		if _, err := os.Stat(cfg.logDir); os.IsNotExist(err) {
			if err = os.MkdirAll(cfg.logDir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create log directory: %s", cfg.logDir)
			}
		}
	}

	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	if !cfg.structured {
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	encoder := zapcore.NewJSONEncoder(encoderCfg)

	if !cfg.structured {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	logger := zap.New(zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel))

	return &LoggerTracer{Config: cfg, Logger: logger}, nil
}

// NewTracing creates a new tracing instance.
func (lt *LoggerTracer) NewTracing(serviceName string) (*Tracing, error) {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize stdouttrace exporter %v\n", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tp)

	if err != nil {
		return nil, err
	}

	tc := tp.Tracer(serviceName)
	ctx := context.WithValue(context.Background(), ctxKey{}, tc)

	return &Tracing{Tracer: tc, Context: ctx}, nil
}

// Start a new span with the given name and return a new context with the span in it.
func (t *Tracing) Start(name string) {
	ss, _ := t.Context.Value(ctxKey{}).(trace.Tracer)
	ctx, span := ss.Start(t.Context, name)

	t.Span = Span{Span: span, Context: ctx}
}

// End the span.
func (t *Tracing) End() {
	t.Span.End()
}

func (t *Tracing) FromContext(ctx context.Context) *zap.Logger {
	childLogger, _ := ctx.Value(ctxKey{}).(*zap.Logger)

	if traceID := trace.SpanFromContext(ctx).SpanContext().TraceID(); traceID.IsValid() {
		childLogger = childLogger.With(zap.String("trace-id", traceID.String()))
	}

	if spanID := trace.SpanFromContext(ctx).SpanContext().SpanID(); spanID.IsValid() {
		childLogger = childLogger.With(zap.String("span-id", spanID.String()))
	}

	return childLogger
}

func (lt *LoggerTracer) FromContext(ctx context.Context) *zap.Logger {
	childLogger, _ := ctx.Value(ctxKey{}).(*zap.Logger)

	if traceID := trace.SpanFromContext(ctx).SpanContext().TraceID(); traceID.IsValid() {
		childLogger = childLogger.With(zap.String("trace-id", traceID.String()))
	}

	if spanID := trace.SpanFromContext(ctx).SpanContext().SpanID(); spanID.IsValid() {
		childLogger = childLogger.With(zap.String("span-id", spanID.String()))
	}

	return childLogger
}

func (lt *LoggerTracer) NewContext(parent context.Context, z *zap.Logger) context.Context {
	return context.WithValue(parent, ctxKey{}, z)
}

func (lt *LoggerTracer) WithTraceID(traceID string) *zap.Logger {
	return lt.Logger.With(zap.String("trace-id", traceID))
}
