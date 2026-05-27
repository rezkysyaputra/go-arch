package domain

import (
	"context"
	"time"
)

const (
	HealthStatusUp   = "up"
	HealthStatusDown = "down"
)

type HealthDependency interface {
	Name() string
	Ping(ctx context.Context) error
}

type HealthUsecase interface {
	Liveness() HealthReport
	Readiness(ctx context.Context) HealthReport
}

type HealthReport struct {
	Status    string
	Checks    []HealthCheck
	CheckedAt time.Time
}

type HealthCheck struct {
	Name    string
	Status  string
	Message string
}
