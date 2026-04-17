package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/luketeo/horizon/internal/config"
)

type contextKey string

const (
	clerkAuthUserKey contextKey = "clerk_auth_user"
)

// NewClerkAuthMiddleware attaches the authenticated Clerk user into the request context.
// If authentication fails, it returns 401 Unauthorized and blocks the request.
func NewClerkAuthMiddleware(_ *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Use Clerk middleware to verify Authorization header and extract session claims
			handler := clerkhttp.WithHeaderAuthorization()(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					claims, ok := clerk.SessionClaimsFromContext(r.Context())
					if !ok {
						http.Error(
							w,
							"Unauthorized: missing or invalid session claims",
							http.StatusUnauthorized,
						)
						return
					}

					authUser, err := user.Get(r.Context(), claims.Subject)
					if err != nil {
						slog.Default().
							ErrorContext(r.Context(), "failed to get user from Clerk", slog.String("claims_subject", claims.Subject), slog.Any("clerk_err", err))
						http.Error(
							w,
							"Unauthorized: failed to get authenticated user",
							http.StatusUnauthorized,
						)
						return
					}

					ctx := context.WithValue(r.Context(), clerkAuthUserKey, authUser)
					r = r.WithContext(ctx)

					next.ServeHTTP(w, r)
				}),
			)
			handler.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// GetClerkUserFromContext retrieves the authenticated Clerk user from context.
func GetClerkUserFromContext(ctx context.Context) (*clerk.User, bool) {
	authUser, ok := ctx.Value(clerkAuthUserKey).(*clerk.User)
	return authUser, ok
}
