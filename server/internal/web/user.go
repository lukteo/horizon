package web

import (
	"context"

	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/platform/httpx"
	"github.com/luketeo/horizon/internal/platform/middleware"
)

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

	user, _, err := h.orgSvc.GetOrCreateUser(ctx, clerkUser)
	if err != nil {
		return nil, err
	}

	return oapi.GetUsersMe200JSONResponse(user), nil
}

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

	_, userID, err := h.orgSvc.GetOrCreateUser(ctx, clerkUser)
	if err != nil {
		return nil, err
	}

	user, err := h.orgSvc.UpdateUser(ctx, userID, request.Body.FirstName, request.Body.LastName)
	if err != nil {
		return nil, err
	}

	return oapi.UpdateUsersMe200JSONResponse(user), nil
}
