package domain // Core domain entity definitions

import "time" // Timestamps

// User is the aggregate root for identity and RBAC.
type User struct {
	ID           string    // UUID string
	Name         string    // Display name
	Email        string    // Unique email
	PasswordHash string    // Bcrypt hash
	Role         string    // admin|user
	Active       bool      // Soft-active flag
	CreatedAt    time.Time // Audit
	UpdatedAt    time.Time // Audit
}