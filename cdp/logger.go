package cdp

import (
	"context"
	"log/slog"
	"os"
)

// Logger is a wrapper around slog.Logger to provide consistent logging.
type Logger struct {
	logger *slog.Logger
	debug  bool
}

// NewLogger creates a new Logger instance.
func NewLogger(customLogger *slog.Logger, debug bool) *Logger {
	if customLogger != nil {
		return &Logger{logger: customLogger, debug: debug}
	}

	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	return &Logger{logger: slog.New(handler), debug: debug}
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, args ...any) {
	if l.debug {
		l.logger.Debug(msg, args...)
	}
}

// Info logs an info message.
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// LogError logs an error with context and returns it. 
// Used for the FailOnException pattern.
func (l *Logger) LogError(ctx context.Context, err error, msg string) {
	l.logger.ErrorContext(ctx, msg, "error", err)
}
