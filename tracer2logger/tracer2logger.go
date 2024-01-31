package tracer2logger

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	global = loggerGlobal()
)

type (
	LoggerGlobal struct {
		endCh    chan struct{}
		logger   *zap.Logger
		syncer   zapcore.WriteSyncer
		provider *sdktrace.TracerProvider
		ctx      context.Context
	}

	Config struct {
		LogDir       string
		ServiceName  string
		Tracing      bool
		Structured   bool
		maxFileSize  int
		maxAge       int
		maxBackups   int
		localTime    bool
		compress     bool
		stdout       bool
		rotateByDate bool
	}

	LogTracer struct {
		name string
		ctx  context.Context
		span trace.Span
	}
)

func (lg *LoggerGlobal) cleanup() {
	if err := lg.provider.Shutdown(lg.ctx); err != nil {
		log.Fatalf("zap logger sync error: %s", err.Error())
	}

	if err := lg.logger.Sync(); err != nil {
		log.Fatalf("zap logger sync error: %s", err.Error())
	}
}

// shutdown shuts down TracerProvider. All registered span processors are shut down
// in the order they were registered and any held computational resources are released.
// After Shutdown is called, all methods are no-ops.
func (lg *LoggerGlobal) shutdown() {
	go func() {
		defer lg.cleanup()
		for {
			select {
			case <-lg.ctx.Done():
				break
			}
		}
	}()
}

// loggerGlobal returns a new LoggerGlobal instance.
func loggerGlobal() *LoggerGlobal {
	return &LoggerGlobal{
		logger: zap.NewNop(),
		syncer: zapcore.AddSync(io.Discard),
		endCh:  make(chan struct{}),
	}
}

// checkDefaults checks if any default value is missing.
func checkDefaults(cfg *Config) {
	if cfg.maxFileSize == 0 {
		cfg.maxFileSize = 100
	}

	if cfg.maxAge == 0 {
		cfg.maxAge = 28
	}

	if cfg.maxBackups == 0 {
		cfg.maxBackups = 7
	}

	if cfg.localTime == false {
		cfg.localTime = true
	}

	if cfg.compress == false {
		cfg.compress = true
	}

	if cfg.rotateByDate == false {
		cfg.rotateByDate = true
	}

	if cfg.LogDir == "" {
		wd, _ := os.Getwd()
		cfg.LogDir = filepath.Join(wd, "logs")
	}
}

// lumberjackSetup returns a new lumberjack logger.
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
		MaxAge:     cfg.maxAge,
		Compress:   cfg.compress,
		LocalTime:  cfg.localTime,
		MaxBackups: cfg.maxBackups,
		MaxSize:    cfg.maxFileSize,
	})

	return file, nil
}

// NewMyLogger returns a new default config.
func NewMyLogger(cfg *Config) error {
	stdout := zapcore.AddSync(os.Stdout)

	file, err := lumberjackSetup(cfg)
	if err != nil {
		return err
	}

	if cfg.Tracing {
		exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return fmt.Errorf("failed to initialize stdouttrace exporter %v\n", err)
		}

		global.provider = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(exporter)),
		)

		global.ctx = context.Background()
		otel.SetTracerProvider(global.provider)
	}

	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	global.logger = zap.New(zapcore.NewTee(
		zapcore.NewCore(fileEncoder, file, level),
		zapcore.NewCore(consoleEncoder, stdout, level),
	))

	global.shutdown()

	return nil
}

// NewTracer returns a Tracer with the given name and options. If a Tracer for
// the given name and options does not exist it is created, otherwise the
// existing Tracer is returned.
//
// If name is empty, DefaultTracerName is used instead.
//
// This method is safe to be called concurrently.
func NewTracer(serviceName string) (*LogTracer, error) {
	if global.provider == nil {
		return nil, fmt.Errorf("tracer provider is nil")
	}

	ctxChild, spanChild := global.provider.Tracer(serviceName).Start(global.ctx, serviceName)

	return &LogTracer{
		ctx:  ctxChild,
		span: spanChild,
		name: serviceName,
	}, nil
}

// End completes the Span. The Span is considered complete and ready to be
// delivered through the rest of the telemetry pipeline after this method
// is called. Therefore, updates to the Span are not allowed after this
// method has been called.
func (tp *LogTracer) End() {
	tp.span.End()
}

// addTraceID adds traceID and spanID to the logger.
func (tp *LogTracer) addTraceID(msg string, fields []zap.Field) []zap.Field {
	if global.provider != nil {
		traceID := ""
		spanID := ""
		if span := trace.SpanContextFromContext(tp.ctx); span.IsValid() {
			traceID = span.TraceID().String()
			spanID = span.SpanID().String()
		}

		fields = append([]zap.Field{
			zap.String("traceName", tp.name),
			zap.String("traceID", traceID),
			zap.String("spanID", spanID),
			zap.String("M", msg),
		}, fields...)
	}

	return fields
}

// Info logs a message at InfoLevel. The message includes any fields passed at the log site, as well as any fields accumulated on the logger.
func (tp *LogTracer) Info(msg string, fields ...zap.Field) {
	global.logger.Info("", tp.addTraceID(msg, fields)...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed at the log site, as well as any fields accumulated on the logger.
func (tp *LogTracer) Warn(msg string, fields ...zap.Field) {
	global.logger.Warn("", tp.addTraceID(msg, fields)...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed at the log site, as well as any fields accumulated on the logger.
func (tp *LogTracer) Error(msg string, fields ...zap.Field) {
	global.logger.Error("", tp.addTraceID(msg, fields)...)
}

// Debug logs a message at DebugLevel. The message includes any fields passed at the log site, as well as any fields accumulated on the logger.
func (tp *LogTracer) Debug(msg string, fields ...zap.Field) {
	global.logger.Debug("", tp.addTraceID(msg, fields)...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed at the log site, as well as any fields accumulated on the logger.
func (tp *LogTracer) Fatal(msg string, fields ...zap.Field) {
	global.logger.Fatal("", tp.addTraceID(msg, fields)...)
}
