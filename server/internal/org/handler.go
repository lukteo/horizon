package org

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/platform/authz"
	"github.com/luketeo/horizon/internal/platform/httpx"
	"github.com/luketeo/horizon/internal/platform/middleware"
	"github.com/luketeo/horizon/internal/user"
)

// Handler serves /orgs/* and /orgs/{id}/members/*. It depends on its own
// Service and on user.Service for identity resolution (Clerk → internal UUID).
type Handler struct {
	svc     *Service
	userSvc *user.Service
}

// NewHandler wires a Handler with the services it needs.
func NewHandler(svc *Service, userSvc *user.Service) *Handler {
	return &Handler{svc: svc, userSvc: userSvc}
}

// requireUser resolves the authenticated Clerk user to an internal user UUID,
// upserting the record on first access.
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

// requireMembership returns (userID, role, true) for an authenticated member of
// orgID, or zero values and false for unauthenticated/non-member callers.
func (h *Handler) requireMembership(
	ctx context.Context,
	orgID uuid.UUID,
) (uuid.UUID, oapi.OrgRole, bool) {
	userID, ok := h.requireUser(ctx)
	if !ok {
		return uuid.Nil, "", false
	}
	role, err := h.svc.GetMembership(ctx, orgID, userID)
	if err != nil {
		return uuid.Nil, "", false
	}
	return userID, role, true
}

// ── Organizations ────────────────────────────────────────────────────────────

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

	orgs, err := h.svc.ListOrgsForUser(ctx, userID)
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

	o, err := h.svc.CreateOrg(ctx, request.Body.Name, request.Body.Slug, userID)
	if err != nil {
		if errors.Is(err, ErrConflict) {
			return oapi.CreateOrganization409ApplicationProblemPlusJSONResponse{
				ConflictApplicationProblemPlusJSONResponse: oapi.ConflictApplicationProblemPlusJSONResponse(
					httpx.Prob(409, "Conflict", "An organisation with that slug already exists"),
				),
			}, nil
		}
		return nil, err
	}
	return oapi.CreateOrganization201JSONResponse(o), nil
}

func (h *Handler) GetOrganization(
	ctx context.Context,
	request oapi.GetOrganizationRequestObject,
) (oapi.GetOrganizationResponseObject, error) {
	userID, _, ok := h.requireMembership(ctx, request.OrgId)
	if !ok {
		return oapi.GetOrganization403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "You are not a member of this organisation"),
			),
		}, nil
	}

	o, err := h.svc.GetOrgForUser(ctx, request.OrgId, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return oapi.GetOrganization404ApplicationProblemPlusJSONResponse{
				NotFoundApplicationProblemPlusJSONResponse: oapi.NotFoundApplicationProblemPlusJSONResponse(
					httpx.Prob(404, "Not Found", "Organisation not found"),
				),
			}, nil
		}
		return nil, err
	}
	return oapi.GetOrganization200JSONResponse(o), nil
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

	o, err := h.svc.UpdateOrg(ctx, request.OrgId, request.Body.Name, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return oapi.UpdateOrganization404ApplicationProblemPlusJSONResponse{
				NotFoundApplicationProblemPlusJSONResponse: oapi.NotFoundApplicationProblemPlusJSONResponse(
					httpx.Prob(404, "Not Found", "Organisation not found"),
				),
			}, nil
		}
		return nil, err
	}
	return oapi.UpdateOrganization200JSONResponse(o), nil
}

// ── Members ──────────────────────────────────────────────────────────────────

func (h *Handler) ListOrganizationMembers(
	ctx context.Context,
	request oapi.ListOrganizationMembersRequestObject,
) (oapi.ListOrganizationMembersResponseObject, error) {
	if _, _, ok := h.requireMembership(ctx, request.OrgId); !ok {
		return oapi.ListOrganizationMembers403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "You are not a member of this organisation"),
			),
		}, nil
	}

	members, err := h.svc.ListMembers(ctx, request.OrgId)
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

	m, err := h.svc.AddMember(ctx, request.OrgId, string(request.Body.Email), request.Body.Role)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return oapi.AddOrganizationMember404ApplicationProblemPlusJSONResponse{
				NotFoundApplicationProblemPlusJSONResponse: oapi.NotFoundApplicationProblemPlusJSONResponse(
					httpx.Prob(404, "Not Found", "No user with that email address exists in Horizon"),
				),
			}, nil
		case errors.Is(err, ErrConflict):
			return oapi.AddOrganizationMember409ApplicationProblemPlusJSONResponse{
				ConflictApplicationProblemPlusJSONResponse: oapi.ConflictApplicationProblemPlusJSONResponse(
					httpx.Prob(409, "Conflict", "User is already a member of this organisation"),
				),
			}, nil
		}
		return nil, err
	}
	return oapi.AddOrganizationMember201JSONResponse(m), nil
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

	m, err := h.svc.UpdateMemberRole(ctx, request.OrgId, request.UserId, request.Body.Role)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return oapi.UpdateOrganizationMember404ApplicationProblemPlusJSONResponse{
				NotFoundApplicationProblemPlusJSONResponse: oapi.NotFoundApplicationProblemPlusJSONResponse(
					httpx.Prob(404, "Not Found", "Member not found"),
				),
			}, nil
		}
		return nil, err
	}
	return oapi.UpdateOrganizationMember200JSONResponse(m), nil
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

	if err := h.svc.RemoveMember(ctx, request.OrgId, request.UserId); err != nil {
		if errors.Is(err, ErrNotFound) {
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
