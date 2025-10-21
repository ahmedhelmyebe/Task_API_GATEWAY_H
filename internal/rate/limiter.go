//internal/rate/: Limiter interface + memory and Redis impl.

package rate // Limiter interface and config shim

import (
	"time"
	"example.com/api-gateway/config"
	"go.uber.org/zap"
)

// Limiter decides if a key may proceed right now.
type Limiter interface {
	Allow(key string) (allowed bool, retryAfter time.Duration)
}

// Cfg is an alias to avoid importing config everywhere.
type Cfg = config.RateLimit

// Noop implements Limiter that always allows (unused, but handy for tests)
type Noop struct{}

func (Noop) Allow(string) (bool, time.Duration) { return true, 0 }

// Common constructor helpers can accept zap logger for diagnostics.
func NewLoggerTagged(l *zap.Logger, name string) *zap.Logger { return l.With(zap.String("limiter", name)) }