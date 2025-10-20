package dto // Request/response DTOs with validation tags

// LoginRequest carries credentials.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginResponse returns a JWT.
type LoginResponse struct {
	Token string `json:"token"`
}