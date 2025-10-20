package middleware // RBAC helpers for route protection

import "github.com/gin-gonic/gin"

// RequireAdmin allows only admin role.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if role, _ := c.Get("auth.role"); role == "admin" {
			c.Next(); return
		}
		c.AbortWithStatusJSON(403, gin.H{"error": "admin only", "code": "forbidden"})
	}
}

// RequireSelfOrAdmin checks path param :id against auth.sub
func RequireSelfOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("auth.role")
		if role == "admin" { c.Next(); return }
		sub, _ := c.Get("auth.sub")
		if sub == c.Param("id") { c.Next(); return }
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden: not owner", "code": "forbidden"})
	}
}