package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/repository"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*models.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	u, _ := args.Get(0).(*models.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	u, _ := args.Get(0).(*models.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) UpdateRating(ctx context.Context, userID uuid.UUID, newRating int) error {
	args := m.Called(ctx, userID, newRating)
	return args.Error(0)
}

func (m *mockUserRepo) SaveRefreshToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, tokenHash, expiresAt)
	return args.Error(0)
}

func (m *mockUserRepo) FindRefreshToken(ctx context.Context, tokenHash string) (*repository.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	rt, _ := args.Get(0).(*repository.RefreshToken)
	return rt, args.Error(1)
}

func (m *mockUserRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

func (m *mockUserRepo) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
