package matchmaking

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Match struct {
	WhiteUserID uuid.UUID
	BlackUserID uuid.UUID
}

type Matcher struct {
	queue    QueueInterface
	interval time.Duration
}

func NewMatcher(client *redis.Client) *Matcher {
	return &Matcher{
		queue:    NewQueue(client),
		interval: 5 * time.Second,
	}
}

func newMatcherWithQueue(queue QueueInterface) *Matcher {
	return &Matcher{
		queue:    queue,
		interval: 5 * time.Second,
	}
}

func (m *Matcher) Start(ctx context.Context, matchCallback func(match *Match) error) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.tryMatch(ctx, matchCallback); err != nil {
				fmt.Printf("matcher error: %v\n", err)
			}
		}
	}
}

func (m *Matcher) tryMatch(ctx context.Context, matchCallback func(match *Match) error) error {
	players, err := m.queue.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("matchmaking.Matcher.tryMatch: %w", err)
	}

	fmt.Printf("[Matchmaking] queue size: %d\n", len(players))
	if len(players) < 2 {
		return nil
	}

	for i := 0; i < len(players)-1; i++ {
		player1 := players[i]

		for j := i + 1; j < len(players); j++ {
			player2 := players[j]

			if m.canMatch(player1, player2) {
				if err := m.performMatch(ctx, player1, player2, matchCallback); err != nil {
					return fmt.Errorf("matchmaking.Matcher.tryMatch: %w", err)
				}
				return nil
			}
		}
	}

	return nil
}

func (m *Matcher) canMatch(player1, player2 QueuedPlayer) bool {
	diff := player1.Rating - player2.Rating
	if diff < 0 {
		diff = -diff
	}

	initialRange := 200
	return diff <= initialRange
}

func (m *Matcher) performMatch(ctx context.Context, player1, player2 QueuedPlayer, matchCallback func(match *Match) error) error {
	if err := m.queue.Dequeue(ctx, player1.UserID); err != nil {
		return fmt.Errorf("matchmaking.Matcher.performMatch: dequeue player1: %w", err)
	}

	if err := m.queue.Dequeue(ctx, player2.UserID); err != nil {
		return fmt.Errorf("matchmaking.Matcher.performMatch: dequeue player2: %w", err)
	}

	match := &Match{
		WhiteUserID: player1.UserID,
		BlackUserID: player2.UserID,
	}

	if err := matchCallback(match); err != nil {
		if err := m.queue.Enqueue(ctx, player1.UserID, player1.Rating); err != nil {
			return fmt.Errorf("matchmaking.Matcher.performMatch: re-enqueue player1: %w", err)
		}
		if err := m.queue.Enqueue(ctx, player2.UserID, player2.Rating); err != nil {
			return fmt.Errorf("matchmaking.Matcher.performMatch: re-enqueue player2: %w", err)
		}
		return fmt.Errorf("matchmaking.Matcher.performMatch: match callback failed: %w", err)
	}

	return nil
}

func (m *Matcher) GetQueuePosition(ctx context.Context, userID uuid.UUID) (int, error) {
	return m.queue.GetPosition(ctx, userID)
}

func (m *Matcher) IsQueuedUser(ctx context.Context, userID uuid.UUID) (bool, error) {
	return m.queue.IsQueued(ctx, userID)
}
