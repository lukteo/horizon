package org_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/google/uuid"

	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/org"
	"github.com/luketeo/horizon/internal/platform/testhelper"
	"github.com/luketeo/horizon/internal/user"
)

func strPtr(s string) *string { return &s }

func fakeClerkUser(id, email string) *clerk.User {
	emailID := "eaddr_" + id
	return &clerk.User{
		ID:                    id,
		FirstName:             strPtr("First"),
		LastName:              strPtr("Last"),
		PrimaryEmailAddressID: &emailID,
		EmailAddresses: []*clerk.EmailAddress{
			{ID: emailID, EmailAddress: email},
		},
	}
}

func seedUser(t *testing.T, db *sql.DB, clerkID, email string) uuid.UUID {
	t.Helper()
	userSvc := user.NewService(user.NewRepo(db), slog.New(slog.NewTextHandler(io.Discard, nil)))
	_, id, err := userSvc.GetOrCreateUser(context.Background(), fakeClerkUser(clerkID, email))
	if err != nil {
		t.Fatalf("seed user %s: %v", clerkID, err)
	}
	return id
}

func newOrgService(t *testing.T) (*org.Service, *sql.DB) {
	t.Helper()
	db := testhelper.DB(t)
	testhelper.Reset(t, db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return org.NewService(org.NewRepo(db), logger), db
}

func TestCreateOrg_AssignsOwnerMembership(t *testing.T) {
	svc, db := newOrgService(t)
	ctx := context.Background()

	userID := seedUser(t, db, "user_create_org", "co@example.com")
	o, err := svc.CreateOrg(ctx, "Acme", nil, userID)
	if err != nil {
		t.Fatalf("CreateOrg: %v", err)
	}

	if o.Slug != "acme" {
		t.Errorf("slug: want acme, got %q", o.Slug)
	}
	if o.MyRole == nil || *o.MyRole != oapi.Owner {
		t.Errorf("MyRole: want owner, got %v", o.MyRole)
	}
	if o.MemberCount == nil || *o.MemberCount != 1 {
		t.Errorf("MemberCount: want 1, got %v", o.MemberCount)
	}

	role, err := svc.GetMembership(ctx, o.Id, userID)
	if err != nil {
		t.Fatalf("GetMembership: %v", err)
	}
	if role != oapi.Owner {
		t.Errorf("persisted role: want owner, got %q", role)
	}
}

func TestCreateOrg_SlugConflictRetries(t *testing.T) {
	svc, db := newOrgService(t)
	ctx := context.Background()

	userID := seedUser(t, db, "user_slug_conflict", "sc@example.com")

	first, err := svc.CreateOrg(ctx, "Acme", nil, userID)
	if err != nil {
		t.Fatalf("first CreateOrg: %v", err)
	}
	second, err := svc.CreateOrg(ctx, "Acme", nil, userID)
	if err != nil {
		t.Fatalf("second CreateOrg: %v", err)
	}
	if first.Slug == second.Slug {
		t.Fatalf("expected retry suffix, got duplicate slug %q", first.Slug)
	}
	if second.Slug != "acme-2" {
		t.Errorf("slug: want acme-2, got %q", second.Slug)
	}
}

func TestCreateOrg_UsesSlugHintWhenProvided(t *testing.T) {
	svc, db := newOrgService(t)
	ctx := context.Background()

	userID := seedUser(t, db, "user_slug_hint", "sh@example.com")
	hint := "custom-slug"
	o, err := svc.CreateOrg(ctx, "Arbitrary Name", &hint, userID)
	if err != nil {
		t.Fatalf("CreateOrg: %v", err)
	}
	if o.Slug != hint {
		t.Errorf("slug: want %q, got %q", hint, o.Slug)
	}
}

func TestListOrgsForUser_ReturnsOnlyMemberOrgs(t *testing.T) {
	svc, db := newOrgService(t)
	ctx := context.Background()

	alice := seedUser(t, db, "user_alice", "alice@example.com")
	bob := seedUser(t, db, "user_bob", "bob@example.com")

	aliceOrg, err := svc.CreateOrg(ctx, "AliceCo", nil, alice)
	if err != nil {
		t.Fatalf("alice org: %v", err)
	}
	_, err = svc.CreateOrg(ctx, "BobCo", nil, bob)
	if err != nil {
		t.Fatalf("bob org: %v", err)
	}

	aliceList, err := svc.ListOrgsForUser(ctx, alice)
	if err != nil {
		t.Fatalf("ListOrgsForUser: %v", err)
	}
	if len(aliceList) != 1 {
		t.Fatalf("want 1 org for alice, got %d", len(aliceList))
	}
	if aliceList[0].Id != aliceOrg.Id {
		t.Errorf("returned %s, want %s", aliceList[0].Id, aliceOrg.Id)
	}
}

func TestGetOrgForUser_NonMemberReturnsNotFound(t *testing.T) {
	svc, db := newOrgService(t)
	ctx := context.Background()

	owner := seedUser(t, db, "user_owner_get", "og@example.com")
	outsider := seedUser(t, db, "user_outsider", "os@example.com")

	o, err := svc.CreateOrg(ctx, "Org", nil, owner)
	if err != nil {
		t.Fatalf("CreateOrg: %v", err)
	}
	_, err = svc.GetOrgForUser(ctx, o.Id, outsider)
	if !errors.Is(err, org.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestUpdateOrg_PatchesName(t *testing.T) {
	svc, db := newOrgService(t)
	ctx := context.Background()

	owner := seedUser(t, db, "user_updater", "up@example.com")
	o, err := svc.CreateOrg(ctx, "Before", nil, owner)
	if err != nil {
		t.Fatalf("CreateOrg: %v", err)
	}

	newName := "After"
	updated, err := svc.UpdateOrg(ctx, o.Id, &newName, owner)
	if err != nil {
		t.Fatalf("UpdateOrg: %v", err)
	}
	if updated.Name != "After" {
		t.Errorf("name: want After, got %q", updated.Name)
	}
}

func TestAddMember_UnknownEmailReturnsNotFound(t *testing.T) {
	svc, db := newOrgService(t)
	ctx := context.Background()

	owner := seedUser(t, db, "user_am_owner", "amo@example.com")
	o, err := svc.CreateOrg(ctx, "Org", nil, owner)
	if err != nil {
		t.Fatalf("CreateOrg: %v", err)
	}

	_, err = svc.AddMember(ctx, o.Id, "ghost@example.com", oapi.Analyst)
	if !errors.Is(err, org.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestAddMember_DuplicateReturnsConflict(t *testing.T) {
	svc, db := newOrgService(t)
	ctx := context.Background()

	owner := seedUser(t, db, "user_am_dup_owner", "amd@example.com")
	_ = seedUser(t, db, "user_am_dup_other", "other@example.com")

	o, err := svc.CreateOrg(ctx, "Org", nil, owner)
	if err != nil {
		t.Fatalf("CreateOrg: %v", err)
	}

	if _, err := svc.AddMember(ctx, o.Id, "other@example.com", oapi.Analyst); err != nil {
		t.Fatalf("first AddMember: %v", err)
	}
	_, err = svc.AddMember(ctx, o.Id, "other@example.com", oapi.Analyst)
	if !errors.Is(err, org.ErrConflict) {
		t.Fatalf("want ErrConflict on duplicate, got %v", err)
	}
}

func TestUpdateMemberRole_ChangesRole(t *testing.T) {
	svc, db := newOrgService(t)
	ctx := context.Background()

	owner := seedUser(t, db, "user_umr_owner", "umro@example.com")
	other := seedUser(t, db, "user_umr_other", "umru@example.com")

	o, err := svc.CreateOrg(ctx, "Org", nil, owner)
	if err != nil {
		t.Fatalf("CreateOrg: %v", err)
	}
	if _, err := svc.AddMember(ctx, o.Id, "umru@example.com", oapi.Analyst); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	updated, err := svc.UpdateMemberRole(ctx, o.Id, other, oapi.Admin)
	if err != nil {
		t.Fatalf("UpdateMemberRole: %v", err)
	}
	if updated.Role != oapi.Admin {
		t.Errorf("role: want admin, got %q", updated.Role)
	}
}

func TestRemoveMember_ExpectsErrNotFoundWhenMissing(t *testing.T) {
	svc, db := newOrgService(t)
	ctx := context.Background()

	owner := seedUser(t, db, "user_rm_owner", "rmo@example.com")
	o, err := svc.CreateOrg(ctx, "Org", nil, owner)
	if err != nil {
		t.Fatalf("CreateOrg: %v", err)
	}

	missing := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	err = svc.RemoveMember(ctx, o.Id, missing)
	if !errors.Is(err, org.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestListMembers_IncludesEmbeddedUser(t *testing.T) {
	svc, db := newOrgService(t)
	ctx := context.Background()

	owner := seedUser(t, db, "user_lm_owner", "lmo@example.com")
	o, err := svc.CreateOrg(ctx, "Org", nil, owner)
	if err != nil {
		t.Fatalf("CreateOrg: %v", err)
	}

	members, err := svc.ListMembers(ctx, o.Id)
	if err != nil {
		t.Fatalf("ListMembers: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("want 1 member, got %d", len(members))
	}
	if members[0].User == nil {
		t.Fatal("expected embedded user")
	}
	if members[0].User.Email != "lmo@example.com" {
		t.Errorf("user email: want lmo@example.com, got %q", members[0].User.Email)
	}
}
