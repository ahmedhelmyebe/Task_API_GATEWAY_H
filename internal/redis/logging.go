// internal/redis/logging.go
package redis // Structured log saving to Redis + async worker

import (
	"context"
	"encoding/json"
	"time"

	// rds "github.com/redis/go-redis/v9"
)

// Client is declared in client.go as a thin alias:
// type Client = redis.UniversalClient

// LogEntry is a structured log record that will be saved to Redis.
type LogEntry struct {
	// RFC3339 timestamp (UTC)
	Timestamp string         `json:"ts"`
	// Upper-case level: INFO | WARN | ERROR, etc.
	Level     string         `json:"level"`
	// Human-readable message
	Message   string         `json:"message"`
	// Arbitrary context key-values (route, userID, ip, requestID, etc.)
	Context   map[string]any `json:"context,omitempty"`
}

// keyDaily returns a per-date Redis key for archival lists.
// Example: logs:2025-10-21
func keyDaily(t time.Time) string {
	return "logs:" + t.UTC().Format("2006-01-02")
}

// keyRecent returns a rolling list key for fast recent fetch.
func keyRecent() string {
	return "logs:recent"
}

// SaveLog persists a log entry into Redis.
// It LPUSH-es the entry into both a recent list and also a per-day list
// and trims the recent list to a reasonable cap.
func SaveLog(ctx context.Context, c Client, entry LogEntry) error {
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	pipe := c.TxPipeline()
	pipe.LPush(ctx, keyRecent(), b)
	pipe.LTrim(ctx, keyRecent(), 0, 2000) // keep last ~2000 entries
	pipe.LPush(ctx, keyDaily(now), b)
	pipe.Expire(ctx, keyDaily(now), 7*24*time.Hour) // keep per-day logs ~7 days
	_, err = pipe.Exec(ctx)
	return err
}

// LoadRecentLogs loads up to N most recent logs from Redis (newest first).
func LoadRecentLogs(ctx context.Context, c Client, limit int) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	raws, err := c.LRange(ctx, keyRecent(), 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}
	out := make([]LogEntry, 0, len(raws))
	for _, s := range raws {
		var e LogEntry
		if err := json.Unmarshal([]byte(s), &e); err == nil {
			out = append(out, e)
		}
	}
	return out, nil
}

// AsyncLogger provides a non-blocking queue that flushes logs to Redis in background.
type AsyncLogger struct {
	c      Client
	ch     chan LogEntry
	closed chan struct{}
}

// NewAsyncLogger creates an async logger with a buffered queue.
func NewAsyncLogger(c Client, buffer int) *AsyncLogger {
	if buffer <= 0 {
		buffer = 1024
	}
	return &AsyncLogger{
		c:      c,
		ch:     make(chan LogEntry, buffer),
		closed: make(chan struct{}),
	}
}

// Start launches the background worker.
func (a *AsyncLogger) Start() {
	go func() {
		defer close(a.closed)
		ctx := context.Background()
		for e := range a.ch {
			// Best-effort with small timeout; we do not block request flow.
			ctxTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
			_ = SaveLog(ctxTimeout, a.c, e)
			cancel()
		}
	}()
}

// Stop closes the queue and waits for the worker to finish.
func (a *AsyncLogger) Stop() {
	close(a.ch)
	<-a.closed
}

// Enqueue adds a log entry to the background queue (non-blocking).
func (a *AsyncLogger) Enqueue(e LogEntry) {
	select {
	case a.ch <- e:
	default:
		// queue full, drop to remain non-blocking
	}
}
