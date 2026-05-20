package matchmaking

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// QueueInterface allows Matcher to be unit-tested without a real Redis instance.
type QueueInterface interface {
	Enqueue(ctx context.Context, userID uuid.UUID, rating int) error
	Dequeue(ctx context.Context, userID uuid.UUID) error
	GetAll(ctx context.Context) ([]QueuedPlayer, error)
	GetByRatingRange(ctx context.Context, minRating, maxRating int) ([]QueuedPlayer, error)
	GetPosition(ctx context.Context, userID uuid.UUID) (int, error)
	IsQueued(ctx context.Context, userID uuid.UUID) (bool, error)
}

type QueuedPlayer struct {
	UserID uuid.UUID
	Rating int
}

type Queue struct {
	client *redis.Client
}

func NewQueue(client *redis.Client) *Queue {
	return &Queue{client: client}
}

func (q *Queue) Enqueue(ctx context.Context, userID uuid.UUID, rating int) error {
	key := "matchmaking_queue"
	score := float64(rating)

	if err := q.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: userID.String(),
	}).Err(); err != nil {
		return fmt.Errorf("matchmaking.Queue.Enqueue: %w", err)
	}

	return nil
}

func (q *Queue) Dequeue(ctx context.Context, userID uuid.UUID) error {
	key := "matchmaking_queue"

	if err := q.client.ZRem(ctx, key, userID.String()).Err(); err != nil {
		return fmt.Errorf("matchmaking.Queue.Dequeue: %w", err)
	}

	return nil
}

func (q *Queue) GetAll(ctx context.Context) ([]QueuedPlayer, error) {
	key := "matchmaking_queue"

	results, err := q.client.ZRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("matchmaking.Queue.GetAll: %w", err)
	}

	var players []QueuedPlayer
	for _, result := range results {
		userID, err := uuid.Parse(result.Member.(string))
		if err != nil {
			continue
		}

		players = append(players, QueuedPlayer{
			UserID: userID,
			Rating: int(result.Score),
		})
	}

	return players, nil
}

func (q *Queue) GetPosition(ctx context.Context, userID uuid.UUID) (int, error) {
	key := "matchmaking_queue"
	rank, err := q.client.ZRank(ctx, key, userID.String()).Result()
	if err == redis.Nil {
		return -1, nil
	}
	if err != nil {
		return -1, fmt.Errorf("matchmaking.Queue.GetPosition: %w", err)
	}
	return int(rank) + 1, nil
}

func (q *Queue) IsQueued(ctx context.Context, userID uuid.UUID) (bool, error) {
	key := "matchmaking_queue"
	_, err := q.client.ZScore(ctx, key, userID.String()).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("matchmaking.Queue.IsQueued: %w", err)
	}
	return true, nil
}

func (q *Queue) GetByRatingRange(ctx context.Context, minRating, maxRating int) ([]QueuedPlayer, error) {
	key := "matchmaking_queue"

	results, err := q.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: strconv.Itoa(minRating),
		Max: strconv.Itoa(maxRating),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("matchmaking.Queue.GetByRatingRange: %w", err)
	}

	var players []QueuedPlayer
	for _, member := range results {
		userID, err := uuid.Parse(member)
		if err != nil {
			continue
		}

		score, err := q.client.ZScore(ctx, key, member).Result()
		if err != nil {
			continue
		}

		players = append(players, QueuedPlayer{
			UserID: userID,
			Rating: int(score),
		})
	}

	return players, nil
}
