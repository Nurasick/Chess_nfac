package matchmaking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockQueue implements QueueInterface for testing — no real Redis needed.
type mockQueue struct {
	players    []QueuedPlayer
	enqueueLog []QueuedPlayer
	dequeueLog []uuid.UUID
	dequeueErr error
	enqueueErr error
	positions  map[uuid.UUID]int
	queuedSet  map[uuid.UUID]bool
}

func newMockQueue(players ...QueuedPlayer) *mockQueue {
	return &mockQueue{
		players:   players,
		positions: make(map[uuid.UUID]int),
		queuedSet: make(map[uuid.UUID]bool),
	}
}

func (m *mockQueue) Enqueue(_ context.Context, userID uuid.UUID, rating int) error {
	if m.enqueueErr != nil {
		return m.enqueueErr
	}
	m.enqueueLog = append(m.enqueueLog, QueuedPlayer{UserID: userID, Rating: rating})
	m.players = append(m.players, QueuedPlayer{UserID: userID, Rating: rating})
	return nil
}

func (m *mockQueue) Dequeue(_ context.Context, userID uuid.UUID) error {
	if m.dequeueErr != nil {
		return m.dequeueErr
	}
	m.dequeueLog = append(m.dequeueLog, userID)
	for i, p := range m.players {
		if p.UserID == userID {
			m.players = append(m.players[:i], m.players[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockQueue) GetAll(_ context.Context) ([]QueuedPlayer, error) {
	return m.players, nil
}

func (m *mockQueue) GetByRatingRange(_ context.Context, min, max int) ([]QueuedPlayer, error) {
	var result []QueuedPlayer
	for _, p := range m.players {
		if p.Rating >= min && p.Rating <= max {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockQueue) GetPosition(_ context.Context, userID uuid.UUID) (int, error) {
	pos, ok := m.positions[userID]
	if !ok {
		return -1, nil
	}
	return pos, nil
}

func (m *mockQueue) IsQueued(_ context.Context, userID uuid.UUID) (bool, error) {
	return m.queuedSet[userID], nil
}

func TestMatcher_canMatch(t *testing.T) {
	matcher := newMatcherWithQueue(newMockQueue())

	tests := []struct {
		name   string
		r1, r2 int
		want   bool
	}{
		{"within 200", 1200, 1350, true},
		{"exactly 200 apart", 1000, 1200, true},
		{"201 apart", 1000, 1201, false},
		{"same rating", 1500, 1500, true},
		{"higher minus lower", 1400, 1200, true},
		{"outside range reversed", 1201, 1000, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p1 := QueuedPlayer{UserID: uuid.New(), Rating: tc.r1}
			p2 := QueuedPlayer{UserID: uuid.New(), Rating: tc.r2}
			assert.Equal(t, tc.want, matcher.canMatch(p1, p2))
		})
	}
}

func TestMatcher_tryMatch(t *testing.T) {
	ctx := context.Background()

	t.Run("fewer than 2 players - no match", func(t *testing.T) {
		q := newMockQueue(QueuedPlayer{UserID: uuid.New(), Rating: 1200})
		m := newMatcherWithQueue(q)

		called := false
		err := m.tryMatch(ctx, func(*Match) error { called = true; return nil })
		require.NoError(t, err)
		assert.False(t, called)
	})

	t.Run("empty queue - no match", func(t *testing.T) {
		q := newMockQueue()
		m := newMatcherWithQueue(q)

		called := false
		err := m.tryMatch(ctx, func(*Match) error { called = true; return nil })
		require.NoError(t, err)
		assert.False(t, called)
	})

	t.Run("2 matching players - callback called", func(t *testing.T) {
		p1 := QueuedPlayer{UserID: uuid.New(), Rating: 1200}
		p2 := QueuedPlayer{UserID: uuid.New(), Rating: 1300}
		q := newMockQueue(p1, p2)
		m := newMatcherWithQueue(q)

		var got *Match
		err := m.tryMatch(ctx, func(match *Match) error { got = match; return nil })
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("2 players outside range - no match", func(t *testing.T) {
		p1 := QueuedPlayer{UserID: uuid.New(), Rating: 1000}
		p2 := QueuedPlayer{UserID: uuid.New(), Rating: 1500}
		q := newMockQueue(p1, p2)
		m := newMatcherWithQueue(q)

		called := false
		err := m.tryMatch(ctx, func(*Match) error { called = true; return nil })
		require.NoError(t, err)
		assert.False(t, called)
	})
}

func TestMatcher_performMatch(t *testing.T) {
	ctx := context.Background()

	p1 := QueuedPlayer{UserID: uuid.New(), Rating: 1200}
	p2 := QueuedPlayer{UserID: uuid.New(), Rating: 1300}

	t.Run("success - both dequeued", func(t *testing.T) {
		q := newMockQueue(p1, p2)
		m := newMatcherWithQueue(q)

		err := m.performMatch(ctx, p1, p2, func(*Match) error { return nil })
		require.NoError(t, err)

		assert.Equal(t, 2, len(q.dequeueLog))
		assert.Contains(t, q.dequeueLog, p1.UserID)
		assert.Contains(t, q.dequeueLog, p2.UserID)
		assert.Empty(t, q.enqueueLog)
	})

	t.Run("callback failure - both re-enqueued", func(t *testing.T) {
		q := newMockQueue(p1, p2)
		m := newMatcherWithQueue(q)

		err := m.performMatch(ctx, p1, p2, func(*Match) error {
			return errors.New("game creation failed")
		})
		require.Error(t, err)

		assert.Equal(t, 2, len(q.dequeueLog))
		require.Equal(t, 2, len(q.enqueueLog))
		assert.Contains(t, q.enqueueLog, p1)
		assert.Contains(t, q.enqueueLog, p2)
	})
}

func TestMatcher_GetQueuePosition(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("user in queue - correct rank", func(t *testing.T) {
		q := newMockQueue()
		q.positions[userID] = 3
		m := newMatcherWithQueue(q)

		pos, err := m.GetQueuePosition(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, 3, pos)
	})

	t.Run("user not in queue - returns -1", func(t *testing.T) {
		q := newMockQueue()
		m := newMatcherWithQueue(q)

		pos, err := m.GetQueuePosition(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, -1, pos)
	})
}

func TestMatcher_IsQueuedUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("queued user - returns true", func(t *testing.T) {
		q := newMockQueue()
		q.queuedSet[userID] = true
		m := newMatcherWithQueue(q)

		ok, err := m.IsQueuedUser(ctx, userID)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("non-queued user - returns false", func(t *testing.T) {
		q := newMockQueue()
		m := newMatcherWithQueue(q)

		ok, err := m.IsQueuedUser(ctx, userID)
		require.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestMatcher_Start_CancelContext(t *testing.T) {
	q := newMockQueue()
	m := newMatcherWithQueue(q)
	m.interval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		m.Start(ctx, func(*Match) error { return nil })
		close(done)
	}()

	cancel()
	<-done // Start must return after ctx is canceled
}

func TestMatcher_performMatch_DequeuePlayer1Error(t *testing.T) {
	ctx := context.Background()
	p1 := QueuedPlayer{UserID: uuid.New(), Rating: 1200}
	p2 := QueuedPlayer{UserID: uuid.New(), Rating: 1300}

	q := newMockQueue(p1, p2)
	q.dequeueErr = errors.New("redis error")
	m := newMatcherWithQueue(q)

	err := m.performMatch(ctx, p1, p2, func(*Match) error { return nil })
	require.Error(t, err)
}

func TestMatcher_tryMatch_PerformMatchError(t *testing.T) {
	ctx := context.Background()
	p1 := QueuedPlayer{UserID: uuid.New(), Rating: 1200}
	p2 := QueuedPlayer{UserID: uuid.New(), Rating: 1300}

	q := newMockQueue(p1, p2)
	q.dequeueErr = errors.New("redis error")
	m := newMatcherWithQueue(q)

	err := m.tryMatch(ctx, func(*Match) error { return nil })
	require.Error(t, err)
}
