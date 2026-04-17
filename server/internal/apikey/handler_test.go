package apikey_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/apikey"
	"github.com/luketeo/horizon/internal/platform/middleware"
	"github.com/luketeo/horizon/internal/platform/testhelper"
	"github.com/luketeo/horizon/internal/services/orgservice"
	"github.com/luketeo/horizon/internal/user"
)

func newSeededHandler(t *testing.T) (*apikey.Handler, oapi.Organization) {
	t.Helper()
	db := testhelper.DB(t)
	testhelper.Reset(t, db)

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	userSvc := user.NewService(user.NewRepo(db), logger)
	orgSvc := orgservice.New(db)
	_, userID, err := userSvc.GetOrCreateUser(
		ctx,
		fakeClerkUser("user_apikey_handler_owner", "ownr@example.com"),
	)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	org, err := orgSvc.CreateOrg(ctx, "Apikey Handler Org", nil, userID)
	if err != nil {
		t.Fatalf("seed org: %v", err)
	}

	svc := apikey.NewService(apikey.NewRepo(db), logger)
	return apikey.NewHandler(svc, userSvc, orgSvc), org
}

func TestListApiKeys_UnauthenticatedReturnsForbidden(t *testing.T) {
	h, org := newSeededHandler(t)

	resp, err := h.ListApiKeys(
		context.Background(),
		oapi.ListApiKeysRequestObject{OrgId: org.Id},
	)
	if err != nil {
		t.Fatalf("ListApiKeys: %v", err)
	}
	if _, ok := resp.(oapi.ListApiKeys403ApplicationProblemPlusJSONResponse); !ok {
		t.Fatalf("want 403, got %T", resp)
	}
}

func TestCreateAndListApiKeys_AdminHappyPath(t *testing.T) {
	h, org := newSeededHandler(t)
	ctx := middleware.WithClerkUser(
		context.Background(),
		fakeClerkUser("user_apikey_handler_owner", "ownr@example.com"),
	)

	createResp, err := h.CreateApiKey(ctx, oapi.CreateApiKeyRequestObject{
		OrgId: org.Id,
		Body: &oapi.CreateApiKeyJSONRequestBody{
			Name:   "test-key",
			Scopes: []string{"read"},
		},
	})
	if err != nil {
		t.Fatalf("CreateApiKey: %v", err)
	}
	created, ok := createResp.(oapi.CreateApiKey201JSONResponse)
	if !ok {
		t.Fatalf("want 201 response, got %T", createResp)
	}
	if created.Key == "" {
		t.Error("raw key: want set, got empty")
	}

	listResp, err := h.ListApiKeys(ctx, oapi.ListApiKeysRequestObject{OrgId: org.Id})
	if err != nil {
		t.Fatalf("ListApiKeys: %v", err)
	}
	list, ok := listResp.(oapi.ListApiKeys200JSONResponse)
	if !ok {
		t.Fatalf("want 200, got %T", listResp)
	}
	if len(list) != 1 {
		t.Fatalf("want 1 key, got %d", len(list))
	}
	if list[0].Id != created.Id {
		t.Errorf("list returned %s, want %s", list[0].Id, created.Id)
	}
}

func TestRevokeApiKey_UnknownReturnsNotFound(t *testing.T) {
	h, org := newSeededHandler(t)
	ctx := middleware.WithClerkUser(
		context.Background(),
		fakeClerkUser("user_apikey_handler_owner", "ownr@example.com"),
	)

	missing := mustParseUUID(t, "00000000-0000-0000-0000-000000000099")
	resp, err := h.RevokeApiKey(ctx, oapi.RevokeApiKeyRequestObject{
		OrgId: org.Id,
		KeyId: missing,
	})
	if err != nil {
		t.Fatalf("RevokeApiKey: %v", err)
	}
	if _, ok := resp.(oapi.RevokeApiKey404ApplicationProblemPlusJSONResponse); !ok {
		t.Fatalf("want 404, got %T", resp)
	}
}

func mustParseUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", s, err)
	}
	return id
}
