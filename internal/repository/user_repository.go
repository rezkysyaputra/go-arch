package repository

import (
	"context"
	"errors"
	"time"

	"go-arch/internal/domain"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type userModel struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:255;not null"`
	Email     string    `gorm:"size:255;uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (userModel) TableName() string {
	return "users"
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, user *domain.User) error {
	model := userModelFromDomain(user)
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		if isUniqueViolation(err) {
			return domain.ErrDuplicateEmail
		}

		return err
	}

	*user = *model.toDomain()
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *GormUserRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var user userModel
	err := r.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user.toDomain(), nil
}

func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user userModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return user.toDomain(), nil
}

func userModelFromDomain(user *domain.User) userModel {
	return userModel{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (m userModel) toDomain() *domain.User {
	return &domain.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
