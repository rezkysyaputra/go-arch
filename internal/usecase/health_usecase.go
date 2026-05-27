package usecase

import (
	"context"
	"time"

	"go-arch/internal/domain"
)

type HealthUsecase struct {
	dependencies []domain.HealthDependency
}

func NewHealthUsecase(dependencies []domain.HealthDependency) *HealthUsecase {
	return &HealthUsecase{dependencies: dependencies}
}

func (uc *HealthUsecase) Liveness() domain.HealthReport {
	return domain.HealthReport{
		Status:    domain.HealthStatusUp,
		CheckedAt: time.Now().UTC(),
	}
}

func (uc *HealthUsecase) Readiness(ctx context.Context) domain.HealthReport {
	report := domain.HealthReport{
		Status:    domain.HealthStatusUp,
		Checks:    make([]domain.HealthCheck, 0, len(uc.dependencies)),
		CheckedAt: time.Now().UTC(),
	}

	for _, dependency := range uc.dependencies {
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := dependency.Ping(checkCtx)
		cancel()

		check := domain.HealthCheck{
			Name:   dependency.Name(),
			Status: domain.HealthStatusUp,
		}
		if err != nil {
			check.Status = domain.HealthStatusDown
			check.Message = err.Error()
			report.Status = domain.HealthStatusDown
		}

		report.Checks = append(report.Checks, check)
	}

	return report
}
