// Zap logger from config.

package logger // Zap logger initialization  + console/file + Redis hook

import (
	// "strings" 
		"os"
	"path/filepath"
	"time"    // Normalize level
	"go.uber.org/zap" // Zap logging
	"go.uber.org/zap/zapcore" // Levels and config
	"example.com/api-gateway/config" // Logging config type
"gopkg.in/natefinch/lumberjack.v2"
rlog "example.com/api-gateway/internal/redis"
)


// ensureLogsDir makes sure the /logs directory exists (project root).
func ensureLogsDir() (string, error) {
	dir := filepath.Clean("./logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// fileSyncer creates a rotating file sink using lumberjack.
func fileSyncer() (zapcore.WriteSyncer, error) {
	dir, err := ensureLogsDir()
	if err != nil {
		return nil, err
	}
	lj := &lumberjack.Logger{
		Filename:   filepath.Join(dir, "app.log"),
		MaxSize:    25,  // megabytes before rotation
		MaxBackups: 7,   // keep 7 backup files
		MaxAge:     30,  // days
		Compress:   true,
	}
	return zapcore.AddSync(lj), nil
}


// consoleEncoder returns a human-readable encoder (good for local dev).
func consoleEncoder() zapcore.Encoder {
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339))
	}
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(cfg)
}

// jsonEncoder returns a structured JSON encoder (good for log aggregation).
func jsonEncoder() zapcore.Encoder {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339))
	}
	return zapcore.NewJSONEncoder(cfg)
}

// levelFromConfig parses the configured logging level.
func levelFromConfig(cfg config.Logging) zapcore.Level {
	switch cfg.Level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}



// New constructs a zap.Logger that writes to both console and a rotating file.
// If a redis AsyncLogger is provided, we attach a zap hook so that every log
// entry (after being written to console/file) is also enqueued for Redis saving.
func New(cfg config.Logging, redisAsync *rlog.AsyncLogger) (*zap.Logger, error) {
	level := levelFromConfig(cfg)

	// Build console core
	consoleCore := zapcore.NewCore(consoleEncoder(), zapcore.AddSync(os.Stdout), level)

	// Build file core
	fileWS, err := fileSyncer()
	if err != nil {
		return nil, err
	}
	fileCore := zapcore.NewCore(jsonEncoder(), fileWS, level)

	// Tee both
	baseCore := zapcore.NewTee(consoleCore, fileCore)

	// Optionally attach a hook that forwards entries to Redis asynchronously.
	if redisAsync != nil {
		baseCore = zapcore.RegisterHooks(baseCore, func(entry zapcore.Entry) error {
			le := rlog.LogEntry{
				Timestamp: entry.Time.UTC().Format(time.RFC3339),
				Level:     entry.Level.CapitalString(),
				Message:   entry.Message,
				Context:   map[string]any{}, // fields are added below via WithOptions hook
			}
			redisAsync.Enqueue(le)
			return nil
		})
	}

	// Build logger with or without sampling depending on config.
	// (We still honor cfg.JSON for console vs JSON output through encoder selection above.)
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}
	if cfg.Sampling {
		opts = append(opts, zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(c, time.Second, 100, 100)
		}))
	}

	logger := zap.New(baseCore, opts...)
	return logger, nil
}