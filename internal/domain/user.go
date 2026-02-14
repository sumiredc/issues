package domain

import "time"

// AuthProvider represents an OAuth provider.
type AuthProvider string

const (
	AuthProviderGoogle AuthProvider = "google"
	AuthProviderGitHub AuthProvider = "github"
)

// User represents an authenticated user.
type User struct {
	ID          int64        `json:"id" db:"id"`
	Provider    AuthProvider `json:"provider" db:"provider"`
	ProviderID  string       `json:"provider_id" db:"provider_id"`
	Email       string       `json:"email" db:"email"`
	DisplayName string       `json:"display_name" db:"display_name"`
	AvatarURL   *string      `json:"avatar_url,omitempty" db:"avatar_url"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
}
