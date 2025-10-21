// internal/http/middleware/rbac.go
package middleware // RBAC helpers for route protection

import "github.com/gin-gonic/gin"

// RequireAdmin returns middleware that allows only users with role "admin".
// It assumes the Authenticated middleware has already set "auth.role".
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, _ := c.Get("auth.role"); role == "admin" {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(403, gin.H{"error": "admin only", "code": "forbidden"})
	}
}

// RequireSelfOrAdmin returns middleware that allows access if:
// 1) role is admin, or
// 2) the authenticated subject ("auth.sub") matches the :id path param.
// Otherwise it returns 403.
func RequireSelfOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("auth.role")
		if role == "admin" {
			c.Next()
			return
		}
		sub, _ := c.Get("auth.sub")
		if sub == c.Param("id") {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden: not owner", "code": "forbidden"})
	}
}
