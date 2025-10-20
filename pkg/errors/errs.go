package errors // Consistent error envelope helpers

import "github.com/gin-gonic/gin"

// JSON writes a standardized error object.
func JSON(c *gin.Context, status int, code, msg string) {
	c.JSON(status, gin.H{"error": msg, "code": code, "traceId": c.GetString("req.id")})
}