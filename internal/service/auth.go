package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	googleOAuth "golang.org/x/oauth2/google"

	"github.com/sumire/issues/internal/domain"
)

// UserStore defines the user data access interface consumed by AuthService.
type UserStore interface {
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	FindByProviderID(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error)
	Upsert(ctx context.Context, user domain.User) (*domain.User, error)
}

// AuthConfig holds OAuth configuration.
type AuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	JWTSecret          string
	FrontendURL        string
}

// AuthService handles authentication logic.
type AuthService struct {
	users     UserStore
	jwtSecret []byte
	google    *oauth2.Config
	github    *oauth2.Config
}

// NewAuthService creates a new AuthService.
func NewAuthService(users UserStore, cfg AuthConfig) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: []byte(cfg.JWTSecret),
		google: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			Endpoint:     googleOAuth.Endpoint,
			Scopes:       []string{"openid", "profile", "email"},
			RedirectURL:  cfg.FrontendURL + "/auth/google/callback",
		},
		github: &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			Endpoint:     github.Endpoint,
			Scopes:       []string{"user:email"},
			RedirectURL:  cfg.FrontendURL + "/auth/github/callback",
		},
	}
}

// GoogleAuthURL returns the Google OAuth authorization URL.
func (s *AuthService) GoogleAuthURL(state string) string {
	return s.google.AuthCodeURL(state)
}

// GitHubAuthURL returns the GitHub OAuth authorization URL.
func (s *AuthService) GitHubAuthURL(state string) string {
	return s.github.AuthCodeURL(state)
}

// TokenPair holds an access token and refresh token.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// GoogleCallback exchanges the authorization code and returns a JWT pair.
func (s *AuthService) GoogleCallback(ctx context.Context, code string) (*domain.User, *TokenPair, error) {
	token, err := s.google.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("google token exchange: %w", err)
	}

	userInfo, err := fetchGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch google user info: %w", err)
	}

	user, err := s.users.Upsert(ctx, domain.User{
		Provider:    domain.AuthProviderGoogle,
		ProviderID:  userInfo.ID,
		Email:       userInfo.Email,
		DisplayName: userInfo.Name,
		AvatarURL:   strPtr(userInfo.Picture),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("upsert google user: %w", err)
	}

	pair, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, pair, nil
}

// GitHubCallback exchanges the authorization code and returns a JWT pair.
func (s *AuthService) GitHubCallback(ctx context.Context, code string) (*domain.User, *TokenPair, error) {
	token, err := s.github.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("github token exchange: %w", err)
	}

	userInfo, err := fetchGitHubUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch github user info: %w", err)
	}

	user, err := s.users.Upsert(ctx, domain.User{
		Provider:    domain.AuthProviderGitHub,
		ProviderID:  fmt.Sprintf("%d", userInfo.ID),
		Email:       userInfo.Email,
		DisplayName: userInfo.Login,
		AvatarURL:   strPtr(userInfo.AvatarURL),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("upsert github user: %w", err)
	}

	pair, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, nil, err
	}

	return user, pair, nil
}

// ValidateToken validates a JWT access token and returns the user ID.
func (s *AuthService) ValidateToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return 0, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, domain.ErrUnauthorized
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "access" {
		return 0, domain.ErrUnauthorized
	}

	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return 0, domain.ErrUnauthorized
	}

	return int64(userIDFloat), nil
}

// RefreshAccessToken validates a refresh token and returns a new token pair.
func (s *AuthService) RefreshAccessToken(refreshToken string) (*TokenPair, error) {
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrUnauthorized
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return nil, domain.ErrUnauthorized
	}

	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	return s.generateTokenPair(int64(userIDFloat))
}

// GetUser retrieves a user by ID.
func (s *AuthService) GetUser(ctx context.Context, userID int64) (*domain.User, error) {
	return s.users.FindByID(ctx, userID)
}

func (s *AuthService) generateTokenPair(userID int64) (*TokenPair, error) {
	now := time.Now()

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"type": "access",
		"iat":  now.Unix(),
		"exp":  now.Add(15 * time.Minute).Unix(),
	})
	accessStr, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"type": "refresh",
		"iat":  now.Unix(),
		"exp":  now.Add(7 * 24 * time.Hour).Unix(),
	})
	refreshStr, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}

type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func fetchGoogleUserInfo(ctx context.Context, accessToken string) (*googleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google user info returned status %d", resp.StatusCode)
	}

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode user info: %w", err)
	}
	return &info, nil
}

type githubUserInfo struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func fetchGitHubUserInfo(ctx context.Context, accessToken string) (*githubUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github user info returned status %d", resp.StatusCode)
	}

	var info githubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode user info: %w", err)
	}

	if info.Email == "" {
		email, err := fetchGitHubPrimaryEmail(ctx, accessToken)
		if err != nil {
			return nil, err
		}
		info.Email = email
	}

	return &info, nil
}

type githubEmail struct {
	Email   string `json:"email"`
	Primary bool   `json:"primary"`
}

func fetchGitHubPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.github.com/user/emails", nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch emails: %w", err)
	}
	defer resp.Body.Close()

	var emails []githubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("decode emails: %w", err)
	}

	for _, e := range emails {
		if e.Primary {
			return e.Email, nil
		}
	}

	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", fmt.Errorf("no email found for github user")
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
