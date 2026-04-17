package apikey_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/google/uuid"

	"github.com/luketeo/horizon/internal/apikey"
	"github.com/luketeo/horizon/internal/platform/testhelper"
	"github.com/luketeo/horizon/internal/services/orgservice"
	"github.com/luketeo/horizon/internal/user"
)

func strPtr(s string) *string { return &s }

func fakeClerkUser(id, primaryEmail string) *clerk.User {
	emailID := "eaddr_" + id
	return &clerk.User{
		ID:                    id,
		FirstName:             strPtr("First"),
		LastName:              strPtr("Last"),
		PrimaryEmailAddressID: &emailID,
		EmailAddresses: []*clerk.EmailAddress{
			{ID: emailID, EmailAddress: primaryEmail},
		},
	}
}

// seedOrg builds a user and an org owned by that user. Returns orgID.
func seedOrg(t *testing.T, db *sql.DB, clerkID, orgName string) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	userSvc := user.NewService(
		user.NewRepo(db),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	_, userID, err := userSvc.GetOrCreateUser(ctx, fakeClerkUser(clerkID, clerkID+"@example.com"))
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	orgSvc := orgservice.New(db)
	org, err := orgSvc.CreateOrg(ctx, orgName, nil, userID)
	if err != nil {
		t.Fatalf("seed org: %v", err)
	}
	return org.Id
}

func newSeededService(t *testing.T) (*apikey.Service, uuid.UUID) {
	t.Helper()
	db := testhelper.DB(t)
	testhelper.Reset(t, db)
	orgID := seedOrg(t, db, "user_apikey_seed", "Apikey Test Org")
	svc := apikey.NewService(
		apikey.NewRepo(db),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	return svc, orgID
}

func TestCreateAPIKey_ReturnsRawKeyAndStoresHash(t *testing.T) {
	svc, orgID := newSeededService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, orgID, "ingestor", []string{"read", "write"})
	if err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}
	if !strings.HasPrefix(created.Key, "hrz_") {
		t.Errorf("raw key prefix: want hrz_, got %q", created.Key)
	}
	if created.Name != "ingestor" {
		t.Errorf("name: want ingestor, got %q", created.Name)
	}
	if len(created.Scopes) != 2 {
		t.Errorf("scopes len: want 2, got %d", len(created.Scopes))
	}

	// The stored hash should be sha256(rawKey), and the raw key must not be in the DB.
	wantHash := sha256.Sum256([]byte(created.Key))
	wantHex := hex.EncodeToString(wantHash[:])

	db := testhelper.DB(t)
	var dbHash string
	if err := db.QueryRowContext(ctx,
		`SELECT key_hash FROM api_keys WHERE id = $1`, created.Id,
	).Scan(&dbHash); err != nil {
		t.Fatalf("select key_hash: %v", err)
	}
	if dbHash != wantHex {
		t.Errorf("stored key_hash mismatch:\nwant %s\n got %s", wantHex, dbHash)
	}
	if dbHash == created.Key {
		t.Error("raw key should not be stored verbatim")
	}
}

func TestListAPIKeys_ExcludesRevoked(t *testing.T) {
	svc, orgID := newSeededService(t)
	ctx := context.Background()

	active, err := svc.Create(ctx, orgID, "active", []string{"read"})
	if err != nil {
		t.Fatalf("create active key: %v", err)
	}
	revoked, err := svc.Create(ctx, orgID, "revoked", []string{"read"})
	if err != nil {
		t.Fatalf("create revoked key: %v", err)
	}
	if err := svc.Revoke(ctx, orgID, revoked.Id); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	keys, err := svc.List(ctx, orgID)
	if err != nil {
		t.Fatalf("ListAPIKeys: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("want 1 active key, got %d", len(keys))
	}
	if keys[0].Id != active.Id {
		t.Errorf("returned key id %s, want %s", keys[0].Id, active.Id)
	}
	if keys[0].Name != "active" {
		t.Errorf("name: want active, got %q", keys[0].Name)
	}
}

func TestRevokeAPIKey_UnknownKeyReturnsNotFound(t *testing.T) {
	svc, orgID := newSeededService(t)
	ctx := context.Background()

	missing := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	err := svc.Revoke(ctx, orgID, missing)
	if !errors.Is(err, apikey.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestRevokeAPIKey_IdempotentOnSecondCall(t *testing.T) {
	svc, orgID := newSeededService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, orgID, "once", []string{})
	if err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}
	if err := svc.Revoke(ctx, orgID, created.Id); err != nil {
		t.Fatalf("first revoke: %v", err)
	}
	err = svc.Revoke(ctx, orgID, created.Id)
	if !errors.Is(err, apikey.ErrNotFound) {
		t.Fatalf("second revoke: want ErrNotFound, got %v", err)
	}
}
