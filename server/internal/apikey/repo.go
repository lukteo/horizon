package apikey

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/luketeo/horizon/generated/horizon/public/model"
	"github.com/luketeo/horizon/generated/horizon/public/table"
	"github.com/luketeo/horizon/generated/oapi"
)

// Repo owns api_keys SQL and row→DTO mapping.
type Repo struct {
	db *sql.DB
}

// NewRepo wires a Repo to the given database.
func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// toOapiApiKey maps a Jet row into the list DTO.
func toOapiApiKey(m model.APIKeys) oapi.ApiKey {
	return oapi.ApiKey{
		Id:          m.ID,
		OrgId:       m.OrgID,
		Name:        m.Name,
		Scopes:      []string(m.Scopes),
		LastUsedAt:  m.LastUsedAt,
		RevokedAt:   m.RevokedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// Insert stores a new API key row keyed on orgID, returning the inserted metadata.
// The raw key value is not persisted — only its hash is.
func (r *Repo) Insert(
	ctx context.Context,
	orgID uuid.UUID,
	name, keyHash string,
	scopes []string,
) (oapi.CreatedApiKey, error) {
	stmt := table.APIKeys.
		INSERT(
			table.APIKeys.OrgID,
			table.APIKeys.Name,
			table.APIKeys.KeyHash,
			table.APIKeys.Scopes,
		).
		VALUES(orgID, name, keyHash, pq.StringArray(scopes)).
		RETURNING(
			table.APIKeys.ID,
			table.APIKeys.OrgID,
			table.APIKeys.Name,
			table.APIKeys.Scopes,
			table.APIKeys.CreatedAt,
			table.APIKeys.UpdatedAt,
		)

	var out model.APIKeys
	if err := stmt.QueryContext(ctx, r.db, &out); err != nil {
		return oapi.CreatedApiKey{}, fmt.Errorf("inserting api key: %w", err)
	}
	return oapi.CreatedApiKey{
		Id:        out.ID,
		OrgId:     out.OrgID,
		Name:      out.Name,
		Scopes:    []string(out.Scopes),
		CreatedAt: out.CreatedAt,
		UpdatedAt: out.UpdatedAt,
	}, nil
}

// ListActive returns all non-revoked keys for an organisation, newest first.
func (r *Repo) ListActive(ctx context.Context, orgID uuid.UUID) ([]oapi.ApiKey, error) {
	stmt := postgres.
		SELECT(table.APIKeys.AllColumns).
		FROM(table.APIKeys).
		WHERE(
			table.APIKeys.OrgID.EQ(postgres.UUID(orgID)).
				AND(table.APIKeys.RevokedAt.IS_NULL()),
		).
		ORDER_BY(table.APIKeys.CreatedAt.DESC())

	var rows []model.APIKeys
	if err := stmt.QueryContext(ctx, r.db, &rows); err != nil {
		return nil, fmt.Errorf("listing api keys: %w", err)
	}

	keys := make([]oapi.ApiKey, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, toOapiApiKey(row))
	}
	return keys, nil
}

// Revoke soft-deletes a key by stamping revoked_at. Returns ErrNotFound if the
// key is unknown or has already been revoked.
func (r *Repo) Revoke(ctx context.Context, orgID, keyID uuid.UUID) error {
	stmt := table.APIKeys.
		UPDATE(table.APIKeys.RevokedAt).
		SET(postgres.NOW()).
		WHERE(
			table.APIKeys.ID.EQ(postgres.UUID(keyID)).
				AND(table.APIKeys.OrgID.EQ(postgres.UUID(orgID))).
				AND(table.APIKeys.RevokedAt.IS_NULL()),
		)

	res, err := stmt.ExecContext(ctx, r.db)
	if err != nil {
		return fmt.Errorf("revoking api key: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

