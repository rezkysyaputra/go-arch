package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-arch/internal/cache"
	"go-arch/internal/config"
	amqpdelivery "go-arch/internal/delivery/amqp"
	httpdelivery "go-arch/internal/delivery/http"
	"go-arch/internal/domain"
	"go-arch/internal/repository"
	"go-arch/internal/usecase"

	"github.com/rs/zerolog"
)

// @title Go Clean Architecture API
// @version 1.0
// @description Scalable Golang backend boilerplate using Gin, PostgreSQL, RabbitMQ, Redis, and Clean Architecture.
// @BasePath /
func main() {
	baseLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	cfg, err := config.Load()
	if err != nil {
		baseLogger.Error().Err(err).Msg("configuration load failed")
		os.Exit(1)
	}

	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		if err := runHealthCheck(cfg); err != nil {
			baseLogger.Error().Err(err).Msg("healthcheck failed")
			os.Exit(1)
		}

		return
	}

	logger := newLogger(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, sqlDB, err := repository.NewPostgresDB(cfg.Database)
	if err != nil {
		logger.Error().Err(err).Msg("database connection failed")
		os.Exit(1)
	}
	defer sqlDB.Close()

	redisClient, err := cache.NewRedisClient(ctx, cfg.Redis)
	if err != nil {
		logger.Error().Err(err).Msg("redis connection failed")
		os.Exit(1)
	}
	defer redisClient.Close()

	rabbitConn, err := amqpdelivery.NewConnection(cfg.RabbitMQ.URL)
	if err != nil {
		logger.Error().Err(err).Msg("rabbitmq connection failed")
		os.Exit(1)
	}
	defer rabbitConn.Close()

	if err := amqpdelivery.DeclareTopology(ctx, rabbitConn, cfg.RabbitMQ); err != nil {
		logger.Error().Err(err).Msg("rabbitmq topology declaration failed")
		os.Exit(1)
	}

	// Manual dependency injection is intentionally explicit at the composition root.
	userRepository := repository.NewGormUserRepository(db)
	eventPublisher := amqpdelivery.NewPublisher(rabbitConn, cfg.RabbitMQ)
	userUsecase := usecase.NewUserUsecase(userRepository, eventPublisher)
	healthUsecase := usecase.NewHealthUsecase([]domain.HealthDependency{
		repository.NewPostgresHealthDependency(sqlDB),
		rabbitConn,
		cache.NewRedisHealthDependency(redisClient),
	})
	router := httpdelivery.NewRouter(userUsecase, healthUsecase, logger)

	consumer := amqpdelivery.NewConsumerService(rabbitConn, cfg.RabbitMQ, &logger)
	if err := consumer.Start(ctx); err != nil {
		logger.Error().Err(err).Msg("rabbitmq consumer failed to start")
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info().Str("addr", server.Addr).Msg("http server started")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error().Err(err).Msg("http server failed")
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("http server shutdown failed")
		os.Exit(1)
	}

	logger.Info().Msg("application stopped")
}

func newLogger(cfg config.Config) zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	var output io.Writer = os.Stdout
	if cfg.Log.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	return zerolog.New(output).
		With().
		Timestamp().
		Str("env", cfg.App.Env).
		Logger()
}

func runHealthCheck(cfg config.Config) error {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%s/health/ready", cfg.HTTP.Port))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck returned status %d", resp.StatusCode)
	}

	return nil
}
