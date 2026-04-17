package org

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/lib/pq"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/luketeo/horizon/generated/horizon/public/model"
	"github.com/luketeo/horizon/generated/horizon/public/table"
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

// ── Scan targets & mappers ───────────────────────────────────────────────────

// orgWithMembership carries an org row plus the caller's membership role and
// the total member count. The embedded model.Organizations captures all
// organizations.* columns; MyRole and MemberCount come in via explicit aliases.
type orgWithMembership struct {
	model.Organizations
	MyRole      string
	MemberCount int32
}

func (o orgWithMembership) toOapi() oapi.Organization {
	role := oapi.OrgRole(o.MyRole)
	count := int(o.MemberCount)
	return oapi.Organization{
		Id:          o.ID,
		Name:        o.Name,
		Slug:        o.Slug,
		Plan:        o.Plan,
		MyRole:      &role,
		MemberCount: &count,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

// memberWithUser carries a membership row with its embedded user row.
type memberWithUser struct {
	model.OrganizationMembers
	User model.Users
}

func (m memberWithUser) toOapi() oapi.OrganizationMember {
	user := oapi.User{
		Id:          m.User.ID,
		Email:       openapi_types.Email(m.User.Email),
		FirstName:   m.User.FirstName,
		LastName:    m.User.LastName,
		AvatarUrl:   m.User.AvatarURL,
		LastLoginAt: m.User.LastLoginAt,
		CreatedAt:   m.User.CreatedAt,
		UpdatedAt:   m.User.UpdatedAt,
	}
	return oapi.OrganizationMember{
		Id:        m.ID,
		OrgId:     m.OrgID,
		UserId:    m.UserID,
		Role:      oapi.OrgRole(m.Role),
		User:      &user,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func userModelToOapi(u model.Users) oapi.User {
	return oapi.User{
		Id:          u.ID,
		Email:       openapi_types.Email(u.Email),
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		AvatarUrl:   u.AvatarURL,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// memberCountsSubquery builds a subquery selecting (org_id, member_count) for
// use as a LEFT JOIN on organizations.
func memberCountsSubquery() (postgres.SelectTable, postgres.ColumnString, postgres.ColumnInteger) {
	orgIDProj := table.OrganizationMembers.OrgID.AS("org_id")
	countProj := postgres.COUNT(postgres.STAR).AS("count")

	sub := postgres.
		SELECT(orgIDProj, countProj).
		FROM(table.OrganizationMembers).
		GROUP_BY(table.OrganizationMembers.OrgID).
		AsTable("mc")

	return sub,
		postgres.StringColumn("org_id").From(sub),
		postgres.IntegerColumn("count").From(sub)
}

// ── Organizations ────────────────────────────────────────────────────────────

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

	insertStmtFor := func(slug string) postgres.InsertStatement {
		return table.Organizations.
			INSERT(
				table.Organizations.Name,
				table.Organizations.Slug,
				table.Organizations.Plan,
				table.Organizations.Settings,
			).
			VALUES(name, slug, "free", "{}").
			RETURNING(
				table.Organizations.ID,
				table.Organizations.CreatedAt,
				table.Organizations.UpdatedAt,
			)
	}

	var inserted model.Organizations

	slug := baseSlug
	for i := range 10 {
		if _, err = tx.ExecContext(ctx, "SAVEPOINT insert_org"); err != nil {
			return oapi.Organization{}, fmt.Errorf("savepoint: %w", err)
		}
		err = insertStmtFor(slug).QueryContext(ctx, tx, &inserted)
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

	memberInsert := table.OrganizationMembers.
		INSERT(
			table.OrganizationMembers.OrgID,
			table.OrganizationMembers.UserID,
			table.OrganizationMembers.Role,
		).
		VALUES(inserted.ID, creatorID, string(oapi.Owner))

	if _, err = memberInsert.ExecContext(ctx, tx); err != nil {
		return oapi.Organization{}, fmt.Errorf("adding creator as owner: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return oapi.Organization{}, fmt.Errorf("commit: %w", err)
	}

	role := oapi.Owner
	count := 1
	return oapi.Organization{
		Id:          inserted.ID,
		Name:        name,
		Slug:        slug,
		Plan:        "free",
		MemberCount: &count,
		MyRole:      &role,
		CreatedAt:   inserted.CreatedAt,
		UpdatedAt:   inserted.UpdatedAt,
	}, nil
}

// ListForUser returns orgs the user is a member of, with their role and the
// current member count embedded.
func (r *Repo) ListForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]oapi.Organization, error) {
	mcSub, mcOrgID, mcCount := memberCountsSubquery()

	stmt := postgres.
		SELECT(
			table.Organizations.AllColumns,
			table.OrganizationMembers.Role.AS("my_role"),
			mcCount.AS("member_count"),
		).
		FROM(
			table.Organizations.
				INNER_JOIN(
					table.OrganizationMembers,
					table.Organizations.ID.EQ(table.OrganizationMembers.OrgID).
						AND(table.OrganizationMembers.UserID.EQ(postgres.UUID(userID))),
				).
				LEFT_JOIN(mcSub, mcOrgID.EQ(table.Organizations.ID)),
		).
		ORDER_BY(table.Organizations.CreatedAt.DESC())

	var rows []orgWithMembership
	if err := stmt.QueryContext(ctx, r.db, &rows); err != nil {
		return nil, fmt.Errorf("listing orgs: %w", err)
	}

	orgs := make([]oapi.Organization, 0, len(rows))
	for _, row := range rows {
		orgs = append(orgs, row.toOapi())
	}
	return orgs, nil
}

// GetForUser returns the org together with the user's role. Returns ErrNotFound
// when the org does not exist or the user is not a member.
func (r *Repo) GetForUser(
	ctx context.Context,
	orgID, userID uuid.UUID,
) (oapi.Organization, error) {
	mcSub, mcOrgID, mcCount := memberCountsSubquery()

	stmt := postgres.
		SELECT(
			table.Organizations.AllColumns,
			table.OrganizationMembers.Role.AS("my_role"),
			mcCount.AS("member_count"),
		).
		FROM(
			table.Organizations.
				INNER_JOIN(
					table.OrganizationMembers,
					table.Organizations.ID.EQ(table.OrganizationMembers.OrgID).
						AND(table.OrganizationMembers.UserID.EQ(postgres.UUID(userID))),
				).
				LEFT_JOIN(mcSub, mcOrgID.EQ(table.Organizations.ID)),
		).
		WHERE(table.Organizations.ID.EQ(postgres.UUID(orgID))).
		LIMIT(1)

	var row orgWithMembership
	if err := stmt.QueryContext(ctx, r.db, &row); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return oapi.Organization{}, ErrNotFound
		}
		return oapi.Organization{}, fmt.Errorf("getting org: %w", err)
	}
	return row.toOapi(), nil
}

// UpdateName patches the org's name. A nil name leaves the column unchanged.
func (r *Repo) UpdateName(ctx context.Context, orgID uuid.UUID, name *string) error {
	nameExpr := postgres.StringExpression(table.Organizations.Name)
	if name != nil {
		nameExpr = postgres.String(*name)
	}

	stmt := table.Organizations.
		UPDATE(table.Organizations.Name, table.Organizations.UpdatedAt).
		SET(nameExpr, postgres.NOW()).
		WHERE(table.Organizations.ID.EQ(postgres.UUID(orgID)))

	if _, err := stmt.ExecContext(ctx, r.db); err != nil {
		return fmt.Errorf("updating org: %w", err)
	}
	return nil
}

// GetMembership returns the user's role in the given org.
func (r *Repo) GetMembership(
	ctx context.Context,
	orgID, userID uuid.UUID,
) (oapi.OrgRole, error) {
	stmt := postgres.
		SELECT(table.OrganizationMembers.Role).
		FROM(table.OrganizationMembers).
		WHERE(
			table.OrganizationMembers.OrgID.EQ(postgres.UUID(orgID)).
				AND(table.OrganizationMembers.UserID.EQ(postgres.UUID(userID))),
		).
		LIMIT(1)

	var row model.OrganizationMembers
	if err := stmt.QueryContext(ctx, r.db, &row); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("getting membership: %w", err)
	}
	return oapi.OrgRole(row.Role), nil
}

// ListMembers returns every member of the org with its embedded user row.
func (r *Repo) ListMembers(
	ctx context.Context,
	orgID uuid.UUID,
) ([]oapi.OrganizationMember, error) {
	stmt := postgres.
		SELECT(
			table.OrganizationMembers.AllColumns,
			table.Users.AllColumns,
		).
		FROM(
			table.OrganizationMembers.
				INNER_JOIN(table.Users, table.Users.ID.EQ(table.OrganizationMembers.UserID)),
		).
		WHERE(table.OrganizationMembers.OrgID.EQ(postgres.UUID(orgID))).
		ORDER_BY(table.OrganizationMembers.CreatedAt.ASC())

	var rows []memberWithUser
	if err := stmt.QueryContext(ctx, r.db, &rows); err != nil {
		return nil, fmt.Errorf("listing members: %w", err)
	}

	members := make([]oapi.OrganizationMember, 0, len(rows))
	for _, row := range rows {
		members = append(members, row.toOapi())
	}
	return members, nil
}

// FindUserIDByEmail resolves an email to its internal user id.
func (r *Repo) FindUserIDByEmail(ctx context.Context, email string) (uuid.UUID, error) {
	stmt := postgres.
		SELECT(table.Users.ID).
		FROM(table.Users).
		WHERE(table.Users.Email.EQ(postgres.String(email))).
		LIMIT(1)

	var row model.Users
	if err := stmt.QueryContext(ctx, r.db, &row); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}
		return uuid.Nil, fmt.Errorf("looking up user by email: %w", err)
	}
	return row.ID, nil
}

