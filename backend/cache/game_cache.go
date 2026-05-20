package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/chess-nfac/backend/models"
)

type GameCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewGameCache(client *redis.Client) *GameCache {
	return &GameCache{
		client: client,
		ttl:    2 * time.Hour,
	}
}

func (gc *GameCache) Set(ctx context.Context, gameID string, game *models.Game) error {
	key := "game:" + gameID
	data, err := json.Marshal(game)
	if err != nil {
		return fmt.Errorf("cache.GameCache.Set: %w", err)
	}

	if err := gc.client.Set(ctx, key, data, gc.ttl).Err(); err != nil {
		return fmt.Errorf("cache.GameCache.Set: %w", err)
	}

	return nil
}

func (gc *GameCache) Get(ctx context.Context, gameID string) (*models.Game, error) {
	key := "game:" + gameID
	val, err := gc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache.GameCache.Get: %w", err)
	}

	var game models.Game
	if err := json.Unmarshal([]byte(val), &game); err != nil {
		return nil, fmt.Errorf("cache.GameCache.Get: %w", err)
	}

	return &game, nil
}

func (gc *GameCache) Delete(ctx context.Context, gameID string) error {
	key := "game:" + gameID
	if err := gc.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("cache.GameCache.Delete: %w", err)
	}
	return nil
}
