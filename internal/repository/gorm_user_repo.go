// internal/repository/gorm_user_repo.go
package repository // GORM-backed adapter (sqlite/mysql/postgres)

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"example.com/api-gateway/internal/domain"
	"go.uber.org/zap"
)

// gormUser is the persistence model mapping to DB table.
// NOTE: Email is unique; consider immutable updates at handler/service layer.
type gormUser struct {
	ID           string    `gorm:"primaryKey;size:36"`
	Name         string
	Email        string    `gorm:"size:191;uniqueIndex"`
	PasswordHash string
	Role         string    `gorm:"size:32;index"`
	Active       bool      `gorm:"index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// toDomain converts persistence model to domain entity.
func (g gormUser) toDomain() domain.User {
	return domain.User{
		ID:           g.ID,
		Name:         g.Name,
		Email:        g.Email,
		PasswordHash: g.PasswordHash,
		Role:         g.Role,
		Active:       g.Active,
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
	}
}

// fromDomain converts domain entity to persistence model.
// Use for full reads/creates. For updates, prefer a map to avoid zero-value skipping.
func fromDomain(u *domain.User) gormUser {
	return gormUser{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		Active:       u.Active,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// gormRepo holds the DB connection and implements UserRepository.
type gormRepo struct {
	db *gorm.DB
}

// NewGormRepo opens a GORM connection for the given driver and DSN.
// It also auto-migrates the gormUser table schema.
func NewGormRepo(driver, dsn string, log *zap.Logger) (*gormRepo, error) {
	var dial gorm.Dialector
	switch driver {
	case "sqlite":
		dial = sqlite.Open(dsn)
	case "mysql":
		dial = mysql.Open(dsn)
	case "postgres":
		dial = postgres.Open(dsn)
	default:
		return nil, errors.New("unsupported gorm driver")
	}

	db, err := gorm.Open(dial, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&gormUser{}); err != nil {
		return nil, err
	}
	return &gormRepo{db: db}, nil
}

// Create inserts a new user and assigns defaults where necessary.
func (r *gormRepo) Create(u *domain.User) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	u.UpdatedAt = time.Now()

	gu := fromDomain(u)
	return r.db.Create(&gu).Error
}

// GetByID fetches a user by primary key.
func (r *gormRepo) GetByID(id string) (*domain.User, error) {
	var gu gormUser
	if err := r.db.First(&gu, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	d := gu.toDomain()
	return &d, nil
}

// GetByEmail fetches a user by unique email.
func (r *gormRepo) GetByEmail(email string) (*domain.User, error) {
	var gu gormUser
	if err := r.db.First(&gu, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	d := gu.toDomain()
	return &d, nil
}

// List returns users with simple paging.
func (r *gormRepo) List(offset, limit int) ([]domain.User, error) {
	var gus []gormUser
	if err := r.db.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&gus).Error; err != nil {
		return nil, err
	}
	out := make([]domain.User, 0, len(gus))
	for _, g := range gus {
		out = append(out, g.toDomain())
	}
	return out, nil
}

// Update persists changes from the provided domain.User.
// ⚠️ Uses a map to ensure "zero values" (e.g., Active=false) are not skipped by GORM.
// Email is treated as immutable here; if you choose to allow changing it, explicitly include it in the map.
func (r *gormRepo) Update(u *domain.User) error {
	u.UpdatedAt = time.Now()

	// Build a map of fields we intend to persist. This avoids the "zero-value is ignored" behavior
	// when passing a struct to Updates().
	update := map[string]any{
		"name":          u.Name,
		"password_hash": u.PasswordHash,
		"role":          u.Role,
		"active":        u.Active,
		"updated_at":    u.UpdatedAt,
		// "email":       u.Email, // ← keep commented to make email immutable by design
	}

	tx := r.db.Model(&gormUser{}).Where("id = ?", u.ID).Updates(update)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a user by id.
func (r *gormRepo) Delete(id string) error {
	res := r.db.Delete(&gormUser{ID: id})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
