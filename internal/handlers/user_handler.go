// //CRUD; /users/me; contains your update path.

// package handlers // HTTP handlers for /users

// import (
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/go-playground/validator/v10"
// 	"example.com/api-gateway/internal/auth"
// 	"example.com/api-gateway/internal/domain"
// 	"example.com/api-gateway/internal/dto"
// 	"example.com/api-gateway/internal/service"
// )

// // UserHandler provides CRUD endpoints for user resources.
// // It binds/validates requests, delegates business logic to UserService,
// // and converts domain entities to DTOs for responses..
// type UserHandler struct {
// 	v *validator.Validate
// 	s *service.UserService
// }

// // NewUserHandler constructs a UserHandler instance with a fresh validator.
// // It wires the provided service into the handler.
// func NewUserHandler(s *service.UserService) *UserHandler { return &UserHandler{v: validator.New(), s: s} }

// // List GET /users (admin only)
// //Fetches a page of users and returns a safe DTO slice.
// func (h *UserHandler) List(c *gin.Context) {
// 	users, err := h.s.List(0, 100)
// 	if err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
// 	out := make([]dto.UserResponse, 0, len(users))
// 	for _, u := range users {
// 		out = append(out, dto.UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active})
// 	}
// 	c.JSON(200, out)
// }

// // Create handles POST /users (admin only).
// //  Step 1: Validate payload
// //  Step 2: Hash password
// //  Step 3: Construct domain.User
// //  Step 4: Persist via service
// //  Step 5: Return safe DTO
// func (h *UserHandler) Create(c *gin.Context) {
// 	var req dto.CreateUserRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		 c.JSON(400, gin.H{"error": "invalid json"})
// 		  return
// 		 }
// 	if err := h.v.Struct(req); err != nil { 
// 		c.JSON(422, gin.H{"error": err.Error()})
// 		 return
// 	 }

// 	hash, _ := auth.Hash(req.Password)
// 	u := &domain.User{ Name: req.Name, Email: req.Email,
// 		 PasswordHash: hash,
// 		  Role: req.Role,
// 		   Active: true, 
// 		   CreatedAt: time.Now(),
// 		    UpdatedAt: time.Now() }
// 	if err := h.s.Create(u); err != nil { 
// 		c.JSON(500, gin.H{"error": err.Error()})
// 		 return
// 	}
// 	c.JSON(201, dto.UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active})
// }

// // Get handles GET /users/:id (admin or self).
// //  Step 1: Fetch by id
// //  Step 2: Return safe DTO or 404
// func (h *UserHandler) Get(c *gin.Context) {
// 	id := c.Param("id")
// 	u, err := h.s.Get(id)
// 	if err != nil {
// 		 c.JSON(404, gin.H{"error": "not found"})
// 		 return
// 		 }
// 	c.JSON(200, dto.UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active})
// }

// // Patch handles PATCH /users/:id (admin or self).
// // IMPORTANT: Email is intentionally immutable here by design.
// //  Step 1: Bind & validate partial payload
// //  Step 2: Load current user
// //  Step 3: Apply only provided fields (name/password/role/active)
// //  Step 4: Persist via service
// // Step 5: Return safe DTO
// func (h *UserHandler) Patch(c *gin.Context) {
// 	//step1 : pull target id from URL 
// 	id := c.Param("id")
// 	//step 2: bind Json into a struct with pointer fields 
// 	var req dto.UpdateUserRequest
// 	if err := c.ShouldBindJSON(&req); err != nil { c.JSON(400, gin.H{"error": "invalid json"}); return }
	
// 	//step 3 : validate fields if present  
// 	if err := h.v.Struct(req); err != nil { c.JSON(422, gin.H{"error": err.Error()}); return }
	
// 	//step 4 : Load current entity so we can selectively path 
// 	u, err := h.s.Get(id); if err != nil { c.JSON(404, gin.H{"error": "not found"}); return }
	
	
// 	//step 5 : Apply only provided feild
// 	if req.Name != nil { u.Name = *req.Name }

// 	//  Password patch block:
// 	//    - hashes the incoming plaintext (if provided and non-empty)
// 	//    - ignores error from hashing
// 	if req.Password != nil { 
// 		// Short statement assigns local 'h' (shadows receiver 'h', but only inside this 'if').
// 		if h, _ := auth.Hash(*req.Password); *req.Password != "" { u.PasswordHash = h } }
	
// 	//
// 	if req.Role != nil { u.Role = *req.Role } //amin , user 
// 	if req.Active != nil { u.Active = *req.Active }//rue or false 
// 	//step 6 : persist changes via services -> repo 
// 	if err := h.s.Update(u); err != nil { c.JSON(500, gin.H{"error": err.Error()}); return }
// 	//step 7: Return DTO from currently patched 'u' in memory
// 	c.JSON(200, dto.UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active})
// }

