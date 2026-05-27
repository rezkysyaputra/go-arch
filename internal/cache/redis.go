package cache

import (
	"context"
	"fmt"

	"go-arch/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisDependency struct {
	client *redis.Client
}

func NewRedisClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping redis: %w; close redis: %v", err, closeErr)
		}

		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

func NewRedisHealthDependency(client *redis.Client) *RedisDependency {
	return &RedisDependency{client: client}
}

func (d *RedisDependency) Name() string {
	return "redis"
}

func (d *RedisDependency) Ping(ctx context.Context) error {
	return d.client.Ping(ctx).Err()
}
