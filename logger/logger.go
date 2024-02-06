package logger

import (
	"context"
	"fmt"
	"github.com/caarlos0/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
)

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

var (
	ll             *Log4U
	globalProvider *sdktrace.TracerProvider
)

type (
	Level int

	Field = zapcore.Field

	Fields map[string]any

	Logger interface {
		Info(msg string, fields ...Field)
		Error(msg string, fields ...Field)
		Warn(msg string, fields ...Field)
		Debug(msg string, fields ...Field)
		Fatal(msg string, fields ...Field)
		Panic(msg string, fields ...Field)
	}

	Trace interface {
		Info(ctx context.Context, msg string, fields ...Field)
		Error(ctx context.Context, msg string, fields ...Field)
		Warn(ctx context.Context, msg string, fields ...Field)
		Debug(ctx context.Context, msg string, fields ...Field)
		Fatal(ctx context.Context, msg string, fields ...Field)
		Panic(ctx context.Context, msg string, fields ...Field)
	}

	Trace4U struct {
		childCtx context.Context
		span     trace.Span
	}

	Log4U struct {
		*zap.Logger
		*Config
	}

	Config struct {
		rotateByDate bool
		structured   bool
		filename     string
		level        Level
		logger       *lumberjack.Logger

		LogDir      string
		ServiceName string
		MaxFileSize int
		MaxAge      int
		MaxBackups  int
		Stdout      bool
	}
)

// selectLevel returns the zapcore.Level based on the given Level.
func selectLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case DPanicLevel:
		return zapcore.DPanicLevel
	case PanicLevel:
		return zapcore.PanicLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// checkDefaults checks if any default value is missing.
func checkDefaults(cfg *Config) {
	if cfg.level == 0 {
		cfg.level = InfoLevel
	}

	if cfg.MaxFileSize == 0 {
		cfg.MaxFileSize = 100
	}

	if cfg.MaxAge == 0 {
		cfg.MaxAge = 28
	}

	if cfg.MaxBackups == 0 {
		cfg.MaxBackups = 7
	}

	if cfg.LogDir == "" {
		wd, _ := os.Getwd()
		cfg.LogDir = filepath.Join(wd, "logs")
	}
}

// lumberjackSetup returns a new lumberjack globalLogger.
func lumberjackSetup(cfg *Config) (zapcore.WriteSyncer, error) {
	checkDefaults(cfg) // check if any default value is missing

	if _, err := os.Stat(cfg.LogDir); os.IsNotExist(err) {
		if err = os.MkdirAll(cfg.LogDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %s", cfg.LogDir)
		}
	}

	// change to write syncer
	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(cfg.LogDir, fmt.Sprintf("%s.log", cfg.ServiceName)),
		Compress:   true,
		LocalTime:  true,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
		MaxSize:    cfg.MaxFileSize,
	})

	return file, nil
}

// NewLogger returns a new Log4U.
func NewLogger(cfg *Config) error {
	file, err := lumberjackSetup(cfg)
	if err != nil {
		return err
	}

	ll = &Log4U{
		Logger: zap.NewNop(),
		Config: cfg,
	}

	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return fmt.Errorf("failed to initialize stdouttrace exporter %v\n", err)
	}

	globalProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
	)

	otel.SetTracerProvider(globalProvider)

	level := zap.NewAtomicLevelAt(selectLevel(cfg.level))

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.MessageKey = "message"
	productionCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	productionCfg.EncodeCaller = zapcore.ShortCallerEncoder
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleEncoder := zapcore.NewJSONEncoder(productionCfg)

	ll.Logger = zap.New(zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, file, level),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
	))

	return nil
}

// SetLevel sets the globalLogger level.
func (l *Log4U) SetLevel(level Level) {
	l.Logger.Core().Enabled(selectLevel(level))
}

// Stop stops the globalLogger.
func (l *Log4U) Stop() {
	if err := l.Logger.Sync(); err != nil {
		log.Errorf("failed to sync globalLogger: %v", err)
	}
}

// Info logs an info message.
func (l *Log4U) Info(msg string, fields ...Field) {
	l.Logger.Info(msg, fields...)
}

// Error logs an error message.
func (l *Log4U) Error(msg string, fields ...Field) {
	l.Logger.Error(msg, fields...)
}

// Warn logs a warning message.
func (l *Log4U) Warn(msg string, fields ...Field) {
	l.Logger.Warn(msg, fields...)
}

// Debug logs a debug message.
func (l *Log4U) Debug(msg string, fields ...Field) {
	l.Logger.Debug(msg, fields...)
}

// Fatal logs a fatal message.
func (l *Log4U) Fatal(msg string, fields ...Field) {
	l.Logger.Fatal(msg, fields...)
}

// Panic logs a panic message.
func (l *Log4U) Panic(msg string, fields ...Field) {
	l.Logger.Panic(msg, fields...)
}

// With returns a new Log4U with the given fields.
func (l *Log4U) With(fields ...Field) *Log4U {
	return &Log4U{Logger: l.Logger.With(fields...)}
}

func NewTracer(serviceName string) *Trace4U {
	ctxChild, childSpan := globalProvider.Tracer(serviceName).Start(context.Background(), serviceName)

	return &Trace4U{
		childCtx: ctxChild,
		span:     childSpan,
	}
}

func (t *Trace4U) End() {
	t.span.End()
}

func (t *Trace4U) getFields(ctx context.Context, fields Fields) []Field {
	var zapFields []Field

	if fields == nil {
		return zapFields
	}

	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}

	zapFields = append(zapFields, zap.String("traceId", trace.SpanContextFromContext(ctx).TraceID().String()))

	_, newSpan := globalProvider.Tracer("logEntry").Start(ctx, "logEntry")
	zapFields = append(zapFields, zap.String("spanId", newSpan.SpanContext().SpanID().String()))

	return zapFields
}

// Info logs an info message with context.
func (t *Trace4U) Info(msg string, fields Fields) {
	zapFields := t.getFields(t.childCtx, fields)
	ll.Info(msg, zapFields...)
}

// Error logs an error message with context.
func (t *Trace4U) Error(msg string, fields Fields) {
	zapFields := t.getFields(t.childCtx, fields)
	ll.Error(msg, zapFields...)
}

// Warn logs a warning message with context.
func (t *Trace4U) Warn(msg string, fields Fields) {
	zapFields := t.getFields(t.childCtx, fields)
	ll.Warn(msg, zapFields...)
}

// Debug logs a debug message with context.
func (t *Trace4U) Debug(msg string, fields Fields) {
	zapFields := t.getFields(t.childCtx, fields)
	ll.Debug(msg, zapFields...)
}

// Fatal logs a fatal message with context.
func (t *Trace4U) Fatal(msg string, fields Fields) {
	zapFields := t.getFields(t.childCtx, fields)
	ll.Fatal(msg, zapFields...)
}

// Panic logs a panic message with context.
func (t *Trace4U) Panic(msg string, fields Fields) {
	zapFields := t.getFields(t.childCtx, fields)
	ll.Panic(msg, zapFields...)
}
