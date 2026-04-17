package web

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/platform/authz"
	"github.com/luketeo/horizon/internal/platform/httpx"
	"github.com/luketeo/horizon/internal/platform/middleware"
	"github.com/luketeo/horizon/internal/services/orgservice"
)

// ── Helpers ──────────────────────────────────────────────────────────────────

// requireUser returns the internal PG user UUID for the authenticated Clerk user,
// upserting the user record on first access.
func (h *Handler) requireUser(ctx context.Context) (uuid.UUID, bool) {
	clerkUser, ok := middleware.GetClerkUserFromContext(ctx)
	if !ok {
		return uuid.Nil, false
	}
	_, userID, err := h.userSvc.GetOrCreateUser(ctx, clerkUser)
	if err != nil {
		return uuid.Nil, false
	}
	return userID, true
}

// requireMembership returns (userID, role) for the authenticated user in orgID.
// Returns (uuid.Nil, "", false) if the user is not authenticated or not a member.
func (h *Handler) requireMembership(
	ctx context.Context,
	orgID uuid.UUID,
) (uuid.UUID, oapi.OrgRole, bool) {
	userID, ok := h.requireUser(ctx)
	if !ok {
		return uuid.Nil, "", false
	}
	role, err := h.orgSvc.GetMembership(ctx, orgID, userID)
	if err != nil {
		return uuid.Nil, "", false
	}
	return userID, role, true
}

// ── Organizations ─────────────────────────────────────────────────────────────

func (h *Handler) ListOrganizations(
	ctx context.Context,
	_ oapi.ListOrganizationsRequestObject,
) (oapi.ListOrganizationsResponseObject, error) {
	userID, ok := h.requireUser(ctx)
	if !ok {
		return oapi.ListOrganizations401ApplicationProblemPlusJSONResponse{
			UnauthorizedApplicationProblemPlusJSONResponse: oapi.UnauthorizedApplicationProblemPlusJSONResponse(
				httpx.Prob(401, "Unauthorized", "Authentication required"),
			),
		}, nil
	}

	orgs, err := h.orgSvc.ListOrgsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if orgs == nil {
		orgs = []oapi.Organization{}
	}
	return oapi.ListOrganizations200JSONResponse(orgs), nil
}

func (h *Handler) CreateOrganization(
	ctx context.Context,
	request oapi.CreateOrganizationRequestObject,
) (oapi.CreateOrganizationResponseObject, error) {
	userID, ok := h.requireUser(ctx)
	if !ok {
		return oapi.CreateOrganization401ApplicationProblemPlusJSONResponse{
			UnauthorizedApplicationProblemPlusJSONResponse: oapi.UnauthorizedApplicationProblemPlusJSONResponse(
				httpx.Prob(401, "Unauthorized", "Authentication required"),
			),
		}, nil
	}

	org, err := h.orgSvc.CreateOrg(ctx, request.Body.Name, request.Body.Slug, userID)
	if err != nil {
		if errors.Is(err, orgservice.ErrConflict) {
			return oapi.CreateOrganization409ApplicationProblemPlusJSONResponse{
				ConflictApplicationProblemPlusJSONResponse: oapi.ConflictApplicationProblemPlusJSONResponse(
					httpx.Prob(409, "Conflict", "An organisation with that slug already exists"),
				),
			}, nil
		}
		return nil, err
	}

	return oapi.CreateOrganization201JSONResponse(org), nil
}

func (h *Handler) GetOrganization(
	ctx context.Context,
	request oapi.GetOrganizationRequestObject,
) (oapi.GetOrganizationResponseObject, error) {
	userID, role, ok := h.requireMembership(ctx, request.OrgId)
	if !ok {
		// Could be unauthenticated or not a member — return 403 for both to avoid enumeration.
		return oapi.GetOrganization403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "You are not a member of this organisation"),
			),
		}, nil
	}
	_ = role // membership is sufficient to view

	org, err := h.orgSvc.GetOrgForUser(ctx, request.OrgId, userID)
	if err != nil {
		if errors.Is(err, orgservice.ErrNotFound) {
			return oapi.GetOrganization404ApplicationProblemPlusJSONResponse{
				NotFoundApplicationProblemPlusJSONResponse: oapi.NotFoundApplicationProblemPlusJSONResponse(
					httpx.Prob(404, "Not Found", "Organisation not found"),
				),
			}, nil
		}
		return nil, err
	}

	return oapi.GetOrganization200JSONResponse(org), nil
}

func (h *Handler) UpdateOrganization(
	ctx context.Context,
	request oapi.UpdateOrganizationRequestObject,
) (oapi.UpdateOrganizationResponseObject, error) {
	userID, role, ok := h.requireMembership(ctx, request.OrgId)
	if !ok {
		return oapi.UpdateOrganization403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "You are not a member of this organisation"),
			),
		}, nil
	}
	if !authz.HasRole(role, oapi.Admin) {
		return oapi.UpdateOrganization403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "Admin or owner role required"),
			),
		}, nil
	}

	org, err := h.orgSvc.UpdateOrg(ctx, request.OrgId, request.Body.Name, userID)
	if err != nil {
		if errors.Is(err, orgservice.ErrNotFound) {
			return oapi.UpdateOrganization404ApplicationProblemPlusJSONResponse{
				NotFoundApplicationProblemPlusJSONResponse: oapi.NotFoundApplicationProblemPlusJSONResponse(
					httpx.Prob(404, "Not Found", "Organisation not found"),
				),
			}, nil
		}
		return nil, err
	}

	return oapi.UpdateOrganization200JSONResponse(org), nil
}

