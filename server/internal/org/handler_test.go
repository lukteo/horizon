package org_test

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/org"
	"github.com/luketeo/horizon/internal/platform/middleware"
	"github.com/luketeo/horizon/internal/platform/testhelper"
	"github.com/luketeo/horizon/internal/user"
)

func newOrgHandler(t *testing.T) (*org.Handler, *user.Service, *org.Service, *sql.DB) {
	t.Helper()
	db := testhelper.DB(t)
	testhelper.Reset(t, db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	userSvc := user.NewService(user.NewRepo(db), logger)
	orgSvc := org.NewService(org.NewRepo(db), logger)
	return org.NewHandler(orgSvc, userSvc), userSvc, orgSvc, db
}

func TestListOrganizations_Unauthenticated(t *testing.T) {
	h, _, _, _ := newOrgHandler(t)

	resp, err := h.ListOrganizations(
		context.Background(),
		oapi.ListOrganizationsRequestObject{},
	)
	if err != nil {
		t.Fatalf("ListOrganizations: %v", err)
	}
	if _, ok := resp.(oapi.ListOrganizations401ApplicationProblemPlusJSONResponse); !ok {
		t.Fatalf("want 401, got %T", resp)
	}
}

func TestCreateAndListOrganizations_HappyPath(t *testing.T) {
	h, _, _, _ := newOrgHandler(t)
	ctx := middleware.WithClerkUser(
		context.Background(),
		fakeClerkUser("user_org_handler_owner", "own@example.com"),
	)

	createResp, err := h.CreateOrganization(ctx, oapi.CreateOrganizationRequestObject{
		Body: &oapi.CreateOrganizationJSONRequestBody{Name: "Handler Org"},
	})
	if err != nil {
		t.Fatalf("CreateOrganization: %v", err)
	}
	created, ok := createResp.(oapi.CreateOrganization201JSONResponse)
	if !ok {
		t.Fatalf("want 201, got %T", createResp)
	}

	listResp, err := h.ListOrganizations(ctx, oapi.ListOrganizationsRequestObject{})
	if err != nil {
		t.Fatalf("ListOrganizations: %v", err)
	}
	list, ok := listResp.(oapi.ListOrganizations200JSONResponse)
	if !ok {
		t.Fatalf("want 200, got %T", listResp)
	}
	if len(list) != 1 || list[0].Id != created.Id {
		t.Fatalf("want [%s], got %+v", created.Id, list)
	}
}

func TestGetOrganization_NonMemberReturns403(t *testing.T) {
	h, userSvc, orgSvc, _ := newOrgHandler(t)
	ctx := context.Background()

	_, ownerID, err := userSvc.GetOrCreateUser(
		ctx,
		fakeClerkUser("user_og_owner", "ogo@example.com"),
	)
	if err != nil {
		t.Fatalf("seed owner: %v", err)
	}
	o, err := orgSvc.CreateOrg(ctx, "Gated", nil, ownerID)
	if err != nil {
		t.Fatalf("CreateOrg: %v", err)
	}

	outsiderCtx := middleware.WithClerkUser(
		context.Background(),
		fakeClerkUser("user_og_outsider", "ogos@example.com"),
	)
	resp, err := h.GetOrganization(outsiderCtx, oapi.GetOrganizationRequestObject{OrgId: o.Id})
	if err != nil {
		t.Fatalf("GetOrganization: %v", err)
	}
	if _, ok := resp.(oapi.GetOrganization403ApplicationProblemPlusJSONResponse); !ok {
		t.Fatalf("want 403, got %T", resp)
	}
}

func TestAddOrganizationMember_UnknownEmailReturns404(t *testing.T) {
	h, _, _, _ := newOrgHandler(t)
	ctx := middleware.WithClerkUser(
		context.Background(),
		fakeClerkUser("user_amh_owner", "amho@example.com"),
	)

	createResp, err := h.CreateOrganization(ctx, oapi.CreateOrganizationRequestObject{
		Body: &oapi.CreateOrganizationJSONRequestBody{Name: "Add Member Org"},
	})
	if err != nil {
		t.Fatalf("CreateOrganization: %v", err)
	}
	created := createResp.(oapi.CreateOrganization201JSONResponse)

	resp, err := h.AddOrganizationMember(ctx, oapi.AddOrganizationMemberRequestObject{
		OrgId: created.Id,
		Body: &oapi.AddOrganizationMemberJSONRequestBody{
			Email: openapi_types.Email("ghost@example.com"),
			Role:  oapi.Analyst,
		},
	})
	if err != nil {
		t.Fatalf("AddOrganizationMember: %v", err)
	}
	if _, ok := resp.(oapi.AddOrganizationMember404ApplicationProblemPlusJSONResponse); !ok {
		t.Fatalf("want 404, got %T", resp)
	}
}
