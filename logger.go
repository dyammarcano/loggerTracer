package loggerTracing

import (
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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

var globalProvider *sdktrace.TracerProvider

type (
	Level  int
	Field  = zapcore.Field
	Fields map[string]any

	Config struct {
		Stdout       bool
		RotateByDate bool
		Unstructured bool
		LogDir       string
		ServiceName  string
		MaxFileSize  int
		MaxAge       int
		MaxBackups   int
		Level        Level
		logger       *lumberjack.Logger
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
	if cfg.Level == 0 {
		cfg.Level = InfoLevel
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
	var core []zapcore.Core

	file, err := lumberjackSetup(cfg)
	if err != nil {
		return err
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

	level := zap.NewAtomicLevelAt(selectLevel(cfg.Level))

	productionCfg := zap.NewProductionEncoderConfig()

	productionCfg.TimeKey = "timestamp"
	productionCfg.MessageKey = "message"
	productionCfg.EncodeCaller = zapcore.ShortCallerEncoder
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	consoleEncoder := zapcore.NewJSONEncoder(productionCfg)
	if cfg.Unstructured {
		consoleEncoder = zapcore.NewConsoleEncoder(productionCfg)
	}

	core = append(core, zapcore.NewCore(consoleEncoder, file, level))
	if cfg.Stdout {
		core = append(core, zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level))
	}

	zap.ReplaceGlobals(zap.New(zapcore.NewTee(core...)))

	return nil
}

// Info logs an info message.
func Info(msg string, fields ...Field) {
	zap.L().Info(msg, fields...)
}

// Error logs an error message.
func Error(msg string, fields ...Field) {
	zap.L().Error(msg, fields...)
}

// Warn logs a warning message.
func Warn(msg string, fields ...Field) {
	zap.L().Warn(msg, fields...)
}

// Debug logs a debug message.
func Debug(msg string, fields ...Field) {
	zap.L().Debug(msg, fields...)
}

// Fatal logs a fatal message.
func Fatal(msg string, fields ...Field) {
	zap.L().Fatal(msg, fields...)
}

// Panic logs a panic message.
func Panic(msg string, fields ...Field) {
	zap.L().Panic(msg, fields...)
}
