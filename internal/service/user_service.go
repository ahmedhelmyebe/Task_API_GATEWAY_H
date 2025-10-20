package service // Business logic around users

import (
	"errors"

	"example.com/api-gateway/internal/domain"
	"example.com/api-gateway/internal/repository"
	"go.uber.org/zap"
)

// UserService coordinates repo operations and invariants.
type UserService struct {
	repo repository.UserRepository
	log  *zap.Logger
}

// NewUserService constructs the service.
func NewUserService(r repository.UserRepository, l *zap.Logger) *UserService {
	return &UserService{repo: r, log: l}
}

// Create creates a user, ensuring unique email is enforced at DB.
func (s *UserService) Create(u *domain.User) error { return s.repo.Create(u) }

// Get returns user by id.
func (s *UserService) Get(id string) (*domain.User, error) { return s.repo.GetByID(id) }

// GetByEmail returns user by email.
func (s *UserService) GetByEmail(email string) (*domain.User, error) { return s.repo.GetByEmail(email) }

// List pages users.
func (s *UserService) List(offset, limit int) ([]domain.User, error) { return s.repo.List(offset, limit) }

// Update updates fields.
func (s *UserService) Update(u *domain.User) error { return s.repo.Update(u) }

// Delete removes a user.
func (s *UserService) Delete(id string) error { return s.repo.Delete(id) }

// CanSelf checks if requester id matches target id.
func (s *UserService) CanSelf(requesterID, targetID string) bool { return requesterID == targetID }

// IsAdmin checks admin role.
func (s *UserService) IsAdmin(role string) bool { return role == "admin" }

// GuardSelfOrAdmin returns error if neither self nor admin.
func (s *UserService) GuardSelfOrAdmin(requesterID, requesterRole, targetID string) error {
	if s.IsAdmin(requesterRole) || s.CanSelf(requesterID, targetID) { return nil }
	return errors.New("forbidden")
}