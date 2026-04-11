package orgservice

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/luketeo/horizon/generated/oapi"
)

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned when a resource already exists or a state conflict occurs.
var ErrConflict = errors.New("conflict")

var nonAlphaNumeric = regexp.MustCompile(`[^a-z0-9]+`)

// Service handles all organisation, member, and API key operations.
type Service struct {
	db *sql.DB
}

// New creates a new Service with the provided database connection.
func New(db *sql.DB) *Service {
	return &Service{db: db}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

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

// generateAPIKey creates a random API key and returns the raw value and its SHA-256 hash.
func generateAPIKey() (rawKey, keyHash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generating random bytes: %w", err)
	}
	rawKey = "hrz_" + hex.EncodeToString(b)
	h := sha256.Sum256([]byte(rawKey))
	keyHash = hex.EncodeToString(h[:])
	return rawKey, keyHash, nil
}

// primaryEmail extracts the primary email address from a Clerk user.
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

// scanUser maps a row of user columns into an oapi.User.
// Expected column order: id, email, first_name, last_name, avatar_url, last_login_at, created_at, updated_at
func scanUser(dest *oapi.User, scanner interface {
	Scan(dest ...any) error
}) error {
	return scanner.Scan(
		&dest.Id, &dest.Email, &dest.FirstName, &dest.LastName,
		&dest.AvatarUrl, &dest.LastLoginAt, &dest.CreatedAt, &dest.UpdatedAt,
	)
}

// ── Users ────────────────────────────────────────────────────────────────────

// GetOrCreateUser upserts a Clerk user into the local users table and returns
// the internal record along with the internal UUID.
func (s *Service) GetOrCreateUser(
	ctx context.Context,
	clerkUser *clerk.User,
) (oapi.User, uuid.UUID, error) {
	email := primaryEmail(clerkUser)

	var imgURL *string
	if clerkUser.ImageURL != nil && *clerkUser.ImageURL != "" {
		imgURL = clerkUser.ImageURL
	}

	const q = `
		INSERT INTO users (clerk_id, email, first_name, last_name, avatar_url, last_login_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (clerk_id) DO UPDATE
		SET email         = EXCLUDED.email,
		    first_name    = COALESCE(EXCLUDED.first_name, users.first_name),
		    last_name     = COALESCE(EXCLUDED.last_name,  users.last_name),
		    avatar_url    = COALESCE(EXCLUDED.avatar_url, users.avatar_url),
		    last_login_at = NOW(),
		    updated_at    = NOW()
		RETURNING id, email, first_name, last_name, avatar_url, last_login_at, created_at, updated_at
	`
	var u oapi.User
	if err := scanUser(&u, s.db.QueryRowContext(ctx, q,
		clerkUser.ID, email, clerkUser.FirstName, clerkUser.LastName, imgURL,
	)); err != nil {
		return oapi.User{}, uuid.Nil, fmt.Errorf("upserting user: %w", err)
	}
	return u, u.Id, nil
}

// UpdateUser updates mutable profile fields for the user.
func (s *Service) UpdateUser(
	ctx context.Context,
	userID uuid.UUID,
	firstName, lastName *string,
) (oapi.User, error) {
	const q = `
		UPDATE users SET
		    first_name = COALESCE($1, first_name),
		    last_name  = COALESCE($2, last_name),
		    updated_at = NOW()
		WHERE id = $3
		RETURNING id, email, first_name, last_name, avatar_url, last_login_at, created_at, updated_at
	`
	var u oapi.User
	if err := scanUser(&u, s.db.QueryRowContext(ctx, q, firstName, lastName, userID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return oapi.User{}, ErrNotFound
		}
		return oapi.User{}, fmt.Errorf("updating user: %w", err)
	}
	return u, nil
}

// GetUserIDByClerkID returns the internal PG UUID for a Clerk user ID.
func (s *Service) GetUserIDByClerkID(ctx context.Context, clerkID string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.db.QueryRowContext(ctx, `SELECT id FROM users WHERE clerk_id = $1`, clerkID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("looking up user by clerk_id: %w", err)
	}
	return id, nil
}

