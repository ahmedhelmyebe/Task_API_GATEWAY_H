// internal/dto/user_dto.go
package dto // User DTOs

// CreateUserRequest is the admin-only payload for creating users.
// Email is required and unique; password must meet minimum policy.
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=admin user"`
}

// UpdateUserRequest is the partial payload for PATCH /users/:id.
// NOTE: Email is intentionally omitted to keep it immutable by design.
// If you decide to support email changes, add it here with pointer semantics
// and update the repository Update() to include it explicitly.
type UpdateUserRequest struct {
	Name     *string `json:"name" validate:"omitempty,min=2"`
	Password *string `json:"password" validate:"omitempty,min=6"`
	Role     *string `json:"role" validate:"omitempty,oneof=admin user"`
	Active   *bool   `json:"active"`
}

// UserResponse is the safe representation returned to clients (no password hash).
type UserResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Active bool   `json:"active"`
}