// // Delete DELETE /users/:id (admin only)
// func (h *UserHandler) Delete(c *gin.Context) {
// 	id := c.Param("id")
// 	if err := h.s.Delete(id); err != nil { c.JSON(404, gin.H{"error": "not found"}); return }
// 	c.Status(204)
// }

// // Me GET /users/me (self)
// func (h *UserHandler) Me(c *gin.Context) {
// 	if sub, ok := c.Get("auth.sub"); ok {
// 		if u, err := h.s.Get(sub.(string)); err == nil {
// 			c.JSON(200, dto.UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active}); return
// 		}
// 	}
// 	c.JSON(404, gin.H{"error": "not found"})
// }

// // PatchMe PATCH /users/me (self)
// func (h *UserHandler) PatchMe(c *gin.Context) {
// 	sub, _ := c.Get("auth.sub")
// 	id := sub.(string)
// 	c.Params = append(c.Params, gin.Param{Key: "id", Value: id}) // reuse Patch logic by setting :id
// 	h.Patch(c)
// }



// internal/handlers/user_handler.go
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

// UserHandler provides CRUD endpoints for user resources.
// It binds/validates requests, delegates business logic to UserService,
// and converts domain entities to DTOs for responses.
type UserHandler struct {
	v *validator.Validate // per-handler validator
	s *service.UserService
}

// NewUserHandler constructs a UserHandler instance with a fresh validator.
// It wires the provided service into the handler.
func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{v: validator.New(), s: s}
}

// List handles GET /users (admin only).
// 🔹 Fetches a page of users and returns a safe DTO slice.
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.s.List(0, 100)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	out := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		out = append(out, dto.UserResponse{
			ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active,
		})
	}
	c.JSON(200, out)
}

// Create handles POST /users (admin only).
// 🔹 Step 1: Validate payload
// 🔹 Step 2: Hash password
// 🔹 Step 3: Construct domain.User
// 🔹 Step 4: Persist via service
// 🔹 Step 5: Return safe DTO
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}
	if err := h.v.Struct(req); err != nil {
		c.JSON(422, gin.H{"error": err.Error()})
		return
	}

	hash, err := auth.Hash(req.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to hash password"})
		return
	}

	u := &domain.User{
		Name:        req.Name,
		Email:       req.Email,
		PasswordHash: hash,
		Role:        req.Role,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.s.Create(u); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, dto.UserResponse{
		ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active,
	})
}

// Get handles GET /users/:id (admin or self).
// 🔹 Step 1: Fetch by id
// 🔹 Step 2: Return safe DTO or 404
func (h *UserHandler) Get(c *gin.Context) {
	id := c.Param("id")
	u, err := h.s.Get(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.JSON(200, dto.UserResponse{
		ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active,
	})
}

// Patch handles PATCH /users/:id (admin or self).
// IMPORTANT: Email is intentionally immutable here by design.
// 🔹 Step 1: Bind & validate partial payload
// 🔹 Step 2: Load current user
// 🔹 Step 3: Apply only provided fields (name/password/role/active)
// 🔹 Step 4: Persist via service
// 🔹 Step 5: Return safe DTO
func (h *UserHandler) Patch(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}
	if err := h.v.Struct(req); err != nil {
		c.JSON(422, gin.H{"error": err.Error()})
		return
	}

	u, err := h.s.Get(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}

	// 🔹 Apply present fields
	if req.Name != nil {
		u.Name = *req.Name
	}
	if req.Password != nil {
		// Only process if non-empty; this supports "change password" vs "leave as is".
		if *req.Password != "" {
			hashed, err := auth.Hash(*req.Password)
			if err != nil {
				c.JSON(500, gin.H{"error": "failed to hash password"})
				return
			}
			u.PasswordHash = hashed
		}
	}
	if req.Role != nil {
		u.Role = *req.Role
	}
	if req.Active != nil {
		u.Active = *req.Active // may be false; repo must not ignore zero values
	}

	if err := h.s.Update(u); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, dto.UserResponse{
		ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active,
	})
}

// Delete handles DELETE /users/:id (admin only).
func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.s.Delete(id); err != nil {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	c.Status(204)
}

// Me handles GET /users/me (self).
// 🔹 Uses auth.sub injected by the Authenticated middleware.
func (h *UserHandler) Me(c *gin.Context) {
	if sub, ok := c.Get("auth.sub"); ok {
		if u, err := h.s.Get(sub.(string)); err == nil {
			c.JSON(200, dto.UserResponse{
				ID: u.ID, Name: u.Name, Email: u.Email, Role: u.Role, Active: u.Active,
			})
			return
		}
	}
	c.JSON(404, gin.H{"error": "not found"})
}

// PatchMe handles PATCH /users/me (self).
// 🔹 Reuses Patch by injecting :id from auth.sub.
func (h *UserHandler) PatchMe(c *gin.Context) {
	sub, _ := c.Get("auth.sub")
	id := sub.(string)
	c.Params = append(c.Params, gin.Param{Key: "id", Value: id})
	h.Patch(c)
}
