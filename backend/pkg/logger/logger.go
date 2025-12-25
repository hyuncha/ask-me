package logger

import (
	"go.uber.org/zap"
)

// Logger wraps zap logger
type Logger struct {
	*zap.SugaredLogger
}

// New creates a new logger instance based on environment
func New(environment string) *Logger {
	var zapLogger *zap.Logger
	var err error

	if environment == "production" {
		// Production logger: JSON format, info level
		zapLogger, err = zap.NewProduction()
	} else {
		// Development logger: console format, debug level
		zapLogger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	return &Logger{
		SugaredLogger: zapLogger.Sugar(),
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() {
	_ = l.SugaredLogger.Sync()
}
