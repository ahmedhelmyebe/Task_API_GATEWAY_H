package middleware // JWT authentication: parse + attach identity

import (
	"strings"

	"github.com/gin-gonic/gin"
	"example.com/api-gateway/config"
	"example.com/api-gateway/internal/auth"
)

// Authenticated ensures valid JWT and stores sub/role in context.
func Authenticated(jwtCfg config.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "authentication required", "code": "unauthorized"})
			return
		}
		token := strings.TrimPrefix(h, "Bearer ")
		claims, err := auth.Parse(jwtCfg, token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token", "code": "unauthorized"})
			return
		}
		c.Set("auth.sub", claims.Sub)
		c.Set("auth.role", claims.Role)
		c.Next()
	}
}