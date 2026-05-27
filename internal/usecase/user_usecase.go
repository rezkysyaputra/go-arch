package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"go-arch/internal/domain"
)

type UserUsecase struct {
	users     domain.UserRepository
	publisher domain.EventPublisher
}

func NewUserUsecase(users domain.UserRepository, publisher domain.EventPublisher) *UserUsecase {
	return &UserUsecase{
		users:     users,
		publisher: publisher,
	}
}

func (uc *UserUsecase) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if input.Name == "" || input.Email == "" {
		return nil, domain.ErrInvalidInput
	}

	existing, err := uc.users.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrEmailTaken
	}

	user := &domain.User{
		Name:  input.Name,
		Email: input.Email,
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return nil, err
	}

	if err := uc.publishUserCreated(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *UserUsecase) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	if id == 0 {
		return nil, domain.ErrInvalidInput
	}

	return uc.users.FindByID(ctx, id)
}

func (uc *UserUsecase) publishUserCreated(ctx context.Context, user *domain.User) error {
	payload, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrPublishFailed, err)
	}

	if err := uc.publisher.Publish(ctx, domain.Event{
		Name:    domain.UserCreatedEvent,
		Payload: payload,
	}); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrPublishFailed, err)
	}

	return nil
}
