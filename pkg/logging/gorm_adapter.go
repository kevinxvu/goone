package logging

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger wraps zap.Logger to implement gorm's logger interface
type GormLogger struct {
	logLevel      gormlogger.LogLevel
	slowThreshold time.Duration
}

// NewGormLogger creates a new GORM logger using zap
func NewGormLogger() *GormLogger {
	return &GormLogger{
		logLevel:      gormlogger.Info,
		slowThreshold: 200 * time.Millisecond,
	}
}

// LogMode sets the log level
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info logs info messages
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Info {
		FromContext(ctx).Sugar().With(TypeField("sql")).Infof(msg, data...)
	}
}

// Warn logs warn messages
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Warn {
		FromContext(ctx).Sugar().With(TypeField("sql")).Warnf(msg, data...)
	}
}

// Error logs error messages
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Error {
		FromContext(ctx).Sugar().With(TypeField("sql")).Errorf(msg, data...)
	}
}

// Trace logs SQL queries
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	logger := FromContext(ctx)
	fields := []zap.Field{
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
		zap.String("query", sql),
		TypeField("sql"),
	}

	switch {
	case err != nil && l.logLevel >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		fields = append(fields, zap.Error(err))
		logger.Error("gorm error", fields...)
	case elapsed > l.slowThreshold && l.slowThreshold != 0 && l.logLevel >= gormlogger.Warn:
		logger.Warn(fmt.Sprintf("slow sql >= %v", l.slowThreshold), fields...)
	case l.logLevel >= gormlogger.Info:
		logger.Debug("gorm trace", fields...)
	}
}
