package repository

import (
	"database/sql"
	"fmt"

	"go-arch/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDB(cfg config.DatabaseConfig) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("open postgres connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, sqlDB, nil
}
