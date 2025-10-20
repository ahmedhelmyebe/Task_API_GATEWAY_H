package test


import (
"testing"
"time"
"example.com/api-gateway/config"
"example.com/api-gateway/internal/rate"
"go.uber.org/zap"
)


func TestMemoryLimiter(t *testing.T) {
cfg := config.RateLimit{
    Enabled: true,
    Strategy: "memory",
    RequestsPerMinute: 2,
    Burst: 2, // was 1 — allow two immediate tokens
}

l := rate.NewMemoryLimiter(cfg, zap.NewNop())
// 1st: allowed
if ok, _ := l.Allow("k"); !ok { t.Fatal("want allow #1") }
// 2nd: burst -> allowed
if ok, _ := l.Allow("k"); !ok { t.Fatal("want allow #2") }
// 3rd: should be denied
if ok, _ := l.Allow("k"); ok { t.Fatal("want deny #3") }
time.Sleep(time.Second * 31) // refill half a minute -> at least one token
if ok, _ := l.Allow("k"); !ok { t.Fatal("want allow after refill") }
}