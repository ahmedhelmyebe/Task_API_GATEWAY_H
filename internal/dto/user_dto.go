package dto // User DTOs

// CreateUserRequest for admin create.
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=admin user"`
}

// UpdateUserRequest for patching fields.
type UpdateUserRequest struct {
	Name     *string `json:"name" validate:"omitempty,min=2"`
	Password *string `json:"password" validate:"omitempty,min=6"`
	Role     *string `json:"role" validate:"omitempty,oneof=admin user"`
	Active   *bool   `json:"active"`
}

// UserResponse returned to clients (no password).
type UserResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Active bool   `json:"active"`
}