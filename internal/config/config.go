package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port        int
	DatabaseURL string
	JWTSecret   string

	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string

	ClaudeCodeBinary string
	ClaudeCodeTimeout time.Duration
	AIWorkerCount    int

	WebhookURL string

	FrontendURL string
}

// Load reads configuration from environment variables and validates required fields.
func Load() (Config, error) {
	port, err := getEnvInt("PORT", 8080)
	if err != nil {
		return Config{}, fmt.Errorf("parse PORT: %w", err)
	}

	timeout, err := getEnvDuration("CLAUDE_CODE_TIMEOUT", 30*time.Minute)
	if err != nil {
		return Config{}, fmt.Errorf("parse CLAUDE_CODE_TIMEOUT: %w", err)
	}

	workerCount, err := getEnvInt("AI_WORKER_COUNT", 3)
	if err != nil {
		return Config{}, fmt.Errorf("parse AI_WORKER_COUNT: %w", err)
	}

	cfg := Config{
		Port:               port,
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/issues?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		ClaudeCodeBinary:   getEnv("CLAUDE_CODE_BINARY", "claude"),
		ClaudeCodeTimeout:  timeout,
		AIWorkerCount:      workerCount,
		WebhookURL:         getEnv("WEBHOOK_URL", ""),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(v)
}

func getEnvDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue, nil
	}
	return time.ParseDuration(v)
}
