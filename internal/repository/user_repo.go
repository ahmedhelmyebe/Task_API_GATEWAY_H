//Repository interface + factory + ErrNotFound.

package repository // Abstraction + constructors for adapters

import (
	"errors"
	"fmt"

	"example.com/api-gateway/config"
	"example.com/api-gateway/internal/domain"
	"go.uber.org/zap"
)

// UserRepository abstracts CRUD regardless of DB.
type UserRepository interface {
	Create(u *domain.User) error
	GetByID(id string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	List(offset, limit int) ([]domain.User, error)
	Update(u *domain.User) error
	Delete(id string) error
}

// NewUserRepository selects concrete adapter by cfg.Database.Driver.
func NewUserRepository(dbCfg config.Database, log *zap.Logger) (UserRepository, error) {
	switch dbCfg.Driver {
	case "sqlite":
		return NewGormRepo("sqlite", dbCfg.DSN, log)
	case "mysql":
		return NewGormRepo("mysql", dbCfg.DSN, log)
	case "postgres":
		return NewGormRepo("postgres", dbCfg.DSN, log)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", dbCfg.Driver)
	}
}

// ErrNotFound is returned when entity is missing.
var ErrNotFound = errors.New("not found")