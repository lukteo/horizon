package apikey

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/platform/authz"
	"github.com/luketeo/horizon/internal/platform/httpx"
	"github.com/luketeo/horizon/internal/platform/middleware"
	"github.com/luketeo/horizon/internal/services/orgservice"
	"github.com/luketeo/horizon/internal/user"
)

// Handler serves the /organizations/{orgId}/api-keys/* endpoints. It reaches
// into user.Service for identity and orgservice.Service for membership role
// checks — the two cross-domain dependencies handlers are allowed to take.
type Handler struct {
	svc     *Service
	userSvc *user.Service
	orgSvc  *orgservice.Service
}

// NewHandler wires a Handler with the services it needs.
func NewHandler(svc *Service, userSvc *user.Service, orgSvc *orgservice.Service) *Handler {
	return &Handler{svc: svc, userSvc: userSvc, orgSvc: orgSvc}
}

// requireMembership resolves identity and org membership in one step.
func (h *Handler) requireMembership(
	ctx context.Context,
	orgID uuid.UUID,
) (oapi.OrgRole, bool) {
	clerkUser, ok := middleware.GetClerkUserFromContext(ctx)
	if !ok {
		return "", false
	}
	_, userID, err := h.userSvc.GetOrCreateUser(ctx, clerkUser)
	if err != nil {
		return "", false
	}
	role, err := h.orgSvc.GetMembership(ctx, orgID, userID)
	if err != nil {
		return "", false
	}
	return role, true
}

func (h *Handler) ListApiKeys(
	ctx context.Context,
	request oapi.ListApiKeysRequestObject,
) (oapi.ListApiKeysResponseObject, error) {
	role, ok := h.requireMembership(ctx, request.OrgId)
	if !ok {
		return oapi.ListApiKeys403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "You are not a member of this organisation"),
			),
		}, nil
	}
	if !authz.HasRole(role, oapi.Admin) {
		return oapi.ListApiKeys403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "Admin or owner role required to view API keys"),
			),
		}, nil
	}

	keys, err := h.svc.List(ctx, request.OrgId)
	if err != nil {
		return nil, err
	}
	if keys == nil {
		keys = []oapi.ApiKey{}
	}
	return oapi.ListApiKeys200JSONResponse(keys), nil
}

func (h *Handler) CreateApiKey(
	ctx context.Context,
	request oapi.CreateApiKeyRequestObject,
) (oapi.CreateApiKeyResponseObject, error) {
	role, ok := h.requireMembership(ctx, request.OrgId)
	if !ok {
		return oapi.CreateApiKey403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "You are not a member of this organisation"),
			),
		}, nil
	}
	if !authz.HasRole(role, oapi.Admin) {
		return oapi.CreateApiKey403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "Admin or owner role required"),
			),
		}, nil
	}

	key, err := h.svc.Create(ctx, request.OrgId, request.Body.Name, request.Body.Scopes)
	if err != nil {
		return nil, err
	}
	return oapi.CreateApiKey201JSONResponse(key), nil
}

func (h *Handler) RevokeApiKey(
	ctx context.Context,
	request oapi.RevokeApiKeyRequestObject,
) (oapi.RevokeApiKeyResponseObject, error) {
	role, ok := h.requireMembership(ctx, request.OrgId)
	if !ok {
		return oapi.RevokeApiKey403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "You are not a member of this organisation"),
			),
		}, nil
	}
	if !authz.HasRole(role, oapi.Admin) {
		return oapi.RevokeApiKey403ApplicationProblemPlusJSONResponse{
			ForbiddenApplicationProblemPlusJSONResponse: oapi.ForbiddenApplicationProblemPlusJSONResponse(
				httpx.Prob(403, "Forbidden", "Admin or owner role required"),
			),
		}, nil
	}

	if err := h.svc.Revoke(ctx, request.OrgId, request.KeyId); err != nil {
		if errors.Is(err, ErrNotFound) {
			return oapi.RevokeApiKey404ApplicationProblemPlusJSONResponse{
				NotFoundApplicationProblemPlusJSONResponse: oapi.NotFoundApplicationProblemPlusJSONResponse(
					httpx.Prob(404, "Not Found", "API key not found"),
				),
			}, nil
		}
		return nil, err
	}
	return oapi.RevokeApiKey204Response{}, nil
}
