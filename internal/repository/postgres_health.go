package repository

import (
	"context"
	"database/sql"
)

type PostgresHealthDependency struct {
	db *sql.DB
}

func NewPostgresHealthDependency(db *sql.DB) *PostgresHealthDependency {
	return &PostgresHealthDependency{db: db}
}

func (d *PostgresHealthDependency) Name() string {
	return "postgres"
}

func (d *PostgresHealthDependency) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}
