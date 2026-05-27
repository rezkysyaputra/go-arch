package usecase

import (
	"context"
	"errors"
	"testing"

	"go-arch/internal/domain"

	"github.com/stretchr/testify/require"
)

type mockHealthDependency struct {
	name string
	err  error
}

func (m mockHealthDependency) Name() string {
	return m.name
}

func (m mockHealthDependency) Ping(ctx context.Context) error {
	return m.err
}

func TestHealthUsecaseReadiness(t *testing.T) {
	t.Run("returns up when every dependency is healthy", func(t *testing.T) {
		uc := NewHealthUsecase([]domain.HealthDependency{
			mockHealthDependency{name: "postgres"},
			mockHealthDependency{name: "rabbitmq"},
			mockHealthDependency{name: "redis"},
		})

		report := uc.Readiness(context.Background())

		require.Equal(t, domain.HealthStatusUp, report.Status)
		require.Len(t, report.Checks, 3)
	})

	t.Run("returns down when a dependency fails", func(t *testing.T) {
		uc := NewHealthUsecase([]domain.HealthDependency{
			mockHealthDependency{name: "postgres"},
			mockHealthDependency{name: "redis", err: errors.New("connection refused")},
		})

		report := uc.Readiness(context.Background())

		require.Equal(t, domain.HealthStatusDown, report.Status)
		require.Equal(t, domain.HealthStatusDown, report.Checks[1].Status)
		require.Equal(t, "connection refused", report.Checks[1].Message)
	})
}
