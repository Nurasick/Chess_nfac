package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("cache.NewRedisClient: %w", err)
	}

	client := redis.NewClient(opts)

	// Verify connection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache.NewRedisClient: ping failed: %w", err)
	}

	return client, nil
}
