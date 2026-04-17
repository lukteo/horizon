package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/luketeo/horizon/generated/horizon/public/model"
	"github.com/luketeo/horizon/generated/horizon/public/table"
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

// toOapi maps a Jet model.Users row into the oapi.User DTO.
func toOapi(m model.Users) oapi.User {
	return oapi.User{
		Id:          m.ID,
		Email:       openapi_types.Email(m.Email),
		FirstName:   m.FirstName,
		LastName:    m.LastName,
		AvatarUrl:   m.AvatarURL,
		LastLoginAt: m.LastLoginAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// Upsert inserts or updates a user keyed on clerk_id and returns the resulting row.
func (r *Repo) Upsert(
	ctx context.Context,
	clerkID, email string,
	firstName, lastName, avatarURL *string,
) (oapi.User, error) {
	stmt := table.Users.
		INSERT(
			table.Users.ClerkID,
			table.Users.Email,
			table.Users.FirstName,
			table.Users.LastName,
			table.Users.AvatarURL,
			table.Users.LastLoginAt,
		).
		VALUES(
			clerkID,
			email,
			firstName,
			lastName,
			avatarURL,
			postgres.NOW(),
		).
		ON_CONFLICT(table.Users.ClerkID).
		DO_UPDATE(
			postgres.SET(
				table.Users.Email.SET(table.Users.EXCLUDED.Email),
				table.Users.FirstName.SET(postgres.StringExp(postgres.COALESCE(
					table.Users.EXCLUDED.FirstName, table.Users.FirstName,
				))),
				table.Users.LastName.SET(postgres.StringExp(postgres.COALESCE(
					table.Users.EXCLUDED.LastName, table.Users.LastName,
				))),
				table.Users.AvatarURL.SET(postgres.StringExp(postgres.COALESCE(
					table.Users.EXCLUDED.AvatarURL, table.Users.AvatarURL,
				))),
				table.Users.LastLoginAt.SET(postgres.NOW()),
				table.Users.UpdatedAt.SET(postgres.NOW()),
			),
		).
		RETURNING(table.Users.AllColumns)

	var out model.Users
	if err := stmt.QueryContext(ctx, r.db, &out); err != nil {
		return oapi.User{}, fmt.Errorf("upserting user: %w", err)
	}
	return toOapi(out), nil
}

// Update sets mutable profile fields on the user. Returns ErrNotFound if no row matches.
func (r *Repo) Update(
	ctx context.Context,
	userID uuid.UUID,
	firstName, lastName *string,
) (oapi.User, error) {
	var (
		firstExp postgres.StringExpression = table.Users.FirstName
		lastExp  postgres.StringExpression = table.Users.LastName
	)
	if firstName != nil {
		firstExp = postgres.String(*firstName)
	}
	if lastName != nil {
		lastExp = postgres.String(*lastName)
	}

	stmt := table.Users.
		UPDATE(table.Users.FirstName, table.Users.LastName, table.Users.UpdatedAt).
		SET(firstExp, lastExp, postgres.NOW()).
		WHERE(table.Users.ID.EQ(postgres.UUID(userID))).
		RETURNING(table.Users.AllColumns)

	var out model.Users
	if err := stmt.QueryContext(ctx, r.db, &out); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return oapi.User{}, ErrNotFound
		}
		return oapi.User{}, fmt.Errorf("updating user: %w", err)
	}
	return toOapi(out), nil
}

// GetIDByClerkID resolves the internal UUID for a Clerk user id.
func (r *Repo) GetIDByClerkID(ctx context.Context, clerkID string) (uuid.UUID, error) {
	stmt := postgres.
		SELECT(table.Users.ID).
		FROM(table.Users).
		WHERE(table.Users.ClerkID.EQ(postgres.String(clerkID))).
		LIMIT(1)

	var out model.Users
	if err := stmt.QueryContext(ctx, r.db, &out); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, fmt.Errorf("looking up user by clerk_id: %w", err)
	}
	return out.ID, nil
}

