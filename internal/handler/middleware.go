package handler

import (
	"log/slog"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/sumire/issues/internal/domain"
	"github.com/sumire/issues/internal/service"
)

const (
	contextKeyUserID = "user_id"
)

// RequestLogger logs each HTTP request with structured fields.
func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			slog.Info("http request",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", c.Response().Status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID),
			)

			return err
		}
	}
}

// JWTAuth validates the Bearer token and injects the user ID into echo context.
func JWTAuth(auth *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return domain.ErrUnauthorized
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return domain.ErrUnauthorized
			}

			userID, err := auth.ValidateToken(parts[1])
			if err != nil {
				return domain.ErrUnauthorized
			}

			c.Set(contextKeyUserID, userID)
			return next(c)
		}
	}
}

// GetUserID extracts the authenticated user ID from echo context.
func GetUserID(c echo.Context) (int64, bool) {
	id, ok := c.Get(contextKeyUserID).(int64)
	return id, ok
}
