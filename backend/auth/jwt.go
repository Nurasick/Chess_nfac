package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID uuid.UUID, username, secret string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("auth.GenerateAccessToken: %w", err)
	}

	return tokenString, nil
}

func GenerateRefreshToken(userID uuid.UUID, secret string) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     claims.ExpiresAt.Unix(),
		"iat":     claims.IssuedAt.Unix(),
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("auth.GenerateRefreshToken: %w", err)
	}

	return tokenString, nil
}

func ParseToken(tokenStr, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth.ParseToken: unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("auth.ParseToken: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("auth.ParseToken: invalid token")
	}

	return claims, nil
}

func ParseRefreshToken(tokenStr, secret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth.ParseRefreshToken: unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("auth.ParseRefreshToken: %w", err)
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("auth.ParseRefreshToken: invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("auth.ParseRefreshToken: user_id claim not found")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("auth.ParseRefreshToken: %w", err)
	}

	return userID, nil
}
