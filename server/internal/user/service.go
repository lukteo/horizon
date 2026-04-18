// Package user owns the user domain: identity sync with Clerk and profile updates.
package user

import (
	"context"
	"errors"
	"log/slog"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/google/uuid"

	"github.com/luketeo/horizon/generated/oapi"
)

// ErrNotFound is returned when a user record does not exist.
var ErrNotFound = errors.New("user not found")

// Service coordinates user identity against Clerk and the local users table.
type Service struct {
	repo   *Repo
	logger *slog.Logger
}

// NewService wires a Service with its repo and structured logger.
func NewService(repo *Repo, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// primaryEmail selects the Clerk user's primary email, falling back to the first listed.
func primaryEmail(u *clerk.User) string {
	for _, ea := range u.EmailAddresses {
		if u.PrimaryEmailAddressID != nil && ea.ID == *u.PrimaryEmailAddressID {
			return ea.EmailAddress
		}
	}
	if len(u.EmailAddresses) > 0 {
		return u.EmailAddresses[0].EmailAddress
	}
	return ""
}

// GetOrCreateUser upserts a Clerk user into the local users table and returns the
// resulting record along with its internal UUID.
func (s *Service) GetOrCreateUser(
	ctx context.Context,
	clerkUser *clerk.User,
) (oapi.User, uuid.UUID, error) {
	email := primaryEmail(clerkUser)

	var imgURL *string
	if clerkUser.ImageURL != nil && *clerkUser.ImageURL != "" {
		imgURL = clerkUser.ImageURL
	}

	u, err := s.repo.Upsert(
		ctx,
		clerkUser.ID,
		email,
		clerkUser.FirstName,
		clerkUser.LastName,
		imgURL,
	)
	if err != nil {
		return oapi.User{}, uuid.Nil, err
	}
	return u, u.Id, nil
}

// UpdateUser updates mutable profile fields on the authenticated user.
func (s *Service) UpdateUser(
	ctx context.Context,
	userID uuid.UUID,
	firstName, lastName *string,
) (oapi.User, error) {
	return s.repo.Update(ctx, userID, firstName, lastName)
}

// GetUserIDByClerkID returns the internal UUID for a given Clerk id.
func (s *Service) GetUserIDByClerkID(ctx context.Context, clerkID string) (uuid.UUID, error) {
	return s.repo.GetIDByClerkID(ctx, clerkID)
}
