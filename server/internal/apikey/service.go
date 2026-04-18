// Package apikey owns the organisation API-key domain: generation, listing, and
// revocation. Raw key material is generated here, hashed once, and returned to
// the caller once only.
package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/luketeo/horizon/generated/oapi"
)

// ErrNotFound indicates the API key does not exist or is already revoked.
var ErrNotFound = errors.New("api key not found")

// Service orchestrates API key operations.
type Service struct {
	repo   *Repo
	logger *slog.Logger
}

// NewService wires a Service with its repo and logger.
func NewService(repo *Repo, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// generateRawKey returns a freshly generated key (with the "hrz_" prefix) and
// its SHA-256 hash.
func generateRawKey() (rawKey, keyHash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generating random bytes: %w", err)
	}
	rawKey = "hrz_" + hex.EncodeToString(b)
	h := sha256.Sum256([]byte(rawKey))
	keyHash = hex.EncodeToString(h[:])
	return rawKey, keyHash, nil
}

// Create generates a new API key, stores its hash, and returns the raw key one
// time only alongside metadata.
func (s *Service) Create(
	ctx context.Context,
	orgID uuid.UUID,
	name string,
	scopes []string,
) (oapi.CreatedApiKey, error) {
	rawKey, keyHash, err := generateRawKey()
	if err != nil {
		return oapi.CreatedApiKey{}, err
	}
	k, err := s.repo.Insert(ctx, orgID, name, keyHash, scopes)
	if err != nil {
		return oapi.CreatedApiKey{}, err
	}
	k.Key = rawKey
	return k, nil
}

// List returns all active keys for an organisation.
func (s *Service) List(ctx context.Context, orgID uuid.UUID) ([]oapi.ApiKey, error) {
	return s.repo.ListActive(ctx, orgID)
}

// Revoke soft-deletes a key. Returns ErrNotFound if the key is unknown or was
// already revoked.
func (s *Service) Revoke(ctx context.Context, orgID, keyID uuid.UUID) error {
	return s.repo.Revoke(ctx, orgID, keyID)
}
