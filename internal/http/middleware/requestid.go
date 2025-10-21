// Ensure X-Request-Id.
package middleware // Request ID + correlation IDs

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID injects an X-Request-Id if absent and exposes to handlers.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-Id")
		if reqID == "" { reqID = uuid.NewString() }
		c.Writer.Header().Set("X-Request-Id", reqID)
		c.Set("req.id", reqID)
		c.Next()
	}
}