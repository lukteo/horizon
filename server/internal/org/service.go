// Package org owns the organisation + membership domain: creating orgs, listing
// them for a user, updating name, and adding/updating/removing members.
package org

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"github.com/luketeo/horizon/generated/oapi"
)

// ErrNotFound is returned when an org, membership, or user cannot be found.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned on duplicate membership.
var ErrConflict = errors.New("conflict")

var nonAlphaNumeric = regexp.MustCompile(`[^a-z0-9]+`)

// Service coordinates organisation + membership operations.
type Service struct {
	repo   *Repo
	logger *slog.Logger
}

// NewService wires a Service with its repo and logger.
func NewService(repo *Repo, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// slugify converts a display name into a URL-safe slug.
func slugify(name string) string {
	s := strings.ToLower(name)
	s = nonAlphaNumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 100 {
		s = s[:100]
	}
	if s == "" {
		s = "org"
	}
	return s
}

// CreateOrg creates a new organisation with the creator assigned as owner.
// When slugHint is non-nil and non-empty it is used verbatim as the slug base;
// otherwise the base is derived from the org name. Slug conflicts are retried
// with a numeric suffix.
func (s *Service) CreateOrg(
	ctx context.Context,
	name string,
	slugHint *string,
	creatorID uuid.UUID,
) (oapi.Organization, error) {
	base := slugify(name)
	if slugHint != nil && *slugHint != "" {
		base = *slugHint
	}
	return s.repo.CreateOrgWithOwner(ctx, name, base, creatorID)
}

// ListOrgsForUser returns every org the given user belongs to.
func (s *Service) ListOrgsForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]oapi.Organization, error) {
	return s.repo.ListForUser(ctx, userID)
}

// GetOrgForUser returns an org plus the user's role. Returns ErrNotFound when
// the user is not a member.
func (s *Service) GetOrgForUser(
	ctx context.Context,
	orgID, userID uuid.UUID,
) (oapi.Organization, error) {
	return s.repo.GetForUser(ctx, orgID, userID)
}

// UpdateOrg patches the org name, then reloads the membership view.
func (s *Service) UpdateOrg(
	ctx context.Context,
	orgID uuid.UUID,
	name *string,
	userID uuid.UUID,
) (oapi.Organization, error) {
	if err := s.repo.UpdateName(ctx, orgID, name); err != nil {
		return oapi.Organization{}, err
	}
	return s.repo.GetForUser(ctx, orgID, userID)
}

// GetMembership returns the user's role for the org.
func (s *Service) GetMembership(
	ctx context.Context,
	orgID, userID uuid.UUID,
) (oapi.OrgRole, error) {
	return s.repo.GetMembership(ctx, orgID, userID)
}

// ListMembers returns the membership list with each user embedded.
func (s *Service) ListMembers(
	ctx context.Context,
	orgID uuid.UUID,
) ([]oapi.OrganizationMember, error) {
	return s.repo.ListMembers(ctx, orgID)
}

// AddMember adds a user (looked up by email) to an org at the given role.
// Returns ErrNotFound if the email isn't registered; ErrConflict on duplicate.
func (s *Service) AddMember(
	ctx context.Context,
	orgID uuid.UUID,
	email string,
	role oapi.OrgRole,
) (oapi.OrganizationMember, error) {
	targetUserID, err := s.repo.FindUserIDByEmail(ctx, email)
	if err != nil {
		return oapi.OrganizationMember{}, err
	}
	m, err := s.repo.InsertMember(ctx, orgID, targetUserID, role)
	if err != nil {
		return oapi.OrganizationMember{}, err
	}
	if u, err := s.repo.GetUser(ctx, targetUserID); err == nil {
		m.User = &u
	} else if !errors.Is(err, ErrNotFound) {
		s.logger.WarnContext(ctx, "failed to embed user on new member", "err", err)
	}
	return m, nil
}

// UpdateMemberRole mutates an existing membership's role.
func (s *Service) UpdateMemberRole(
	ctx context.Context,
	orgID, userID uuid.UUID,
	role oapi.OrgRole,
) (oapi.OrganizationMember, error) {
	m, err := s.repo.UpdateMemberRole(ctx, orgID, userID, role)
	if err != nil {
		return oapi.OrganizationMember{}, err
	}
	if u, err := s.repo.GetUser(ctx, userID); err == nil {
		m.User = &u
	} else if !errors.Is(err, ErrNotFound) {
		s.logger.WarnContext(ctx, "failed to embed user on updated member", "err", err)
	}
	return m, nil
}

// RemoveMember deletes a membership. Returns ErrNotFound when no row matched.
func (s *Service) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	return s.repo.RemoveMember(ctx, orgID, userID)
}
