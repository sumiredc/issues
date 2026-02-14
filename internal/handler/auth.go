package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sumire/issues/internal/domain"
	"github.com/sumire/issues/internal/service"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// GoogleRedirect redirects the user to Google's OAuth consent page.
func (h *AuthHandler) GoogleRedirect(c echo.Context) error {
	state := generateState()
	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})
	return c.Redirect(http.StatusTemporaryRedirect, h.auth.GoogleAuthURL(state))
}

// GoogleCallback handles the OAuth callback from Google.
func (h *AuthHandler) GoogleCallback(c echo.Context) error {
	if err := validateOAuthState(c); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}

	code := c.QueryParam("code")
	if code == "" {
		return fmt.Errorf("%w: missing code parameter", domain.ErrInvalidInput)
	}

	user, tokens, err := h.auth.GoogleCallback(c.Request().Context(), code)
	if err != nil {
		return err
	}

	return JSON(c, http.StatusOK, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

// GitHubRedirect redirects the user to GitHub's OAuth consent page.
func (h *AuthHandler) GitHubRedirect(c echo.Context) error {
	state := generateState()
	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})
	return c.Redirect(http.StatusTemporaryRedirect, h.auth.GitHubAuthURL(state))
}

// GitHubCallback handles the OAuth callback from GitHub.
func (h *AuthHandler) GitHubCallback(c echo.Context) error {
	if err := validateOAuthState(c); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}

	code := c.QueryParam("code")
	if code == "" {
		return fmt.Errorf("%w: missing code parameter", domain.ErrInvalidInput)
	}

	user, tokens, err := h.auth.GitHubCallback(c.Request().Context(), code)
	if err != nil {
		return err
	}

	return JSON(c, http.StatusOK, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

// Me returns the currently authenticated user.
func (h *AuthHandler) Me(c echo.Context) error {
	userID, ok := GetUserID(c)
	if !ok {
		return domain.ErrUnauthorized
	}

	user, err := h.auth.GetUser(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	return JSON(c, http.StatusOK, user)
}

// refreshRequest is the request body for token refresh.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Refresh generates a new token pair from a refresh token.
func (h *AuthHandler) Refresh(c echo.Context) error {
	var body refreshRequest
	if err := c.Bind(&body); err != nil {
		return fmt.Errorf("%w: invalid request body", domain.ErrInvalidInput)
	}
	if err := c.Validate(body); err != nil {
		return err
	}

	tokens, err := h.auth.RefreshAccessToken(body.RefreshToken)
	if err != nil {
		return err
	}

	return JSON(c, http.StatusOK, tokens)
}

func generateState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "fallback-state"
	}
	return base64.URLEncoding.EncodeToString(b)
}

func validateOAuthState(c echo.Context) error {
	cookie, err := c.Cookie("oauth_state")
	if err != nil {
		return fmt.Errorf("missing oauth_state cookie")
	}

	queryState := c.QueryParam("state")
	if queryState == "" || queryState != cookie.Value {
		return fmt.Errorf("state mismatch")
	}

	return nil
}
