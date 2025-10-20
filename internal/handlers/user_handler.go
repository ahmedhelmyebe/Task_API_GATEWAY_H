package handlers // HTTP handlers for /users

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"example.com/api-gateway/internal/auth"
	"example.com/api-gateway/internal/domain"
	"example.com/api-gateway/internal/dto"
	"example.com/api-gateway/internal/service"
)

// UserHandler provides CRUD endpoints.
type UserHandler struct {
	v *validator.Validate
	s *service.UserService
}

// NewUserHandler builds handler.
func NewUserHandler(s *service.UserService) *UserHandler { return &UserHandler{v: validator.New(), s: s} }

// List GET /users (admin only)
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.s.List(0, 100)
	if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
	out := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		out = append(out, dto.UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active})
	}
	c.JSON(200, out)
}

// Create POST /users (admin only)
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": "invalid json"}); return }
	if err := h.v.Struct(req); err != nil { c.JSON(422, gin.H{"error": err.Error()}); return }
	hash, _ := auth.Hash(req.Password)
	u := &domain.User{ Name: req.Name, Email: req.Email, PasswordHash: hash, Role: req.Role, Active: true, CreatedAt: time.Now(), UpdatedAt: time.Now() }
	if err := h.s.Create(u); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
	c.JSON(201, dto.UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active})
}

// Get GET /users/:id (admin or self)
func (h *UserHandler) Get(c *gin.Context) {
	id := c.Param("id")
	u, err := h.s.Get(id)
	if err != nil { c.JSON(404, gin.H{"error": "not found"}); return }
	c.JSON(200, dto.UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active})
}

// Patch PATCH /users/:id (admin or self)
func (h *UserHandler) Patch(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": "invalid json"}); return }
	if err := h.v.Struct(req); err != nil { c.JSON(422, gin.H{"error": err.Error()}); return }
	u, err := h.s.Get(id); if err != nil { c.JSON(404, gin.H{"error": "not found"}); return }
	if req.Name != nil { u.Name = *req.Name }
	if req.Password != nil { if h, _ := auth.Hash(*req.Password); *req.Password != "" { u.PasswordHash = h } }
	if req.Role != nil { u.Role = *req.Role }
	if req.Active != nil { u.Active = *req.Active }
	if err := h.s.Update(u); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
	c.JSON(200, dto.UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active})
}

// Delete DELETE /users/:id (admin only)
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.s.Delete(id); err != nil { c.JSON(404, gin.H{"error": "not found"}); return }
	c.Status(204)
}

// Me GET /users/me (self)
func (h *UserHandler) Me(c *gin.Context) {
	if sub, ok := c.Get("auth.sub"); ok {
		if u, err := h.s.Get(sub.(string)); err == nil {
			c.JSON(200, dto.UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active}); return
		}
	}
	c.JSON(404, gin.H{"error": "not found"})
}

// PatchMe PATCH /users/me (self)
func (h *UserHandler) PatchMe(c *gin.Context) {
	sub, _ := c.Get("auth.sub")
	id := sub.(string)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: id}) // reuse Patch logic by setting :id
	h.Patch(c)
}