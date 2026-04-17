package user

import (
	"context"

	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/platform/httpx"
	"github.com/luketeo/horizon/internal/platform/middleware"
)

// Handler exposes the /users/me HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler wires a Handler with its user service.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Service returns the underlying service so other domains can resolve identity.
func (h *Handler) Service() *Service { return h.svc }

// GetUsersMe returns the authenticated user's profile, upserting on first access.
func (h *Handler) GetUsersMe(
	ctx context.Context,
	_ oapi.GetUsersMeRequestObject,
) (oapi.GetUsersMeResponseObject, error) {
	clerkUser, ok := middleware.GetClerkUserFromContext(ctx)
	if !ok {
		return oapi.GetUsersMe401ApplicationProblemPlusJSONResponse{
			UnauthorizedApplicationProblemPlusJSONResponse: oapi.UnauthorizedApplicationProblemPlusJSONResponse(
				httpx.Prob(401, "Unauthorized", "Missing authentication context"),
			),
		}, nil
	}

	u, _, err := h.svc.GetOrCreateUser(ctx, clerkUser)
	if err != nil {
		return nil, err
	}
	return oapi.GetUsersMe200JSONResponse(u), nil
}

// UpdateUsersMe updates the authenticated user's mutable profile fields.
func (h *Handler) UpdateUsersMe(
	ctx context.Context,
	request oapi.UpdateUsersMeRequestObject,
) (oapi.UpdateUsersMeResponseObject, error) {
	clerkUser, ok := middleware.GetClerkUserFromContext(ctx)
	if !ok {
		return oapi.UpdateUsersMe401ApplicationProblemPlusJSONResponse{
			UnauthorizedApplicationProblemPlusJSONResponse: oapi.UnauthorizedApplicationProblemPlusJSONResponse(
				httpx.Prob(401, "Unauthorized", "Missing authentication context"),
			),
		}, nil
	}

	_, userID, err := h.svc.GetOrCreateUser(ctx, clerkUser)
	if err != nil {
		return nil, err
	}

	u, err := h.svc.UpdateUser(ctx, userID, request.Body.FirstName, request.Body.LastName)
	if err != nil {
		return nil, err
	}
	return oapi.UpdateUsersMe200JSONResponse(u), nil
}