// InsertMember adds a user to an org at the given role. Returns ErrConflict on
// duplicate membership (unique violation).
func (r *Repo) InsertMember(
	ctx context.Context,
	orgID, userID uuid.UUID,
	role oapi.OrgRole,
) (oapi.OrganizationMember, error) {
	stmt := table.OrganizationMembers.
		INSERT(
			table.OrganizationMembers.OrgID,
			table.OrganizationMembers.UserID,
			table.OrganizationMembers.Role,
		).
		VALUES(orgID, userID, string(role)).
		RETURNING(table.OrganizationMembers.AllColumns)

	var row model.OrganizationMembers
	if err := stmt.QueryContext(ctx, r.db, &row); err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return oapi.OrganizationMember{}, ErrConflict
		}
		return oapi.OrganizationMember{}, fmt.Errorf("inserting member: %w", err)
	}
	return oapi.OrganizationMember{
		Id:        row.ID,
		OrgId:     row.OrgID,
		UserId:    row.UserID,
		Role:      oapi.OrgRole(row.Role),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

// UpdateMemberRole mutates an existing membership's role.
func (r *Repo) UpdateMemberRole(
	ctx context.Context,
	orgID, userID uuid.UUID,
	role oapi.OrgRole,
) (oapi.OrganizationMember, error) {
	stmt := table.OrganizationMembers.
		UPDATE(table.OrganizationMembers.Role, table.OrganizationMembers.UpdatedAt).
		SET(postgres.String(string(role)), postgres.NOW()).
		WHERE(
			table.OrganizationMembers.OrgID.EQ(postgres.UUID(orgID)).
				AND(table.OrganizationMembers.UserID.EQ(postgres.UUID(userID))),
		).
		RETURNING(table.OrganizationMembers.AllColumns)

	var row model.OrganizationMembers
	if err := stmt.QueryContext(ctx, r.db, &row); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return oapi.OrganizationMember{}, ErrNotFound
		}
		return oapi.OrganizationMember{}, fmt.Errorf("updating member role: %w", err)
	}
	return oapi.OrganizationMember{
		Id:        row.ID,
		OrgId:     row.OrgID,
		UserId:    row.UserID,
		Role:      oapi.OrgRole(row.Role),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

// RemoveMember deletes a membership. Returns ErrNotFound when no row matched.
func (r *Repo) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	stmt := table.OrganizationMembers.
		DELETE().
		WHERE(
			table.OrganizationMembers.OrgID.EQ(postgres.UUID(orgID)).
				AND(table.OrganizationMembers.UserID.EQ(postgres.UUID(userID))),
		)

	res, err := stmt.ExecContext(ctx, r.db)
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
	stmt := postgres.
		SELECT(table.Users.AllColumns).
		FROM(table.Users).
		WHERE(table.Users.ID.EQ(postgres.UUID(userID))).
		LIMIT(1)

	var row model.Users
	if err := stmt.QueryContext(ctx, r.db, &row); err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return oapi.User{}, ErrNotFound
		}
		return oapi.User{}, fmt.Errorf("getting user: %w", err)
	}
	return userModelToOapi(row), nil
}
