// internal/http/handlers/logs_handler.go
package handlers // Admin logs endpoint

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	rlog "example.com/api-gateway/internal/redis"
)

// LogsHandler exposes admin-only endpoints to inspect logs saved in Redis.
type LogsHandler struct {
	RClient rlog.Client // Redis client
}

// NewLogsHandler builds a new LogsHandler.
func NewLogsHandler(rc rlog.Client) *LogsHandler {
	return &LogsHandler{RClient: rc}
}

// ListRecent handles GET /api/logs?limit=N (admin only).
// It loads up to N recent logs from Redis and returns JSON.
func (h *LogsHandler) ListRecent(c *gin.Context) {
	limit := 100
	if q := c.Query("limit"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 {
			limit = v
		}
	}
	logs, err := rlog.LoadRecentLogs(c.Request.Context(), h.RClient, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"count": len(logs),
		"items": logs,
	})
}
