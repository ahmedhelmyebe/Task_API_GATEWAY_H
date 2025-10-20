package logger // Zap logger initialization

import (
	"strings"      // Normalize level
	"go.uber.org/zap" // Zap logging
	"go.uber.org/zap/zapcore" // Levels and config
	"example.com/api-gateway/config" // Logging config type
)

// New builds a *zap.Logger from config.
func New(cfg config.Logging) *zap.Logger {
	lvl := zapcore.InfoLevel // default
	switch strings.ToLower(cfg.Level) { // parse textual level
	case "debug": lvl = zapcore.DebugLevel
	case "info": lvl = zapcore.InfoLevel
	case "warn": lvl = zapcore.WarnLevel
	case "error": lvl = zapcore.ErrorLevel
	}

	zc := zap.NewProductionConfig() // start from production baseline
	zc.Level = zap.NewAtomicLevelAt(lvl) // set level
	zc.Encoding = func() string { if cfg.JSON { return "json" }; return "console" }() // encoding
	if !cfg.Sampling { // disable sampler if requested
		zc.Sampling = nil
	}
	logger, _ := zc.Build() // build logger (ignore minor error)
	return logger
}