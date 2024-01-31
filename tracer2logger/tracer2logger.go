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
	"os"
	"path/filepath"
)

var (
	global = setMyLogger()
)

type (
	MyLogger struct {
		logger      *zap.Logger
		writeSyncer zapcore.WriteSyncer
		tp          *sdktrace.TracerProvider
		structured  bool
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
	}

	MTracer struct {
		serviceName string
		ctx         context.Context
	}
)

// setMyLogger returns a new MyLogger instance.
func setMyLogger() *MyLogger {
	// Create a new zap logger
	logger, _ := zap.NewProduction()

	return &MyLogger{
		logger:      logger,
		writeSyncer: zapcore.AddSync(os.Stdout),
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

	if cfg.logDir == "" {
		wd, _ := os.Getwd()
		cfg.logDir = filepath.Join(wd, "logs")
	}
}

// lumberjackSetup returns a new lumberjack logger.
func lumberjackSetup(cfg *Config) *lumberjack.Logger {
	// create the file path with the service name
	filePath := filepath.Join(cfg.logDir, fmt.Sprintf("%s.log", cfg.serviceName))

	return &lumberjack.Logger{
		Filename:   filePath,
		MaxAge:     cfg.maxAge,
		Compress:   cfg.compress,
		LocalTime:  cfg.localTime,
		MaxBackups: cfg.maxBackups,
		MaxSize:    cfg.maxFileSize,
	}
}

// NewMyLogger returns a new default config.
func NewMyLogger(cfg *Config) error {
	checkDefaults(cfg) // check if any default value is missing

	if !cfg.stdout {
		if _, err := os.Stat(cfg.logDir); os.IsNotExist(err) {
			if err = os.MkdirAll(cfg.logDir, 0755); err != nil {
				return fmt.Errorf("failed to create log directory: %s", cfg.logDir)
			}
		}

		global.writeSyncer = zapcore.AddSync(lumberjackSetup(cfg)) // change to write syncer
	}

	if cfg.tracing {
		if err := initTracer(); err != nil {
			return err
		}
	}

	encoderCfg := zap.NewProductionEncoderConfig()     // default encoder config
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder // change the time format
	encoder := zapcore.NewJSONEncoder(encoderCfg)      // default encoder is JSON

	if !cfg.structured {
		global.structured = !cfg.structured
		encoder = zapcore.NewConsoleEncoder(encoderCfg) // change the encoder to console
	}

	global.logger = zap.New(zapcore.NewCore(encoder, global.writeSyncer, zapcore.InfoLevel)) // change the logger

	return nil
}

// initTracer creates a new tracing instance.
func initTracer() error {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return fmt.Errorf("failed to initialize stdouttrace exporter %v\n", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp)

	global.tp = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(global.tp)
	return nil
}

// SetTracer returns a new tracer provider.
func SetTracer(serviceName string) (*MTracer, error) {
	if global.tp == nil {
		return nil, fmt.Errorf("tracer provider is nil")
	}

	ctx, _ := global.tp.Tracer(serviceName).Start(context.Background(), serviceName)

	return &MTracer{
		ctx:         ctx,
		serviceName: serviceName,
	}, nil
}

// Shutdown the tracer provider.
func (tp *MTracer) Shutdown() error {
	return global.tp.Shutdown(tp.ctx)
}

// addTraceID adds traceID and spanID to the logger.
func (tp *MTracer) addTraceID(msg string, fields []zap.Field) []zap.Field {
	if global.tp != nil {
		traceID := ""
		spanID := ""
		if span := trace.SpanContextFromContext(tp.ctx); span.IsValid() {
			traceID = span.TraceID().String()
			spanID = span.SpanID().String()
		}

		fields = append([]zap.Field{
			zap.String("traceName", tp.serviceName),
			zap.String("traceID", traceID),
			zap.String("spanID", spanID),
			zap.String("M", msg),
		}, fields...)

		//fields = append(fields, zap.String("traceID", traceID), zap.String("spanID", spanID), zap.String("traceName", tp.serviceName))
	}

	return fields
}

// Info logs a message at InfoLevel. The message includes any fields passed at the log site, as well as any fields accumulated on the logger.
func (tp *MTracer) Info(msg string, fields ...zap.Field) {
	global.logger.Info("", tp.addTraceID(msg, fields)...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed at the log site, as well as any fields accumulated on the logger.
func (tp *MTracer) Warn(msg string, fields ...zap.Field) {
	global.logger.Warn("", tp.addTraceID(msg, fields)...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed at the log site, as well as any fields accumulated on the logger.
func (tp *MTracer) Error(msg string, fields ...zap.Field) {
	global.logger.Error("", tp.addTraceID(msg, fields)...)
}

// Debug logs a message at DebugLevel. The message includes any fields passed at the log site, as well as any fields accumulated on the logger.
func (tp *MTracer) Debug(msg string, fields ...zap.Field) {
	global.logger.Debug("", tp.addTraceID(msg, fields)...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed at the log site, as well as any fields accumulated on the logger.
func (tp *MTracer) Fatal(msg string, fields ...zap.Field) {
	global.logger.Fatal("", tp.addTraceID(msg, fields)...)
}

//import (
//	"context"
//	"fmt"
//	"go.uber.org/zap"
//	"go.uber.org/zap/zapcore"
//	"gopkg.in/natefinch/lumberjack.v2"
//	"os"
//	"path/filepath"
//
//	"go.opentelemetry.io/otel"
//	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
//	sdktrace "go.opentelemetry.io/otel/sdk/trace"
//	"go.opentelemetry.io/otel/trace"
//)
//
//var loggerTracer *LoggerTracer
//
//type (
//	ctxKey struct{}
//
//	LoggerTracer struct {
//		trace.Tracer
//		*zap.Logger
//		*Config
//	}
//
//	Config struct {
//		logDir       string
//		serviceName  string
//		maxFileSize  int
//		maxAge       int
//		maxBackups   int
//		localTime    bool
//		compress     bool
//		stdout       bool
//		rotateByDate bool
//		tracing      bool
//		structured   bool
//	}
//
//	Tracing struct {
//		Span
//		trace.Tracer
//		context.Context
//	}
//
//	Span struct {
//		trace.Span
//		context.Context
//	}
//)
//
//// checkDefaults checks if any default value is missing.
//func checkDefaults(cfg *Config) {
//	if cfg.maxFileSize == 0 {
//		cfg.maxFileSize = 100
//	}
//
//	if cfg.maxAge == 0 {
//		cfg.maxAge = 28
//	}
//
//	if cfg.maxBackups == 0 {
//		cfg.maxBackups = 7
//	}
//
//	if cfg.localTime == false {
//		cfg.localTime = true
//	}
//
//	if cfg.compress == false {
//		cfg.compress = true
//	}
//
//	if cfg.rotateByDate == false {
//		cfg.rotateByDate = true
//	}
//
//	if cfg.logDir == "" {
//		wd, _ := os.Getwd()
//		cfg.logDir = filepath.Join(wd, "logs")
//	}
//}
//
//// lumberjackSetup returns a new lumberjack logger.
//func lumberjackSetup(cfg *Config) *lumberjack.Logger {
//	filePath := filepath.Join(cfg.logDir, fmt.Sprintf("%s.log", cfg.serviceName))
//
//	return &lumberjack.Logger{
//		MaxAge:     cfg.maxAge,
//		Compress:   cfg.compress,
//		Filename:   filePath,
//		LocalTime:  cfg.localTime,
//		MaxBackups: cfg.maxBackups,
//		MaxSize:    cfg.maxFileSize,
//	}
//}
//
//// NewLoggerTracer returns a new default config.
//func NewLoggerTracer(cfg *Config) error {
//	checkDefaults(cfg) // check if any default value is missing
//
//	writeSyncer := zapcore.AddSync(os.Stdout)
//
//	if !cfg.stdout {
//		writeSyncer = zapcore.AddSync(lumberjackSetup(cfg))
//
//		if _, err := os.Stat(cfg.logDir); os.IsNotExist(err) {
//			if err = os.MkdirAll(cfg.logDir, 0755); err != nil {
//				return fmt.Errorf("failed to create log directory: %s", cfg.logDir)
//			}
//		}
//	}
//
//	encoderCfg := zap.NewProductionEncoderConfig()
//	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
//	encoder := zapcore.NewJSONEncoder(encoderCfg)
//
//	if !cfg.structured {
//		encoder = zapcore.NewConsoleEncoder(encoderCfg)
//	}
//
//	logger := zap.New(zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel))
//
//	loggerTracer = &LoggerTracer{Config: cfg, Logger: logger}
//
//	return nil
//}
//
//// NewTracer creates a new tracing instance.
//func (lt *LoggerTracer) NewTracer(serviceName string) (*Tracing, error) {
//	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
//	if err != nil {
//		return nil, fmt.Errorf("failed to initialize stdouttrace exporter %v\n", err)
//	}
//
//	bsp := sdktrace.NewBatchSpanProcessor(exp)
//
//	tp := sdktrace.NewTracerProvider(
//		sdktrace.WithSampler(sdktrace.AlwaysSample()),
//		sdktrace.WithSpanProcessor(bsp),
//	)
//
//	otel.SetTracerProvider(tp)
//
//	if err != nil {
//		return nil, err
//	}
//
//	tc := tp.Tracer(serviceName)
//	ctx := context.WithValue(context.Background(), ctxKey{}, tc)
//
//	return &Tracing{Tracer: tc, Context: ctx}, nil
//}
//
//// Start a new span with the given name and return a new context with the span in it.
//func (t *Tracing) Start(name string) {
//	ss, _ := t.Context.Value(ctxKey{}).(trace.Tracer)
//	ctx, span := ss.Start(t.Context, name)
//
//	t.Span = Span{Span: span, Context: ctx}
//}
//
//// End the span.
//func (t *Tracing) End() {
//	t.Span.End()
//}
//
//// Fatal logs a message at fatal level.
//func Fatal(format string, fields ...any) {
//	loggerTracer.Fatal(fmt.Sprintf(format, fields...))
//}
//
//// Debug logs a message at debug level.
//func Debug(format string, fields ...any) {
//	loggerTracer.Debug(fmt.Sprintf(format, fields...))
//}
//
//// Warn logs a message at warn level.
//func Warn(format string, fields ...any) {
//	loggerTracer.Warn(fmt.Sprintf(format, fields...))
//}
//
//// Error logs a message at error level.
//func Error(format string, fields ...any) {
//	loggerTracer.Error(fmt.Sprintf(format, fields...))
//}
//
//// Info logs a message at info level.
//func Info(format string, fields ...any) {
//	loggerTracer.Info(fmt.Sprintf(format, fields...))
//}
//
////func (t *Tracing) FromContext(ctx context.Context) *zap.Logger {
////	childLogger, _ := ctx.Value(ctxKey{}).(*zap.Logger)
////
////	if traceID := trace.SpanFromContext(ctx).SpanContext().TraceID(); traceID.IsValid() {
////		childLogger = childLogger.With(zap.String("trace-id", traceID.String()))
////	}
////
////	if spanID := trace.SpanFromContext(ctx).SpanContext().SpanID(); spanID.IsValid() {
////		childLogger = childLogger.With(zap.String("span-id", spanID.String()))
////	}
////
////	return childLogger
////}
////
////func (lt *LoggerTracer) FromContext(ctx context.Context) *zap.Logger {
////	childLogger, _ := ctx.Value(ctxKey{}).(*zap.Logger)
////
////	if traceID := trace.SpanFromContext(ctx).SpanContext().TraceID(); traceID.IsValid() {
////		childLogger = childLogger.With(zap.String("trace-id", traceID.String()))
////	}
////
////	if spanID := trace.SpanFromContext(ctx).SpanContext().SpanID(); spanID.IsValid() {
////		childLogger = childLogger.With(zap.String("span-id", spanID.String()))
////	}
////
////	return childLogger
////}
////
////func (lt *LoggerTracer) NewContext(parent context.Context, z *zap.Logger) context.Context {
////	return context.WithValue(parent, ctxKey{}, z)
////}
////
////func (lt *LoggerTracer) WithTraceID(traceID string) *zap.Logger {
////	return lt.Logger.With(zap.String("trace-id", traceID))
////}
