package web

import (
	"context"
	"errors"

	"github.com/luketeo/horizon/generated/oapi"
)

func (h *Handler) GetUsersMe(
	_ context.Context,
	_ oapi.GetUsersMeRequestObject,
) (oapi.GetUsersMeResponseObject, error) {
	return nil, errors.New("oopsie")
}
