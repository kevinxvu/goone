package logging

import (
	"context"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type contextKey string

const loggerKey = contextKey("logger")

var (
	defaultLogger     *zap.Logger
	defaultLoggerOnce sync.Once
)

var conf = &Config{
	Level:      zapcore.InfoLevel,
	TimeFormat: "2006-01-02 15:04:05",
}

// Config holds logging configuration
type Config struct {
	Level      zapcore.Level
	FilePath   string
	TimeFormat string // e.g., "2006-01-02 15:04:05" or "02/01/2006 03:04 PM"
}

// SetConfig updates the global logging configuration
func SetConfig(c *Config) {
	conf = &Config{
		Level:      c.Level,
		FilePath:   c.FilePath,
		TimeFormat: c.TimeFormat,
	}
}

// SetLevel updates just the log level
func SetLevel(l zapcore.Level) {
	conf.Level = l
}

// Type returns a logger with a type field for identification
func Type(name string) *zap.Logger {
	return DefaultLogger().With(zap.String("type", name))
}

// NewLogger creates a new logger with dual output (console + file)
func NewLogger(cfg *Config) *zap.Logger {
	var cores []zapcore.Core

	// Console output (JSON format)
	consoleEncoderConfig := zap.NewProductionEncoderConfig()
	consoleEncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(cfg.TimeFormat)
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	consoleCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(consoleEncoderConfig),
		zapcore.Lock(os.Stdout),
		zap.NewAtomicLevelAt(cfg.Level),
	)
	cores = append(cores, consoleCore)

	// File output (structured JSON with rotation)
	if cfg.FilePath != "" {
		lumberjackLogger := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     15, // days
			Compress:   true,
		}
		fileEncoderConfig := zap.NewProductionEncoderConfig()
		fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(fileEncoderConfig),
			zapcore.AddSync(lumberjackLogger),
			zap.NewAtomicLevelAt(cfg.Level),
		)
		cores = append(cores, fileCore)
	}

	return zap.New(
		zapcore.NewTee(cores...),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}

// DefaultLogger returns the singleton logger instance
func DefaultLogger() *zap.Logger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = NewLogger(conf)
	})
	return defaultLogger
}

// WithContext stores a logger in the context
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves a logger from context, or returns default
func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return DefaultLogger()
	}
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	return DefaultLogger()
}

// ErrField is a helper to create an error field for logging
func ErrField(err error) zap.Field {
	return zap.Error(err)
}

// TypeField is a helper to create a type field for logging
func TypeField(value string) zap.Field {
	return zap.String("type", value)
}
