package orgservice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

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
