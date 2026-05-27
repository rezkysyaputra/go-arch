package http

import (
	nethttp "net/http"
	"time"

	"go-arch/internal/domain"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	health domain.HealthUsecase
}

func NewHealthHandler(health domain.HealthUsecase) *HealthHandler {
	return &HealthHandler{health: health}
}

// Liveness godoc
// @Summary Liveness probe
// @Description Checks whether the API process is alive.
// @Tags health
// @Produce json
// @Success 200 {object} apiResponse{data=healthResponse}
// @Router /health/live [get]
func (h *HealthHandler) Liveness(ctx *gin.Context) {
	respondSuccess(ctx, nethttp.StatusOK, "service is alive", newHealthResponse(h.health.Liveness()))
}

// Readiness godoc
// @Summary Readiness probe
// @Description Checks PostgreSQL, RabbitMQ, and Redis dependencies.
// @Tags health
// @Produce json
// @Success 200 {object} apiResponse{data=healthResponse}
// @Failure 503 {object} apiResponse{data=healthResponse}
// @Router /health/ready [get]
func (h *HealthHandler) Readiness(ctx *gin.Context) {
	report := h.health.Readiness(ctx.Request.Context())
	status := nethttp.StatusOK
	if report.Status != domain.HealthStatusUp {
		status = nethttp.StatusServiceUnavailable
	}

	if status != nethttp.StatusOK {
		respondFailureWithData(ctx, status, "service is not ready", newHealthResponse(report), nil)
		return
	}

	respondSuccess(ctx, status, "service is ready", newHealthResponse(report))
}

type healthResponse struct {
	Status    string                `json:"status"`
	Checks    []healthCheckResponse `json:"checks,omitempty"`
	CheckedAt string                `json:"checked_at"`
}

type healthCheckResponse struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func newHealthResponse(report domain.HealthReport) healthResponse {
	checks := make([]healthCheckResponse, 0, len(report.Checks))
	for _, check := range report.Checks {
		checks = append(checks, healthCheckResponse{
			Name:    check.Name,
			Status:  check.Status,
			Message: check.Message,
		})
	}

	return healthResponse{
		Status:    report.Status,
		Checks:    checks,
		CheckedAt: report.CheckedAt.Format(time.RFC3339),
	}
}
