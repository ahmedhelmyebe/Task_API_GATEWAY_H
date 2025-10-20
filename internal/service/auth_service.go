package service // Auth service combines repo + jwt + password

import (
	"errors"

	"example.com/api-gateway/config"
	"example.com/api-gateway/internal/auth"
	"example.com/api-gateway/internal/domain"
	"example.com/api-gateway/internal/repository"
	"go.uber.org/zap"
)

// AuthService validates credentials and issues JWTs.
type AuthService struct {
	repo repository.UserRepository // read user by email
	jwt  config.JWT                // signing config
	log  *zap.Logger               // logger
}

// NewAuthService wires dependencies.
func NewAuthService(r repository.UserRepository, jwtCfg config.JWT, l *zap.Logger) *AuthService {
	return &AuthService{repo: r, jwt: jwtCfg, log: l}
}

// Login checks credentials and returns token string.
func (s *AuthService) Login(email, password string) (string, *domain.User, error) {
	u, err := s.repo.GetByEmail(email)
	if err != nil { return "", nil, err }
	if !u.Active { return "", nil, errors.New("inactive user") }
	if !auth.Verify(u.PasswordHash, password) { return "", nil, errors.New("invalid credentials") }
	tok, err := auth.Sign(s.jwt, u.ID, u.Role)
	if err != nil { return "", nil, err }
	return tok, u, nil
}