// ── Organizations ─────────────────────────────────────────────────────────────

// CreateOrg creates a new organisation and assigns the creator as owner.
// If slugHint is provided and valid it is used; otherwise one is derived from name.
// Slug conflicts automatically append a numeric suffix.
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

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return oapi.Organization{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	const insertOrg = `
		INSERT INTO organizations (name, slug, plan, settings)
		VALUES ($1, $2, 'free', '{}')
		RETURNING id, created_at, updated_at
	`

	var (
		orgID     uuid.UUID
		createdAt time.Time
		updatedAt time.Time
	)

	slug := base
	for i := range 10 {
		err = tx.QueryRowContext(ctx, insertOrg, name, slug).Scan(&orgID, &createdAt, &updatedAt)
		if err == nil {
			break
		}
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			slug = fmt.Sprintf("%s-%d", base, i+2)
			continue
		}
		return oapi.Organization{}, fmt.Errorf("inserting org: %w", err)
	}
	if err != nil {
		return oapi.Organization{}, fmt.Errorf("inserting org (all retries exhausted): %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO organization_members (org_id, user_id, role) VALUES ($1, $2, 'owner')`,
		orgID, creatorID,
	)
	if err != nil {
		return oapi.Organization{}, fmt.Errorf("adding creator as owner: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return oapi.Organization{}, fmt.Errorf("commit: %w", err)
	}

	role := oapi.Owner
	count := 1
	return oapi.Organization{
		Id:          orgID,
		Name:        name,
		Slug:        slug,
		Plan:        "free",
		MemberCount: &count,
		MyRole:      &role,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// ListOrgsForUser returns all organisations the user is a member of.
func (s *Service) ListOrgsForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]oapi.Organization, error) {
	const q = `
		SELECT o.id, o.name, o.slug, o.plan, o.created_at, o.updated_at,
		       om.role,
		       (SELECT COUNT(*)::int FROM organization_members m WHERE m.org_id = o.id) AS member_count
		FROM organizations o
		JOIN organization_members om ON o.id = om.org_id AND om.user_id = $1
		ORDER BY o.created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("listing orgs: %w", err)
	}
	defer rows.Close()

	var orgs []oapi.Organization
	for rows.Next() {
		var (
			org         oapi.Organization
			role        oapi.OrgRole
			memberCount int
		)
		if err := rows.Scan(
			&org.Id, &org.Name, &org.Slug, &org.Plan, &org.CreatedAt, &org.UpdatedAt,
			&role, &memberCount,
		); err != nil {
			return nil, fmt.Errorf("scanning org row: %w", err)
		}
		org.MyRole = &role
		org.MemberCount = &memberCount
		orgs = append(orgs, org)
	}
	return orgs, rows.Err()
}

// GetOrgForUser returns a single org with the user's role.
// Returns ErrNotFound if the org doesn't exist or the user is not a member.
func (s *Service) GetOrgForUser(
	ctx context.Context,
	orgID, userID uuid.UUID,
) (oapi.Organization, error) {
	const q = `
		SELECT o.id, o.name, o.slug, o.plan, o.created_at, o.updated_at,
		       om.role,
		       (SELECT COUNT(*)::int FROM organization_members m WHERE m.org_id = o.id) AS member_count
		FROM organizations o
		JOIN organization_members om ON o.id = om.org_id AND om.user_id = $2
		WHERE o.id = $1
	`
	var (
		org         oapi.Organization
		role        oapi.OrgRole
		memberCount int
	)
	err := s.db.QueryRowContext(ctx, q, orgID, userID).Scan(
		&org.Id, &org.Name, &org.Slug, &org.Plan, &org.CreatedAt, &org.UpdatedAt,
		&role, &memberCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return oapi.Organization{}, ErrNotFound
	}
	if err != nil {
		return oapi.Organization{}, fmt.Errorf("getting org: %w", err)
	}
	org.MyRole = &role
	org.MemberCount = &memberCount
	return org, nil
}

// UpdateOrg updates the organisation's name.
func (s *Service) UpdateOrg(
	ctx context.Context,
	orgID uuid.UUID,
	name *string,
	userID uuid.UUID,
) (oapi.Organization, error) {
	_, err := s.db.ExecContext(ctx, `
		UPDATE organizations SET name = COALESCE($1, name), updated_at = NOW() WHERE id = $2
	`, name, orgID)
	if err != nil {
		return oapi.Organization{}, fmt.Errorf("updating org: %w", err)
	}
	return s.GetOrgForUser(ctx, orgID, userID)
}

// ── Membership ───────────────────────────────────────────────────────────────

// GetMembership returns the user's role in the org, or ErrNotFound if not a member.
func (s *Service) GetMembership(
	ctx context.Context,
	orgID, userID uuid.UUID,
) (oapi.OrgRole, error) {
	var role oapi.OrgRole
	err := s.db.QueryRowContext(ctx,
		`SELECT role FROM organization_members WHERE org_id = $1 AND user_id = $2`,
		orgID, userID,
	).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("getting membership: %w", err)
	}
	return role, nil
}

// ListMembers returns all members of an org with embedded user info.
func (s *Service) ListMembers(
	ctx context.Context,
	orgID uuid.UUID,
) ([]oapi.OrganizationMember, error) {
	const q = `
		SELECT om.id, om.org_id, om.user_id, om.role, om.created_at, om.updated_at,
		       u.id, u.email, u.first_name, u.last_name, u.avatar_url, u.last_login_at, u.created_at, u.updated_at
		FROM organization_members om
		JOIN users u ON om.user_id = u.id
		WHERE om.org_id = $1
		ORDER BY om.created_at ASC
	`
	rows, err := s.db.QueryContext(ctx, q, orgID)
	if err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}
	defer rows.Close()

	var members []oapi.OrganizationMember
	for rows.Next() {
		var (
			m oapi.OrganizationMember
			u oapi.User
		)
		if err := rows.Scan(
			&m.Id, &m.OrgId, &m.UserId, &m.Role, &m.CreatedAt, &m.UpdatedAt,
			&u.Id, &u.Email, &u.FirstName, &u.LastName, &u.AvatarUrl, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning member row: %w", err)
		}
		m.User = &u
		members = append(members, m)
	}
	return members, rows.Err()
}

// AddMember adds a user (looked up by email) to an org.
// Returns ErrNotFound if no user with that email exists in the system.
// Returns ErrConflict if the user is already a member.
func (s *Service) AddMember(
	ctx context.Context,
	orgID uuid.UUID,
	email string,
	role oapi.OrgRole,
) (oapi.OrganizationMember, error) {
	var targetUserID uuid.UUID
	err := s.db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = $1`, email).
		Scan(&targetUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return oapi.OrganizationMember{}, ErrNotFound
	}
	if err != nil {
		return oapi.OrganizationMember{}, fmt.Errorf("looking up user by email: %w", err)
	}

	const insertQ = `
		INSERT INTO organization_members (org_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, org_id, user_id, role, created_at, updated_at
	`
	var m oapi.OrganizationMember
	err = s.db.QueryRowContext(ctx, insertQ, orgID, targetUserID, role).
		Scan(&m.Id, &m.OrgId, &m.UserId, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return oapi.OrganizationMember{}, ErrConflict
		}
		return oapi.OrganizationMember{}, fmt.Errorf("inserting member: %w", err)
	}

	var u oapi.User
	if err := s.db.QueryRowContext(ctx,
		`SELECT id, email, first_name, last_name, avatar_url, last_login_at, created_at, updated_at FROM users WHERE id = $1`,
		targetUserID,
	).Scan(&u.Id, &u.Email, &u.FirstName, &u.LastName, &u.AvatarUrl, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt); err == nil {
		m.User = &u
	}
	return m, nil
}

