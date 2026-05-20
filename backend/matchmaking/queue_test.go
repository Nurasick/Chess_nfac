package matchmaking

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestQueue(t *testing.T) (*Queue, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return NewQueue(client), mr
}

func TestQueue_Enqueue_Dequeue(t *testing.T) {
	ctx := context.Background()
	q, _ := newTestQueue(t)
	userID := uuid.New()

	require.NoError(t, q.Enqueue(ctx, userID, 1200))

	queued, err := q.IsQueued(ctx, userID)
	require.NoError(t, err)
	assert.True(t, queued)

	require.NoError(t, q.Dequeue(ctx, userID))

	queued, err = q.IsQueued(ctx, userID)
	require.NoError(t, err)
	assert.False(t, queued)
}

func TestQueue_GetAll(t *testing.T) {
	ctx := context.Background()
	q, _ := newTestQueue(t)

	p1 := QueuedPlayer{UserID: uuid.New(), Rating: 1000}
	p2 := QueuedPlayer{UserID: uuid.New(), Rating: 1200}

	require.NoError(t, q.Enqueue(ctx, p1.UserID, p1.Rating))
	require.NoError(t, q.Enqueue(ctx, p2.UserID, p2.Rating))

	players, err := q.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, players, 2)
}

func TestQueue_GetAll_Empty(t *testing.T) {
	ctx := context.Background()
	q, _ := newTestQueue(t)

	players, err := q.GetAll(ctx)
	require.NoError(t, err)
	assert.Empty(t, players)
}

func TestQueue_GetPosition(t *testing.T) {
	ctx := context.Background()
	q, _ := newTestQueue(t)
	userID := uuid.New()

	pos, err := q.GetPosition(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, -1, pos)

	require.NoError(t, q.Enqueue(ctx, userID, 1200))

	pos, err = q.GetPosition(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 1, pos)
}

func TestQueue_IsQueued(t *testing.T) {
	ctx := context.Background()
	q, _ := newTestQueue(t)
	userID := uuid.New()

	queued, err := q.IsQueued(ctx, userID)
	require.NoError(t, err)
	assert.False(t, queued)

	require.NoError(t, q.Enqueue(ctx, userID, 1200))

	queued, err = q.IsQueued(ctx, userID)
	require.NoError(t, err)
	assert.True(t, queued)
}

func TestQueue_GetByRatingRange(t *testing.T) {
	ctx := context.Background()
	q, _ := newTestQueue(t)

	p1 := QueuedPlayer{UserID: uuid.New(), Rating: 1000}
	p2 := QueuedPlayer{UserID: uuid.New(), Rating: 1200}
	p3 := QueuedPlayer{UserID: uuid.New(), Rating: 1500}

	require.NoError(t, q.Enqueue(ctx, p1.UserID, p1.Rating))
	require.NoError(t, q.Enqueue(ctx, p2.UserID, p2.Rating))
	require.NoError(t, q.Enqueue(ctx, p3.UserID, p3.Rating))

	players, err := q.GetByRatingRange(ctx, 1000, 1300)
	require.NoError(t, err)
	assert.Len(t, players, 2)

	for _, p := range players {
		assert.True(t, p.Rating >= 1000 && p.Rating <= 1300)
	}
}

func TestQueue_GetByRatingRange_Empty(t *testing.T) {
	ctx := context.Background()
	q, _ := newTestQueue(t)

	players, err := q.GetByRatingRange(ctx, 1000, 1200)
	require.NoError(t, err)
	assert.Empty(t, players)
}
