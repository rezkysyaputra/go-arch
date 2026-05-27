package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"go-arch/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog"
)

func main() {
	baseLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	if len(os.Args) < 2 {
		baseLogger.Error().Str("usage", "go run ./cmd/migrate up|down").Msg("migration command is required")
		os.Exit(1)
	}

	command := os.Args[1]
	if command == "--" && len(os.Args) > 2 {
		command = os.Args[2]
	}

	cfg, err := config.Load()
	if err != nil {
		baseLogger.Error().Err(err).Msg("configuration load failed")
		os.Exit(1)
	}
	logger := newLogger(cfg)

	m, err := migrate.New("file://migrations", cfg.Database.MigrationURL())
	if err != nil {
		logger.Error().Err(err).Msg("migration init failed")
		os.Exit(1)
	}
	defer m.Close()

	switch command {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	default:
		err = fmt.Errorf("unsupported migration command %q", command)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		logger.Info().Msg("database schema already up to date")
		return
	}
	if err != nil {
		logger.Error().Err(err).Msg("migration failed")
		os.Exit(1)
	}

	logger.Info().Str("command", command).Msg("migration completed")
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
