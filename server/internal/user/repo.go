package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/luketeo/horizon/generated/oapi"
)

// Repo owns the user-table SQL and row→DTO mapping.
type Repo struct {
	db *sql.DB
}

// NewRepo builds a Repo backed by the given database handle.
func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// scanUser maps a row of user columns into an oapi.User.
// Expected column order: id, email, first_name, last_name, avatar_url, last_login_at, created_at, updated_at
func scanUser(dest *oapi.User, scanner interface {
	Scan(dest ...any) error
}) error {
	return scanner.Scan(
		&dest.Id, &dest.Email, &dest.FirstName, &dest.LastName,
		&dest.AvatarUrl, &dest.LastLoginAt, &dest.CreatedAt, &dest.UpdatedAt,
	)
}

// Upsert inserts or updates a user keyed on clerk_id and returns the resulting row.
func (r *Repo) Upsert(
	ctx context.Context,
	clerkID, email string,
	firstName, lastName, avatarURL *string,
) (oapi.User, error) {
	const q = `
		INSERT INTO users (clerk_id, email, first_name, last_name, avatar_url, last_login_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (clerk_id) DO UPDATE
		SET email         = EXCLUDED.email,
		    first_name    = COALESCE(EXCLUDED.first_name, users.first_name),
		    last_name     = COALESCE(EXCLUDED.last_name,  users.last_name),
		    avatar_url    = COALESCE(EXCLUDED.avatar_url, users.avatar_url),
		    last_login_at = NOW(),
		    updated_at    = NOW()
		RETURNING id, email, first_name, last_name, avatar_url, last_login_at, created_at, updated_at
	`
	var u oapi.User
	if err := scanUser(&u, r.db.QueryRowContext(ctx, q,
		clerkID, email, firstName, lastName, avatarURL,
	)); err != nil {
		return oapi.User{}, fmt.Errorf("upserting user: %w", err)
	}
	return u, nil
}

// Update sets mutable profile fields on the user. Returns ErrNotFound if no row matches.
func (r *Repo) Update(
	ctx context.Context,
	userID uuid.UUID,
	firstName, lastName *string,
) (oapi.User, error) {
	const q = `
		UPDATE users SET
		    first_name = COALESCE($1, first_name),
		    last_name  = COALESCE($2, last_name),
		    updated_at = NOW()
		WHERE id = $3
		RETURNING id, email, first_name, last_name, avatar_url, last_login_at, created_at, updated_at
	`
	var u oapi.User
	if err := scanUser(&u, r.db.QueryRowContext(ctx, q, firstName, lastName, userID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return oapi.User{}, ErrNotFound
		}
		return oapi.User{}, fmt.Errorf("updating user: %w", err)
	}
	return u, nil
}

// GetIDByClerkID resolves the internal UUID for a Clerk user id.
func (r *Repo) GetIDByClerkID(ctx context.Context, clerkID string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE clerk_id = $1`, clerkID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("looking up user by clerk_id: %w", err)
	}
	return id, nil
}
