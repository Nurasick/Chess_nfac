package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	Port        string
	Environment string
	LogLevel    string
	CORSOrigins []string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Port:        getEnvOrDefault("PORT", "8080"),
		Environment: getEnvOrDefault("ENVIRONMENT", "dev"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config.Load: DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return nil, fmt.Errorf("config.Load: REDIS_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("config.Load: JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("config.Load: JWT_SECRET must be at least 32 characters")
	}

	// Parse CORS origins
	corsOrigins := getEnvOrDefault("CORS_ORIGINS", "http://localhost:5173")
	cfg.CORSOrigins = strings.Split(corsOrigins, ",")
	for i := range cfg.CORSOrigins {
		cfg.CORSOrigins[i] = strings.TrimSpace(cfg.CORSOrigins[i])
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