// UpdateMemberRole changes the role of an existing member.
func (s *Service) UpdateMemberRole(
	ctx context.Context,
	orgID, userID uuid.UUID,
	role oapi.OrgRole,
) (oapi.OrganizationMember, error) {
	const q = `
		UPDATE organization_members SET role = $1, updated_at = NOW()
		WHERE org_id = $2 AND user_id = $3
		RETURNING id, org_id, user_id, role, created_at, updated_at
	`
	var m oapi.OrganizationMember
	err := s.db.QueryRowContext(ctx, q, role, orgID, userID).
		Scan(&m.Id, &m.OrgId, &m.UserId, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return oapi.OrganizationMember{}, ErrNotFound
	}
	if err != nil {
		return oapi.OrganizationMember{}, fmt.Errorf("updating member role: %w", err)
	}

	var u oapi.User
	if err := s.db.QueryRowContext(ctx,
		`SELECT id, email, first_name, last_name, avatar_url, last_login_at, created_at, updated_at FROM users WHERE id = $1`,
		userID,
	).Scan(&u.Id, &u.Email, &u.FirstName, &u.LastName, &u.AvatarUrl, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt); err == nil {
		m.User = &u
	}
	return m, nil
}

// RemoveMember removes a user from an org. Returns ErrNotFound if the membership doesn't exist.
func (s *Service) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM organization_members WHERE org_id = $1 AND user_id = $2`,
		orgID, userID,
	)
	if err != nil {
		return fmt.Errorf("removing member: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ── API Keys ─────────────────────────────────────────────────────────────────

// CreateAPIKey generates a new API key for the org.
// The returned CreatedApiKey includes the raw key value, which is only returned once.
func (s *Service) CreateAPIKey(
	ctx context.Context,
	orgID uuid.UUID,
	name string,
	scopes []string,
) (oapi.CreatedApiKey, error) {
	rawKey, keyHash, err := generateAPIKey()
	if err != nil {
		return oapi.CreatedApiKey{}, fmt.Errorf("generating key: %w", err)
	}

	const q = `
		INSERT INTO api_keys (org_id, name, key_hash, scopes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, org_id, name, scopes, created_at, updated_at
	`
	var (
		k        oapi.CreatedApiKey
		dbScopes pq.StringArray
	)
	err = s.db.QueryRowContext(ctx, q, orgID, name, keyHash, pq.Array(scopes)).
		Scan(&k.Id, &k.OrgId, &k.Name, &dbScopes, &k.CreatedAt, &k.UpdatedAt)
	if err != nil {
		return oapi.CreatedApiKey{}, fmt.Errorf("inserting api key: %w", err)
	}
	k.Scopes = []string(dbScopes)
	k.Key = rawKey
	return k, nil
}

// ListAPIKeys returns all active (non-revoked) API keys for the org.
func (s *Service) ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]oapi.ApiKey, error) {
	const q = `
		SELECT id, org_id, name, scopes, last_used_at, revoked_at, created_at, updated_at
		FROM api_keys
		WHERE org_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, q, orgID)
	if err != nil {
		return nil, fmt.Errorf("listing api keys: %w", err)
	}
	defer rows.Close()

	var keys []oapi.ApiKey
	for rows.Next() {
		var (
			k        oapi.ApiKey
			dbScopes pq.StringArray
		)
		if err := rows.Scan(&k.Id, &k.OrgId, &k.Name, &dbScopes,
			&k.LastUsedAt, &k.RevokedAt, &k.CreatedAt, &k.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning api key: %w", err)
		}
		k.Scopes = []string(dbScopes)
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// RevokeAPIKey soft-deletes an API key by setting its revoked_at timestamp.
// Returns ErrNotFound if no active key with that ID exists for the org.
func (s *Service) RevokeAPIKey(ctx context.Context, orgID, keyID uuid.UUID) error {
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE api_keys SET revoked_at = NOW() WHERE id = $1 AND org_id = $2 AND revoked_at IS NULL`,
		keyID,
		orgID,
	)
	if err != nil {
		return fmt.Errorf("revoking api key: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
