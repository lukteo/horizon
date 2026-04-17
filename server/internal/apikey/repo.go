package apikey

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"

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

// Insert stores a new API key row keyed on orgID, returning the inserted metadata.
// The raw key value is not persisted — only its hash is.
func (r *Repo) Insert(
	ctx context.Context,
	orgID uuid.UUID,
	name, keyHash string,
	scopes []string,
) (oapi.CreatedApiKey, error) {
	const q = `
		INSERT INTO api_keys (org_id, name, key_hash, scopes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, org_id, name, scopes, created_at, updated_at
	`
	var (
		k        oapi.CreatedApiKey
		dbScopes pq.StringArray
	)
	err := r.db.QueryRowContext(ctx, q, orgID, name, keyHash, pq.Array(scopes)).
		Scan(&k.Id, &k.OrgId, &k.Name, &dbScopes, &k.CreatedAt, &k.UpdatedAt)
	if err != nil {
		return oapi.CreatedApiKey{}, fmt.Errorf("inserting api key: %w", err)
	}
	k.Scopes = []string(dbScopes)
	return k, nil
}

// ListActive returns all non-revoked keys for an organisation, newest first.
func (r *Repo) ListActive(ctx context.Context, orgID uuid.UUID) ([]oapi.ApiKey, error) {
	const q = `
		SELECT id, org_id, name, scopes, last_used_at, revoked_at, created_at, updated_at
		FROM api_keys
		WHERE org_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, q, orgID)
	if err != nil {
		return nil, fmt.Errorf("listing api keys: %w", err)
	}
	defer rows.Close()

	var keys []oapi.ApiKey
	for rows.Next() {
		var (
			k        oapi.ApiKey
			dbScopes pq.StringArray
		)
		if err := rows.Scan(&k.Id, &k.OrgId, &k.Name, &dbScopes,
			&k.LastUsedAt, &k.RevokedAt, &k.CreatedAt, &k.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning api key: %w", err)
		}
		k.Scopes = []string(dbScopes)
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// Revoke soft-deletes a key by stamping revoked_at. Returns ErrNotFound if the
// key is unknown or has already been revoked.
func (r *Repo) Revoke(ctx context.Context, orgID, keyID uuid.UUID) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE api_keys SET revoked_at = NOW() WHERE id = $1 AND org_id = $2 AND revoked_at IS NULL`,
		keyID,
		orgID,
	)
	if err != nil {
		return fmt.Errorf("revoking api key: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
