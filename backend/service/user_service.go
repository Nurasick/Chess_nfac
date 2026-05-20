package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/chess-nfac/backend/auth"
	"github.com/chess-nfac/backend/models"
	"github.com/chess-nfac/backend/repository"
	"github.com/chess-nfac/backend/utils"
)

type UserService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewUserService(userRepo repository.UserRepository, jwtSecret string) *UserService {
	return &UserService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum)
}

func (s *UserService) Register(ctx context.Context, username, email, password, city string) (*models.User, error) {
	if err := utils.ValidateUsername(username); err != nil {
		return nil, utils.NewAppError("validation_error", err.Error(), 400)
	}
	if err := utils.ValidateEmail(email); err != nil {
		return nil, utils.NewAppError("validation_error", err.Error(), 400)
	}
	if err := utils.ValidatePassword(password); err != nil {
		return nil, utils.NewAppError("validation_error", err.Error(), 400)
	}
	if err := utils.ValidateCity(city); err != nil {
		return nil, utils.NewAppError("validation_error", err.Error(), 400)
	}

	existing, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("user_service.Register: check existing username: %w", err)
	}
	if existing != nil {
		return nil, utils.NewAppError("user_exists", "Username already taken", 409)
	}

	existing, err = s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user_service.Register: check existing email: %w", err)
	}
	if existing != nil {
		return nil, utils.NewAppError("email_exists", "Email already registered", 409)
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("user_service.Register: hash password: %w", err)
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		City:         city,
		Rating:       1200,
		GamesPlayed:  0,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("user_service.Register: create user: %w", err)
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (*models.User, string, string, error) {
	if username == "" || password == "" {
		return nil, "", "", utils.NewAppError("validation_error", "Username and password are required", 400)
	}

	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, "", "", fmt.Errorf("user_service.Login: get user: %w", err)
	}
	if user == nil {
		return nil, "", "", utils.NewAppError("invalid_credentials", "Invalid username or password", 401)
	}

	if err := auth.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, "", "", utils.NewAppError("invalid_credentials", "Invalid username or password", 401)
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Username, s.jwtSecret)
	if err != nil {
		return nil, "", "", fmt.Errorf("user_service.Login: generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, "", "", fmt.Errorf("user_service.Login: generate refresh token: %w", err)
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	tokenHash := hashRefreshToken(refreshToken)
	if err := s.userRepo.SaveRefreshToken(ctx, user.ID, tokenHash, expiresAt); err != nil {
		return nil, "", "", fmt.Errorf("user_service.Login: save refresh token: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

func (s *UserService) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	tokenHash := hashRefreshToken(refreshToken)

	rt, err := s.userRepo.FindRefreshToken(ctx, tokenHash)
	if err != nil {
		return "", "", fmt.Errorf("user_service.RefreshTokens: find refresh token: %w", err)
	}
	if rt == nil || rt.Revoked || rt.ExpiresAt.Before(time.Now()) {
		return "", "", utils.NewAppError("invalid_token", "Invalid or expired refresh token", 401)
	}

	if err := s.userRepo.RevokeRefreshToken(ctx, tokenHash); err != nil {
		return "", "", fmt.Errorf("user_service.RefreshTokens: revoke old token: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, rt.UserID)
	if err != nil {
		return "", "", fmt.Errorf("user_service.RefreshTokens: get user: %w", err)
	}
	if user == nil {
		return "", "", utils.NewAppError("user_not_found", "User not found", 404)
	}

	accessToken, err := auth.GenerateAccessToken(user.ID, user.Username, s.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("user_service.RefreshTokens: generate access token: %w", err)
	}

	newRefreshToken, err := auth.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("user_service.RefreshTokens: generate refresh token: %w", err)
	}

	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)
	newTokenHash := hashRefreshToken(newRefreshToken)
	if err := s.userRepo.SaveRefreshToken(ctx, user.ID, newTokenHash, newExpiresAt); err != nil {
		return "", "", fmt.Errorf("user_service.RefreshTokens: save refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

func (s *UserService) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user_service.GetByID: %w", err)
	}
	if user == nil {
		return nil, utils.NewAppError("user_not_found", "User not found", 404)
	}
	return user, nil
}

func (s *UserService) Logout(ctx context.Context, userID uuid.UUID) error {
	if err := s.userRepo.RevokeAllUserTokens(ctx, userID); err != nil {
		return fmt.Errorf("user_service.Logout: %w", err)
	}
	return nil
}
