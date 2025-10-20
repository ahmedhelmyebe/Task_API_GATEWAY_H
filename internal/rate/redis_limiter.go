package rate // Redis-backed fixed-window-ish token bucket

import (
	"context"
	"time"

	"example.com/api-gateway/internal/redis"
	"go.uber.org/zap"
)

// redisLimiter implements Limiter using Redis INCR + TTL.
type redisLimiter struct {
	c   redis.Client
	cfg Cfg
	log *zap.Logger
}

// NewRedisLimiter constructs the limiter.
func NewRedisLimiter(c redis.Client, cfg Cfg, log *zap.Logger) Limiter {
	return &redisLimiter{c: c, cfg: cfg, log: NewLoggerTagged(log, "redis")}
}

// Allow uses a per-minute key with burst windowing.
func (r *redisLimiter) Allow(key string) (bool, time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	bucketKey := "rl:" + key + ":" + time.Now().Format("200601021504") // YYYYMMDDHHmm
	count, err := r.c.Incr(ctx, bucketKey).Result()
	if err != nil {
		r.log.Warn("redis incr fail", zap.Error(err))
		return true, 0 // fail-open
	}
	if count == 1 {
		_ = r.c.Expire(ctx, bucketKey, time.Minute).Err()
	}
	if int(count) <= r.cfg.RequestsPerMinute+r.cfg.Burst {
		return true, 0
	}
	// compute seconds left in current minute
	ttl, _ := r.c.TTL(ctx, bucketKey).Result()
	if ttl < 0 {
		ttl = 10 * time.Second
	}
	return false, ttl
}
