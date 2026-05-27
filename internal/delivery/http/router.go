package http

import (
	"go-arch/docs"
	"go-arch/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(userUsecase domain.UserUsecase, healthUsecase domain.HealthUsecase, logger zerolog.Logger) *gin.Engine {
	router := gin.New()
	router.Use(RequestIDMiddleware(), ZerologMiddleware(logger), gin.Recovery())

	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	healthHandler := NewHealthHandler(healthUsecase)
	router.GET("/health/live", healthHandler.Liveness)
	router.GET("/health/ready", healthHandler.Readiness)
	router.GET("/health", healthHandler.Liveness)

	userHandler := NewUserHandler(userUsecase)

	users := router.Group("/users")
	{
		users.POST("", userHandler.Create)
		users.GET("/:id", userHandler.GetByID)
	}

	return router
}
