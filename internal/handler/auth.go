package handler

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
func (h *AuthHandler) GoogleRedirect(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})
	http.Redirect(w, r, h.auth.GoogleAuthURL(state), http.StatusTemporaryRedirect)
}

// GoogleCallback handles the OAuth callback from Google.
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if err := validateOAuthState(r); err != nil {
		WriteError(w, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		WriteError(w, fmt.Errorf("%w: missing code parameter", domain.ErrInvalidInput))
		return
	}

	user, tokens, err := h.auth.GoogleCallback(r.Context(), code)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

// GitHubRedirect redirects the user to GitHub's OAuth consent page.
func (h *AuthHandler) GitHubRedirect(w http.ResponseWriter, r *http.Request) {
	state := generateState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})
	http.Redirect(w, r, h.auth.GitHubAuthURL(state), http.StatusTemporaryRedirect)
}

// GitHubCallback handles the OAuth callback from GitHub.
func (h *AuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	if err := validateOAuthState(r); err != nil {
		WriteError(w, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		WriteError(w, fmt.Errorf("%w: missing code parameter", domain.ErrInvalidInput))
		return
	}

	user, tokens, err := h.auth.GitHubCallback(r.Context(), code)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]any{
		"user":   user,
		"tokens": tokens,
	})
}

// Me returns the currently authenticated user.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		WriteError(w, domain.ErrUnauthorized)
		return
	}

	user, err := h.auth.GetUser(r.Context(), userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

// Refresh generates a new token pair from a refresh token.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, fmt.Errorf("%w: invalid request body", domain.ErrInvalidInput))
		return
	}

	if body.RefreshToken == "" {
		WriteError(w, fmt.Errorf("%w: refresh_token is required", domain.ErrInvalidInput))
		return
	}

	tokens, err := h.auth.RefreshAccessToken(body.RefreshToken)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, tokens)
}

// JWTAuth is middleware that validates the JWT Bearer token and injects the user ID into context.
func JWTAuth(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				WriteError(w, domain.ErrUnauthorized)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				WriteError(w, domain.ErrUnauthorized)
				return
			}

			userID, err := auth.ValidateToken(parts[1])
			if err != nil {
				WriteError(w, domain.ErrUnauthorized)
				return
			}

			ctx := SetUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func generateState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "fallback-state"
	}
	return base64.URLEncoding.EncodeToString(b)
}

func validateOAuthState(r *http.Request) error {
	cookie, err := r.Cookie("oauth_state")
	if err != nil {
		return fmt.Errorf("missing oauth_state cookie")
	}

	queryState := r.URL.Query().Get("state")
	if queryState == "" || queryState != cookie.Value {
		return fmt.Errorf("state mismatch")
	}

	return nil
}
