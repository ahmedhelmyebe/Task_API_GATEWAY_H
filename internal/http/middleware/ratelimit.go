package middleware // Rate limit middleware wiring

import (
	"net"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"example.com/api-gateway/internal/rate"
)

// RateLimit creates a middleware using provided Limiter.
// Key strategy: if auth.sub exists -> per-user; else per-IP.
func RateLimit(limiter rate.Limiter, rpm int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil { c.Next(); return }
		key := clientIP(c)
		if sub, ok := c.Get("auth.sub"); ok { key = sub.(string) }
		allowed, retry := limiter.Allow(key)
		remaining := "unknown" // memory limiter doesn't track, but we expose standard headers
		c.Writer.Header().Set("X-RateLimit-Limit", strconv.Itoa(rpm))
		c.Writer.Header().Set("X-RateLimit-Remaining", remaining)
		if !allowed {
			c.Writer.Header().Set("Retry-After", strconv.Itoa(int(retry.Round(time.Second).Seconds())))
			c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded", "code": "too_many_requests"})
			return
		}
		c.Next()
	}
}

// clientIP extracts best-effort client IP.
func clientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" { ip = "0.0.0.0" }
	// Normalize IPv6-mapped IPv4
	if parsed := net.ParseIP(ip); parsed != nil {
		if v4 := parsed.To4(); v4 != nil { return v4.String() }
	}
	return ip
}