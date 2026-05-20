package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewSessionCache(client *redis.Client) *SessionCache {
	return &SessionCache{
		client: client,
		ttl:    5 * time.Minute,
	}
}

func (sc *SessionCache) SetOnline(ctx context.Context, userID string) error {
	key := "online:" + userID
	if err := sc.client.Set(ctx, key, "true", sc.ttl).Err(); err != nil {
		return fmt.Errorf("cache.SessionCache.SetOnline: %w", err)
	}
	return nil
}

func (sc *SessionCache) IsOnline(ctx context.Context, userID string) (bool, error) {
	key := "online:" + userID
	val, err := sc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("cache.SessionCache.IsOnline: %w", err)
	}
	return val == "true", nil
}

func (sc *SessionCache) SetOffline(ctx context.Context, userID string) error {
	key := "online:" + userID
	if err := sc.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache.SessionCache.SetOffline: %w", err)
	}
	return nil
}
