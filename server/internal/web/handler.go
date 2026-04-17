// Package web aggregates the domain-owned HTTP handlers into a single value
// that satisfies oapi.StrictServerInterface. Each domain (user, org, apikey)
// owns its own handler under internal/<domain>; this type forwards to them.
package web

import (
	"context"

	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/apikey"
	"github.com/luketeo/horizon/internal/config"
	"github.com/luketeo/horizon/internal/org"
	"github.com/luketeo/horizon/internal/user"
)

// Handler is the aggregate HTTP handler. It holds one pointer per domain
// package and forwards each oapi StrictServerInterface method to the owning
// domain handler.
type Handler struct {
	config *config.Config

	userH   *user.Handler
	orgH    *org.Handler
	apikeyH *apikey.Handler
}

// Compile-time guarantee that every oapi route has a concrete implementation.
var _ oapi.StrictServerInterface = (*Handler)(nil)

// NewHandler wires up each domain package with its own repo, service, and
// handler, then composes them into a single aggregator.
func NewHandler(cfg *config.Config) *Handler {
	db := cfg.DB()
	logger := cfg.Logger()

	userSvc := user.NewService(user.NewRepo(db), logger)
	orgSvc := org.NewService(org.NewRepo(db), logger)
	apikeySvc := apikey.NewService(apikey.NewRepo(db), logger)

	return &Handler{
		config:  cfg,
		userH:   user.NewHandler(userSvc),
		orgH:    org.NewHandler(orgSvc, userSvc),
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

// ── Organisation endpoint forwarders ─────────────────────────────────────────

func (h *Handler) ListOrganizations(
	ctx context.Context,
	req oapi.ListOrganizationsRequestObject,
) (oapi.ListOrganizationsResponseObject, error) {
	return h.orgH.ListOrganizations(ctx, req)
}

func (h *Handler) CreateOrganization(
	ctx context.Context,
	req oapi.CreateOrganizationRequestObject,
) (oapi.CreateOrganizationResponseObject, error) {
	return h.orgH.CreateOrganization(ctx, req)
}

func (h *Handler) GetOrganization(
	ctx context.Context,
	req oapi.GetOrganizationRequestObject,
) (oapi.GetOrganizationResponseObject, error) {
	return h.orgH.GetOrganization(ctx, req)
}

func (h *Handler) UpdateOrganization(
	ctx context.Context,
	req oapi.UpdateOrganizationRequestObject,
) (oapi.UpdateOrganizationResponseObject, error) {
	return h.orgH.UpdateOrganization(ctx, req)
}

func (h *Handler) ListOrganizationMembers(
	ctx context.Context,
	req oapi.ListOrganizationMembersRequestObject,
) (oapi.ListOrganizationMembersResponseObject, error) {
	return h.orgH.ListOrganizationMembers(ctx, req)
}

func (h *Handler) AddOrganizationMember(
	ctx context.Context,
	req oapi.AddOrganizationMemberRequestObject,
) (oapi.AddOrganizationMemberResponseObject, error) {
	return h.orgH.AddOrganizationMember(ctx, req)
}

func (h *Handler) UpdateOrganizationMember(
	ctx context.Context,
	req oapi.UpdateOrganizationMemberRequestObject,
) (oapi.UpdateOrganizationMemberResponseObject, error) {
	return h.orgH.UpdateOrganizationMember(ctx, req)
}

func (h *Handler) RemoveOrganizationMember(
	ctx context.Context,
	req oapi.RemoveOrganizationMemberRequestObject,
) (oapi.RemoveOrganizationMemberResponseObject, error) {
	return h.orgH.RemoveOrganizationMember(ctx, req)
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
