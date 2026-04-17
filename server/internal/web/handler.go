package web

import (
	"context"

	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/apikey"
	"github.com/luketeo/horizon/internal/config"
	"github.com/luketeo/horizon/internal/services/orgservice"
	"github.com/luketeo/horizon/internal/user"
)

// Handler is the aggregate HTTP handler. Each carved-out domain lives under its
// own package with its own *Handler; the aggregator holds one pointer per
// domain and forwards oapi StrictServerInterface calls to them. Org/member
// endpoints are still declared directly on *Handler via web/org.go; they will
// move to an internal/org package in Phase 4.
type Handler struct {
	config  *config.Config
	orgSvc  *orgservice.Service
	userSvc *user.Service

	userH   *user.Handler
	apikeyH *apikey.Handler
}

// Compile-time guarantee that every oapi route has a concrete implementation.
var _ oapi.StrictServerInterface = (*Handler)(nil)

func NewHandler(cfg *config.Config) *Handler {
	db := cfg.DB()
	logger := cfg.Logger()

	userSvc := user.NewService(user.NewRepo(db), logger)
	orgSvc := orgservice.New(db)
	apikeySvc := apikey.NewService(apikey.NewRepo(db), logger)

	return &Handler{
		config:  cfg,
		orgSvc:  orgSvc,
		userSvc: userSvc,

		userH:   user.NewHandler(userSvc),
		apikeyH: apikey.NewHandler(apikeySvc, userSvc, orgSvc),
	}
}

// ── User endpoint forwarders ─────────────────────────────────────────────────

func (h *Handler) GetUsersMe(
	ctx context.Context,
	req oapi.GetUsersMeRequestObject,
) (oapi.GetUsersMeResponseObject, error) {
	return h.userH.GetUsersMe(ctx, req)
}

func (h *Handler) UpdateUsersMe(
	ctx context.Context,
	req oapi.UpdateUsersMeRequestObject,
) (oapi.UpdateUsersMeResponseObject, error) {
	return h.userH.UpdateUsersMe(ctx, req)
}

// ── API-key endpoint forwarders ──────────────────────────────────────────────

func (h *Handler) ListApiKeys(
	ctx context.Context,
	req oapi.ListApiKeysRequestObject,
) (oapi.ListApiKeysResponseObject, error) {
	return h.apikeyH.ListApiKeys(ctx, req)
}

func (h *Handler) CreateApiKey(
	ctx context.Context,
	req oapi.CreateApiKeyRequestObject,
) (oapi.CreateApiKeyResponseObject, error) {
	return h.apikeyH.CreateApiKey(ctx, req)
}

func (h *Handler) RevokeApiKey(
	ctx context.Context,
	req oapi.RevokeApiKeyRequestObject,
) (oapi.RevokeApiKeyResponseObject, error) {
	return h.apikeyH.RevokeApiKey(ctx, req)
}
