package logging

import (
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// EchoLogger wraps zap.Logger to implement echo.Logger interface
type EchoLogger struct {
	logger *zap.Logger
	level  log.Lvl
}

// NewEchoLogger creates a new echo logger using zap
func NewEchoLogger() echo.Logger {
	return &EchoLogger{
		logger: DefaultLogger(),
		level:  log.INFO,
	}
}

// Output returns the output writer (not used with zap)
func (l *EchoLogger) Output() io.Writer {
	return io.Discard
}

// SetOutput sets the output writer (not used with zap)
func (l *EchoLogger) SetOutput(w io.Writer) {}

// Prefix returns the prefix (not used with zap)
func (l *EchoLogger) Prefix() string {
	return ""
}

// SetPrefix sets the prefix (not used with zap)
func (l *EchoLogger) SetPrefix(p string) {}

// Level returns the log level
func (l *EchoLogger) Level() log.Lvl {
	return l.level
}

// SetLevel sets the log level
func (l *EchoLogger) SetLevel(v log.Lvl) {
	l.level = v

	// Update zap logger level
	var zapLevel zapcore.Level
	switch v {
	case log.DEBUG:
		zapLevel = zapcore.DebugLevel
	case log.INFO:
		zapLevel = zapcore.InfoLevel
	case log.WARN:
		zapLevel = zapcore.WarnLevel
	case log.ERROR:
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	SetLevel(zapLevel)
}

// SetHeader sets the header (not used with zap)
func (l *EchoLogger) SetHeader(h string) {}

// Print logs a message at Print level
func (l *EchoLogger) Print(i ...interface{}) {
	l.logger.Sugar().Info(i...)
}

// Printf logs a formatted message at Print level
func (l *EchoLogger) Printf(format string, args ...interface{}) {
	l.logger.Sugar().Infof(format, args...)
}

// Printj logs a JSON object at Print level
func (l *EchoLogger) Printj(j log.JSON) {
	fields := make([]zap.Field, 0, len(j))
	for k, v := range j {
		fields = append(fields, zap.Any(k, v))
	}
	l.logger.Info("", fields...)
}

// Debug logs a message at Debug level
func (l *EchoLogger) Debug(i ...interface{}) {
	l.logger.Sugar().Debug(i...)
}

// Debugf logs a formatted message at Debug level
func (l *EchoLogger) Debugf(format string, args ...interface{}) {
	l.logger.Sugar().Debugf(format, args...)
}

// Debugj logs a JSON object at Debug level
func (l *EchoLogger) Debugj(j log.JSON) {
	fields := make([]zap.Field, 0, len(j))
	for k, v := range j {
		fields = append(fields, zap.Any(k, v))
	}
	l.logger.Debug("", fields...)
}

// Info logs a message at Info level
func (l *EchoLogger) Info(i ...interface{}) {
	l.logger.Sugar().Info(i...)
}

// Infof logs a formatted message at Info level
func (l *EchoLogger) Infof(format string, args ...interface{}) {
	l.logger.Sugar().Infof(format, args...)
}

// Infoj logs a JSON object at Info level
func (l *EchoLogger) Infoj(j log.JSON) {
	fields := make([]zap.Field, 0, len(j))
	for k, v := range j {
		fields = append(fields, zap.Any(k, v))
	}
	l.logger.Info("", fields...)
}

// Warn logs a message at Warn level
func (l *EchoLogger) Warn(i ...interface{}) {
	l.logger.Sugar().Warn(i...)
}

// Warnf logs a formatted message at Warn level
func (l *EchoLogger) Warnf(format string, args ...interface{}) {
	l.logger.Sugar().Warnf(format, args...)
}

// Warnj logs a JSON object at Warn level
func (l *EchoLogger) Warnj(j log.JSON) {
	fields := make([]zap.Field, 0, len(j))
	for k, v := range j {
		fields = append(fields, zap.Any(k, v))
	}
	l.logger.Warn("", fields...)
}

// Error logs a message at Error level
func (l *EchoLogger) Error(i ...interface{}) {
	l.logger.Sugar().Error(i...)
}

// Errorf logs a formatted message at Error level
func (l *EchoLogger) Errorf(format string, args ...interface{}) {
	l.logger.Sugar().Errorf(format, args...)
}

// Errorj logs a JSON object at Error level
func (l *EchoLogger) Errorj(j log.JSON) {
	fields := make([]zap.Field, 0, len(j))
	for k, v := range j {
		fields = append(fields, zap.Any(k, v))
	}
	l.logger.Error("", fields...)
}

// Fatal logs a message at Fatal level and exits
func (l *EchoLogger) Fatal(i ...interface{}) {
	l.logger.Sugar().Fatal(i...)
}

// Fatalf logs a formatted message at Fatal level and exits
func (l *EchoLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Sugar().Fatalf(format, args...)
}

// Fatalj logs a JSON object at Fatal level and exits
func (l *EchoLogger) Fatalj(j log.JSON) {
	fields := make([]zap.Field, 0, len(j))
	for k, v := range j {
		fields = append(fields, zap.Any(k, v))
	}
	l.logger.Fatal("", fields...)
}

// Panic logs a message at Panic level and panics
func (l *EchoLogger) Panic(i ...interface{}) {
	l.logger.Sugar().Panic(i...)
}

// Panicf logs a formatted message at Panic level and panics
func (l *EchoLogger) Panicf(format string, args ...interface{}) {
	l.logger.Sugar().Panicf(format, args...)
}

// Panicj logs a JSON object at Panic level and panics
func (l *EchoLogger) Panicj(j log.JSON) {
	fields := make([]zap.Field, 0, len(j))
	for k, v := range j {
		fields = append(fields, zap.Any(k, v))
	}
	l.logger.Panic("", fields...)
}
