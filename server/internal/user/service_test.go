package user_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/google/uuid"

	"github.com/luketeo/horizon/internal/platform/testhelper"
	"github.com/luketeo/horizon/internal/user"
)

func strPtr(s string) *string { return &s }

func mustParseUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", s, err)
	}
	return id
}

// fakeClerkUser builds a *clerk.User sufficient for GetOrCreateUser tests.
// It wires a primary email, optional names, and optional avatar URL.
func fakeClerkUser(id, primaryEmail string, firstName, lastName, imgURL *string) *clerk.User {
	emailID := "eaddr_" + id
	return &clerk.User{
		ID:                    id,
		FirstName:             firstName,
		LastName:              lastName,
		ImageURL:              imgURL,
		PrimaryEmailAddressID: &emailID,
		EmailAddresses: []*clerk.EmailAddress{
			{ID: emailID, EmailAddress: primaryEmail},
		},
	}
}

func newUserService(t *testing.T) *user.Service {
	t.Helper()
	db := testhelper.DB(t)
	testhelper.Reset(t, db)
	return user.NewService(user.NewRepo(db), slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestGetOrCreateUser_InsertsNewUser(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	clerkU := fakeClerkUser(
		"user_insert_1",
		"alice@example.com",
		strPtr("Alice"),
		strPtr("Anderson"),
		strPtr("https://img.example.com/a.png"),
	)

	got, id, err := svc.GetOrCreateUser(ctx, clerkU)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}
	if id != got.Id {
		t.Fatalf("id mismatch: returned id=%s but got.Id=%s", id, got.Id)
	}
	if got.Email != "alice@example.com" {
		t.Errorf("email: want alice@example.com, got %q", got.Email)
	}
	if got.FirstName == nil || *got.FirstName != "Alice" {
		t.Errorf("first_name: want Alice, got %v", got.FirstName)
	}
	if got.LastName == nil || *got.LastName != "Anderson" {
		t.Errorf("last_name: want Anderson, got %v", got.LastName)
	}
	if got.AvatarUrl == nil || *got.AvatarUrl != "https://img.example.com/a.png" {
		t.Errorf("avatar_url: want set, got %v", got.AvatarUrl)
	}
	if got.LastLoginAt == nil {
		t.Error("last_login_at: want non-nil on insert")
	}
}

func TestGetOrCreateUser_UpsertBumpsLastLogin(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	clerkU := fakeClerkUser(
		"user_upsert_1",
		"bob@example.com",
		strPtr("Bob"),
		strPtr("Barker"),
		nil,
	)

	first, _, err := svc.GetOrCreateUser(ctx, clerkU)
	if err != nil {
		t.Fatalf("first GetOrCreateUser: %v", err)
	}
	if first.LastLoginAt == nil {
		t.Fatal("first LastLoginAt nil")
	}

	// Sleep just enough to move the clock past NOW() resolution.
	time.Sleep(10 * time.Millisecond)

	second, _, err := svc.GetOrCreateUser(ctx, clerkU)
	if err != nil {
		t.Fatalf("second GetOrCreateUser: %v", err)
	}
	if second.Id != first.Id {
		t.Errorf("upsert created a new row: id changed from %s to %s", first.Id, second.Id)
	}
	if second.LastLoginAt == nil || !second.LastLoginAt.After(*first.LastLoginAt) {
		t.Errorf(
			"last_login_at not bumped: first=%v second=%v",
			first.LastLoginAt,
			second.LastLoginAt,
		)
	}
	if !second.UpdatedAt.After(first.UpdatedAt) {
		t.Errorf("updated_at not bumped: first=%v second=%v", first.UpdatedAt, second.UpdatedAt)
	}
}

func TestGetOrCreateUser_CoalescesNilFieldsOnUpdate(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	initial := fakeClerkUser(
		"user_coalesce_1",
		"carol@example.com",
		strPtr("Carol"),
		strPtr("Cross"),
		strPtr("https://img.example.com/c.png"),
	)
	first, _, err := svc.GetOrCreateUser(ctx, initial)
	if err != nil {
		t.Fatalf("first GetOrCreateUser: %v", err)
	}

	// Re-upsert with nil first_name/last_name/avatar: the existing values should survive.
	clearedUpdate := fakeClerkUser(
		"user_coalesce_1",
		"carol@example.com",
		nil, nil, nil,
	)
	second, _, err := svc.GetOrCreateUser(ctx, clearedUpdate)
	if err != nil {
		t.Fatalf("second GetOrCreateUser: %v", err)
	}
	if second.FirstName == nil || *second.FirstName != "Carol" {
		t.Errorf(
			"first_name should be preserved: got %v want %v",
			second.FirstName,
			first.FirstName,
		)
	}
	if second.LastName == nil || *second.LastName != "Cross" {
		t.Errorf("last_name should be preserved: got %v want %v", second.LastName, first.LastName)
	}
	if second.AvatarUrl == nil || *second.AvatarUrl != "https://img.example.com/c.png" {
		t.Errorf(
			"avatar_url should be preserved: got %v want %v",
			second.AvatarUrl,
			first.AvatarUrl,
		)
	}
}

func TestUpdateUser_SetsProvidedFields(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	clerkU := fakeClerkUser(
		"user_update_1",
		"dan@example.com",
		strPtr("Dan"),
		strPtr("Dawson"),
		nil,
	)
	_, id, err := svc.GetOrCreateUser(ctx, clerkU)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	updated, err := svc.UpdateUser(ctx, id, strPtr("Daniel"), strPtr("Dobson"))
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if updated.FirstName == nil || *updated.FirstName != "Daniel" {
		t.Errorf("first_name: want Daniel, got %v", updated.FirstName)
	}
	if updated.LastName == nil || *updated.LastName != "Dobson" {
		t.Errorf("last_name: want Dobson, got %v", updated.LastName)
	}
}

func TestUpdateUser_CoalescesNilFields(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	clerkU := fakeClerkUser(
		"user_update_nil_1",
		"eve@example.com",
		strPtr("Eve"),
		strPtr("Evans"),
		nil,
	)
	_, id, err := svc.GetOrCreateUser(ctx, clerkU)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	updated, err := svc.UpdateUser(ctx, id, nil, strPtr("Ellis"))
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if updated.FirstName == nil || *updated.FirstName != "Eve" {
		t.Errorf("first_name should be preserved as Eve, got %v", updated.FirstName)
	}
	if updated.LastName == nil || *updated.LastName != "Ellis" {
		t.Errorf("last_name: want Ellis, got %v", updated.LastName)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	missingID := mustParseUUID(t, "00000000-0000-0000-0000-000000000001")
	_, err := svc.UpdateUser(ctx, missingID, strPtr("Ghost"), nil)
	if !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("UpdateUser missing row: want ErrNotFound, got %v", err)
	}
}

func TestGetUserIDByClerkID_Roundtrip(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	clerkU := fakeClerkUser(
		"user_lookup_1",
		"fay@example.com",
		strPtr("Fay"),
		strPtr("Fisher"),
		nil,
	)
	_, insertedID, err := svc.GetOrCreateUser(ctx, clerkU)
	if err != nil {
		t.Fatalf("GetOrCreateUser: %v", err)
	}

	got, err := svc.GetUserIDByClerkID(ctx, clerkU.ID)
	if err != nil {
		t.Fatalf("GetUserIDByClerkID: %v", err)
	}
	if got != insertedID {
		t.Errorf("lookup returned %s, want %s", got, insertedID)
	}
}

func TestGetUserIDByClerkID_NotFound(t *testing.T) {
	svc := newUserService(t)
	ctx := context.Background()

	_, err := svc.GetUserIDByClerkID(ctx, "user_missing_999")
	if !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
