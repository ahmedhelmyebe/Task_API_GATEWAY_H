package rate // In-memory token bucket limiter

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// memoryLimiter holds buckets per key.
type memoryLimiter struct {
	mu     sync.Mutex
	cfg    Cfg
	buckets map[string]*bucket
	log    *zap.Logger
}

// bucket stores token state.
type bucket struct {
	tokens     float64
	lastRefill time.Time
}

// NewMemoryLimiter constructs the limiter.
func NewMemoryLimiter(cfg Cfg, log *zap.Logger) Limiter {
	return &memoryLimiter{cfg: cfg, buckets: make(map[string]*bucket), log: NewLoggerTagged(log, "memory")}
}

// Allow checks/updates tokens.
func (m *memoryLimiter) Allow(key string) (bool, time.Duration) {
	m.mu.Lock(); defer m.mu.Unlock()
	b, ok := m.buckets[key]
	if !ok { // create new bucket
		b = &bucket{tokens: float64(m.cfg.Burst), lastRefill: time.Now()}
		m.buckets[key] = b
	}
	// Refill proportional to elapsed time
	elapsed := time.Since(b.lastRefill).Minutes()
	refill := elapsed * float64(m.cfg.RequestsPerMinute) // tokens per minute
	b.tokens = min(float64(m.cfg.Burst), b.tokens+refill)
	b.lastRefill = time.Now()
	if b.tokens >= 1 { // consume a token
		b.tokens -= 1
		return true, 0
	}
	// Not enough tokens -> compute retry-after until next token
	need := 1 - b.tokens
	perToken := time.Minute / time.Duration(max(1, m.cfg.RequestsPerMinute))
	return false, time.Duration(need*float64(perToken))
}

func min(a, b float64) float64 { if a < b { return a }; return b }
func max(a, b int) int { if a > b { return a }; return b }