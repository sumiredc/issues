package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/sumire/issues/internal/domain"
)

// UserRepository handles user data access operations.
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID retrieves a user by their ID.
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user,
		`SELECT id, provider, provider_id, email, display_name, avatar_url, created_at, updated_at
		 FROM users WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find user by id %d: %w", id, err)
	}
	return &user, nil
}

// FindByProviderID retrieves a user by their OAuth provider and provider ID.
func (r *UserRepository) FindByProviderID(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user,
		`SELECT id, provider, provider_id, email, display_name, avatar_url, created_at, updated_at
		 FROM users WHERE provider = $1 AND provider_id = $2`, provider, providerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("find user by provider %s/%s: %w", provider, providerID, err)
	}
	return &user, nil
}

// Upsert creates a new user or updates an existing one based on provider + provider_id.
// Returns the created or updated user.
func (r *UserRepository) Upsert(ctx context.Context, user domain.User) (*domain.User, error) {
	var result domain.User
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO users (provider, provider_id, email, display_name, avatar_url)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (provider, provider_id)
		 DO UPDATE SET email = EXCLUDED.email,
		               display_name = EXCLUDED.display_name,
		               avatar_url = EXCLUDED.avatar_url,
		               updated_at = NOW()
		 RETURNING id, provider, provider_id, email, display_name, avatar_url, created_at, updated_at`,
		user.Provider, user.ProviderID, user.Email, user.DisplayName, user.AvatarURL,
	).StructScan(&result)
	if err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}
	return &result, nil
}