// ── Members ───────────────────────────────────────────────────────────────────

func (h *Handler) ListOrganizationMembers(
	ctx context.Context,
	request oapi.ListOrganizationMembersRequestObject,
) (oapi.ListOrganizationMembersResponseObject, error) {
	_, _, ok := h.requireMembership(ctx, request.OrgId)
	if !ok {
		return oapi.ListOrganizationMembers403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "You are not a member of this organisation"),
			),
		}, nil
	}

	members, err := h.orgSvc.ListMembers(ctx, request.OrgId)
	if err != nil {
		return nil, err
	}
	if members == nil {
		members = []oapi.OrganizationMember{}
	}
	return oapi.ListOrganizationMembers200JSONResponse(members), nil
}

func (h *Handler) AddOrganizationMember(
	ctx context.Context,
	request oapi.AddOrganizationMemberRequestObject,
) (oapi.AddOrganizationMemberResponseObject, error) {
	_, role, ok := h.requireMembership(ctx, request.OrgId)
	if !ok {
		return oapi.AddOrganizationMember403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "You are not a member of this organisation"),
			),
		}, nil
	}
	if !authz.HasRole(role, oapi.Admin) {
		return oapi.AddOrganizationMember403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "Admin or owner role required"),
			),
		}, nil
	}

	member, err := h.orgSvc.AddMember(
		ctx,
		request.OrgId,
		string(request.Body.Email),
		request.Body.Role,
	)
	if err != nil {
		switch {
		case errors.Is(err, orgservice.ErrNotFound):
			return oapi.AddOrganizationMember404ApplicationProblemPlusJSONResponse{
				NotFoundApplicationProblemPlusJSONResponse: oapi.NotFoundApplicationProblemPlusJSONResponse(
					httpx.Prob(
						404,
						"Not Found",
						"No user with that email address exists in Horizon",
					),
				),
			}, nil
		case errors.Is(err, orgservice.ErrConflict):
			return oapi.AddOrganizationMember409ApplicationProblemPlusJSONResponse{
				ConflictApplicationProblemPlusJSONResponse: oapi.ConflictApplicationProblemPlusJSONResponse(
					httpx.Prob(409, "Conflict", "User is already a member of this organisation"),
				),
			}, nil
		}
		return nil, err
	}

	return oapi.AddOrganizationMember201JSONResponse(member), nil
}

func (h *Handler) UpdateOrganizationMember(
	ctx context.Context,
	request oapi.UpdateOrganizationMemberRequestObject,
) (oapi.UpdateOrganizationMemberResponseObject, error) {
	_, role, ok := h.requireMembership(ctx, request.OrgId)
	if !ok {
		return oapi.UpdateOrganizationMember403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "You are not a member of this organisation"),
			),
		}, nil
	}
	if !authz.HasRole(role, oapi.Admin) {
		return oapi.UpdateOrganizationMember403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "Admin or owner role required"),
			),
		}, nil
	}

	member, err := h.orgSvc.UpdateMemberRole(ctx, request.OrgId, request.UserId, request.Body.Role)
	if err != nil {
		if errors.Is(err, orgservice.ErrNotFound) {
			return oapi.UpdateOrganizationMember404ApplicationProblemPlusJSONResponse{
				NotFoundApplicationProblemPlusJSONResponse: oapi.NotFoundApplicationProblemPlusJSONResponse(
					httpx.Prob(404, "Not Found", "Member not found"),
				),
			}, nil
		}
		return nil, err
	}

	return oapi.UpdateOrganizationMember200JSONResponse(member), nil
}

func (h *Handler) RemoveOrganizationMember(
	ctx context.Context,
	request oapi.RemoveOrganizationMemberRequestObject,
) (oapi.RemoveOrganizationMemberResponseObject, error) {
	_, role, ok := h.requireMembership(ctx, request.OrgId)
	if !ok {
		return oapi.RemoveOrganizationMember403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "You are not a member of this organisation"),
			),
		}, nil
	}
	if !authz.HasRole(role, oapi.Admin) {
		return oapi.RemoveOrganizationMember403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "Admin or owner role required"),
			),
		}, nil
	}

	if err := h.orgSvc.RemoveMember(ctx, request.OrgId, request.UserId); err != nil {
		if errors.Is(err, orgservice.ErrNotFound) {
			return oapi.RemoveOrganizationMember404ApplicationProblemPlusJSONResponse{
				NotFoundApplicationProblemPlusJSONResponse: oapi.NotFoundApplicationProblemPlusJSONResponse(
					httpx.Prob(404, "Not Found", "Member not found"),
				),
			}, nil
		}
		return nil, err
	}

	return oapi.RemoveOrganizationMember204Response{}, nil
}
