package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/sumire/issues/internal/config"
	"github.com/sumire/issues/internal/handler"
	"github.com/sumire/issues/internal/repository"
	"github.com/sumire/issues/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := sqlx.Connect("pgx", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	slog.Info("database connected")

	userRepo := repository.NewUserRepository(db)

	authSvc := service.NewAuthService(userRepo, service.AuthConfig{
		GoogleClientID:     cfg.GoogleClientID,
		GoogleClientSecret: cfg.GoogleClientSecret,
		GitHubClientID:     cfg.GitHubClientID,
		GitHubClientSecret: cfg.GitHubClientSecret,
		JWTSecret:          cfg.JWTSecret,
		FrontendURL:        cfg.FrontendURL,
	})

	authHandler := handler.NewAuthHandler(authSvc)

	e := echo.New()
	e.HideBanner = true
	e.Validator = handler.NewAppValidator()
	e.HTTPErrorHandler = handler.HTTPErrorHandler

	e.Use(middleware.RequestID())
	e.Use(handler.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{echo.HeaderXRequestID},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	e.GET("/health", func(c echo.Context) error {
		return handler.JSON(c, http.StatusOK, map[string]string{"status": "ok"})
	})

	v1 := e.Group("/api/v1")

	// Auth routes (public)
	auth := v1.Group("/auth")
	auth.GET("/google", authHandler.GoogleRedirect)
	auth.GET("/google/callback", authHandler.GoogleCallback)
	auth.GET("/github", authHandler.GitHubRedirect)
	auth.GET("/github/callback", authHandler.GitHubCallback)
	auth.POST("/refresh", authHandler.Refresh)

	// Protected routes
	protected := v1.Group("")
	protected.Use(handler.JWTAuth(authSvc))

	protected.GET("/auth/me", authHandler.Me)

	// TODO: project routes
	// TODO: issue routes
	// TODO: notification routes

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := e.Start(fmt.Sprintf(":%d", cfg.Port)); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	slog.Info("server stopped gracefully")
	return nil
}
