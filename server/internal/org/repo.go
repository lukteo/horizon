package org

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/luketeo/horizon/generated/oapi"
)

// Repo owns organizations + organization_members SQL and row→DTO mapping.
type Repo struct {
	db *sql.DB
}

// NewRepo wires a Repo backed by the given database handle.
func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// CreateOrgWithOwner inserts an organisation and its owner membership in one
// transaction, retrying on slug unique-violations with `<base>-2`, `<base>-3`, …
// Unique violations inside a tx normally abort the tx; we wrap each insert in a
// SAVEPOINT so we can roll back to a clean state and retry.
func (r *Repo) CreateOrgWithOwner(
	ctx context.Context,
	name, baseSlug string,
	creatorID uuid.UUID,
) (oapi.Organization, error) {
	tx, err := r.db.BeginTx(ctx, nil)
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

	slug := baseSlug
	for i := range 10 {
		if _, err = tx.ExecContext(ctx, "SAVEPOINT insert_org"); err != nil {
			return oapi.Organization{}, fmt.Errorf("savepoint: %w", err)
		}
		err = tx.QueryRowContext(ctx, insertOrg, name, slug).
			Scan(&orgID, &createdAt, &updatedAt)
		if err == nil {
			if _, relErr := tx.ExecContext(ctx, "RELEASE SAVEPOINT insert_org"); relErr != nil {
				return oapi.Organization{}, fmt.Errorf("release savepoint: %w", relErr)
			}
			break
		}
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if _, rbErr := tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT insert_org"); rbErr != nil {
				return oapi.Organization{}, fmt.Errorf("rollback to savepoint: %w", rbErr)
			}
			slug = fmt.Sprintf("%s-%d", baseSlug, i+2)
			continue
		}
		return oapi.Organization{}, fmt.Errorf("inserting org: %w", err)
	}
	if err != nil {
		return oapi.Organization{}, fmt.Errorf("inserting org (all retries exhausted): %w", err)
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO organization_members (org_id, user_id, role) VALUES ($1, $2, 'owner')`,
		orgID, creatorID,
	); err != nil {
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

// ListForUser returns orgs the user is a member of, with their role and the
// current member count embedded.
func (r *Repo) ListForUser(
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
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("listing orgs: %w", err)
	}
	defer rows.Close()

	var orgs []oapi.Organization
	for rows.Next() {
		var (
			o           oapi.Organization
			role        oapi.OrgRole
			memberCount int
		)
		if err := rows.Scan(
			&o.Id, &o.Name, &o.Slug, &o.Plan, &o.CreatedAt, &o.UpdatedAt,
			&role, &memberCount,
		); err != nil {
			return nil, fmt.Errorf("scanning org row: %w", err)
		}
		o.MyRole = &role
		o.MemberCount = &memberCount
		orgs = append(orgs, o)
	}
	return orgs, rows.Err()
}

// GetForUser returns the org together with the user's role. Returns ErrNotFound
// when the org does not exist or the user is not a member.
func (r *Repo) GetForUser(
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
		o           oapi.Organization
		role        oapi.OrgRole
		memberCount int
	)
	err := r.db.QueryRowContext(ctx, q, orgID, userID).Scan(
		&o.Id, &o.Name, &o.Slug, &o.Plan, &o.CreatedAt, &o.UpdatedAt,
		&role, &memberCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return oapi.Organization{}, ErrNotFound
	}
	if err != nil {
		return oapi.Organization{}, fmt.Errorf("getting org: %w", err)
	}
	o.MyRole = &role
	o.MemberCount = &memberCount
	return o, nil
}

// UpdateName patches the org's name. A nil name leaves the column unchanged.
func (r *Repo) UpdateName(ctx context.Context, orgID uuid.UUID, name *string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE organizations SET name = COALESCE($1, name), updated_at = NOW() WHERE id = $2`,
		name, orgID,
	)
	if err != nil {
		return fmt.Errorf("updating org: %w", err)
	}
	return nil
}

// GetMembership returns the user's role in the given org.
func (r *Repo) GetMembership(
	ctx context.Context,
	orgID, userID uuid.UUID,
) (oapi.OrgRole, error) {
	var role oapi.OrgRole
	err := r.db.QueryRowContext(ctx,
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

// ListMembers returns every member of the org with its embedded user row.
func (r *Repo) ListMembers(
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
	rows, err := r.db.QueryContext(ctx, q, orgID)
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

// FindUserIDByEmail resolves an email to its internal user id.
func (r *Repo) FindUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("looking up user by email: %w", err)
	}
	return id, nil
}

// InsertMember adds a user to an org at the given role. Returns ErrConflict on
// duplicate membership (unique violation).
func (r *Repo) InsertMember(
	ctx context.Context,
	orgID, userID uuid.UUID,
	role oapi.OrgRole,
) (oapi.OrganizationMember, error) {
	const q = `
		INSERT INTO organization_members (org_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, org_id, user_id, role, created_at, updated_at
	`
	var m oapi.OrganizationMember
	err := r.db.QueryRowContext(ctx, q, orgID, userID, role).
		Scan(&m.Id, &m.OrgId, &m.UserId, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return oapi.OrganizationMember{}, ErrConflict
		}
		return oapi.OrganizationMember{}, fmt.Errorf("inserting member: %w", err)
	}
	return m, nil
}

// UpdateMemberRole mutates an existing membership's role.
func (r *Repo) UpdateMemberRole(
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
	err := r.db.QueryRowContext(ctx, q, role, orgID, userID).
		Scan(&m.Id, &m.OrgId, &m.UserId, &m.Role, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return oapi.OrganizationMember{}, ErrNotFound
	}
	if err != nil {
		return oapi.OrganizationMember{}, fmt.Errorf("updating member role: %w", err)
	}
	return m, nil
}

// RemoveMember deletes a membership. Returns ErrNotFound when no row matched.
func (r *Repo) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM organization_members WHERE org_id = $1 AND user_id = $2`,
		orgID, userID,
	)
	if err != nil {
		return fmt.Errorf("removing member: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// GetUser loads a single user row for embedding into a member response.
func (r *Repo) GetUser(ctx context.Context, userID uuid.UUID) (oapi.User, error) {
	const q = `
		SELECT id, email, first_name, last_name, avatar_url, last_login_at, created_at, updated_at
		FROM users WHERE id = $1
	`
	var u oapi.User
	if err := r.db.QueryRowContext(ctx, q, userID).Scan(
		&u.Id, &u.Email, &u.FirstName, &u.LastName, &u.AvatarUrl, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return oapi.User{}, ErrNotFound
		}
		return oapi.User{}, fmt.Errorf("getting user: %w", err)
	}
	return u, nil
}
