package domain

import (
	"context"
	"time"
)

type User struct {
	ID        uint
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateUserInput struct {
	Name  string
	Email string
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uint) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
}

type UserUsecase interface {
	Create(ctx context.Context, input CreateUserInput) (*User, error)
	GetByID(ctx context.Context, id uint) (*User, error)
}
