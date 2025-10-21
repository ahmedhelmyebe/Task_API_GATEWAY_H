//returns JWT.

package handlers // HTTP handlers for /auth

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"example.com/api-gateway/internal/dto"
	"example.com/api-gateway/internal/service"
)

// AuthHandler exposes login endpoint.
type AuthHandler struct {
	v *validator.Validate
	s *service.AuthService
}

// NewAuthHandler builds the handler.
func NewAuthHandler(s *service.AuthService) *AuthHandler { 
	return &AuthHandler{
		v: validator.New(),
		 s: s}
 }

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"}); return
	}
	if err := h.v.Struct(req); err != nil {
		c.JSON(422, gin.H{"error": err.Error()}); return
	}
	token, u, err := h.s.Login(req.Email, req.Password)
	if err != nil { c.JSON(401, gin.H{"error": err.Error()}); return }
	c.JSON(200, dto.LoginResponse{Token: token})
	_ = u // could return user profile too if desired
}