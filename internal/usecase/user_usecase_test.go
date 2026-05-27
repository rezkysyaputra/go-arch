package usecase

import (
	"context"
	"testing"

	"go-arch/internal/domain"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	args := m.Called(ctx, id)
	user, _ := args.Get(0).(*domain.User)
	return user, args.Error(1)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	user, _ := args.Get(0).(*domain.User)
	return user, args.Error(1)
}

type mockEventPublisher struct {
	mock.Mock
}

func (m *mockEventPublisher) Publish(ctx context.Context, event domain.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestUserUsecaseCreate(t *testing.T) {
	t.Run("creates a user and publishes an event", func(t *testing.T) {
		ctx := context.Background()
		users := new(mockUserRepository)
		publisher := new(mockEventPublisher)

		users.On("FindByEmail", ctx, "ada@example.com").
			Return((*domain.User)(nil), domain.ErrUserNotFound).
			Once()
		users.On("Create", ctx, mock.MatchedBy(func(user *domain.User) bool {
			if user.Name == "Ada Lovelace" && user.Email == "ada@example.com" {
				user.ID = 42
			}

			return user.Name == "Ada Lovelace" && user.Email == "ada@example.com"
		})).
			Return(nil).
			Once()
		publisher.On("Publish", ctx, mock.MatchedBy(func(event domain.Event) bool {
			return event.Name == domain.UserCreatedEvent && len(event.Payload) > 0
		})).
			Return(nil).
			Once()

		uc := NewUserUsecase(users, publisher)

		user, err := uc.Create(ctx, domain.CreateUserInput{
			Name:  " Ada Lovelace ",
			Email: "ADA@EXAMPLE.COM",
		})

		require.NoError(t, err)
		require.Equal(t, uint(42), user.ID)
		require.Equal(t, "Ada Lovelace", user.Name)
		require.Equal(t, "ada@example.com", user.Email)
		users.AssertExpectations(t)
		publisher.AssertExpectations(t)
	})

	t.Run("rejects duplicate email", func(t *testing.T) {
		ctx := context.Background()
		users := new(mockUserRepository)
		publisher := new(mockEventPublisher)

		users.On("FindByEmail", ctx, "ada@example.com").
			Return(&domain.User{ID: 1, Email: "ada@example.com"}, nil).
			Once()

		uc := NewUserUsecase(users, publisher)

		_, err := uc.Create(ctx, domain.CreateUserInput{
			Name:  "Ada Lovelace",
			Email: "ada@example.com",
		})

		require.ErrorIs(t, err, domain.ErrEmailTaken)
		users.AssertExpectations(t)
		publisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	})
}
