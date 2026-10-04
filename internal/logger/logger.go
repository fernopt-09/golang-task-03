package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a zap logger with the given level (debug, info, warn, error).
// If the level is invalid, info is used.
func New(levelStr string) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(levelStr)
	if err != nil {
		level = zapcore.InfoLevel
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	return cfg.Build()
}
