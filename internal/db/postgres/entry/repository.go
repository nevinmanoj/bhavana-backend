package entry

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/entry"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

type eventRepository struct {
}

func NewEntryWriteRepository() entry.EntryWriteRepository {
	return &eventRepository{}
}
func NewEntryReadRepository() entry.EntryReadRepository {
	return &eventRepository{}
}

// entries
func (e *eventRepository) GetAllEntries(ctx context.Context, db sqlx.ExtContext, filter entry.EntryFilter) ([]entry.EntryFull, error) {
	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	entries := []entry.EntryFull{}
	baseQuery := `SELECT
	t.id AS "entry.id",
	t.chest_number AS "entry.chest_number",
    t.school_id  AS "entry.school_id",
    t.event_id   AS "entry.event_id",
    t.created_at AS "entry.created_at",
	s.name AS school_name,
	s.address AS school_address,
	e.title AS event_title,
	e.category AS category
	FROM entries t
	JOIN schools s ON s.id = t.school_id
	JOIN events e ON e.id = t.event_id`
	args := []any{}
	conditions := []string{}
	if scope.UserID != nil && role == rbac.UserRoleJudge {
		baseQuery += " JOIN event_judges ej ON ej.event_id = e.id "
		conditions = append(conditions, "ej.user_id = ?")
		args = append(args, *scope.UserID)
	}
	if scope.UserID != nil && role == rbac.UserRoleSchoolAdmin {
		conditions = append(conditions, "s.school_admin = ?")
		args = append(args, *scope.UserID)
	}
	finalQuery, finalargs, err := buildEntryQuery(baseQuery, conditions, args, filter)
	if err != nil {
		return nil, entry.ErrInternal
	}
	err = sqlx.SelectContext(
		ctx, db,
		&entries,
		finalQuery, finalargs...,
	)
	if err != nil {
		return nil, entry.ErrInternal
	}
	return entries, nil
}
func (e *eventRepository) GetEntryByID(ctx context.Context, db sqlx.ExtContext, entryId int64) (*entry.EntryFull, error) {
	entries := []entry.EntryFull{}
	args := []any{entryId}
	baseQuery := `SELECT
	t.id AS "entry.id",
    t.school_id  AS "entry.school_id",
	t.chest_number AS "entry.chest_number",
    t.event_id   AS "entry.event_id",
    t.created_at AS "entry.created_at",
	s.name AS school_name,
	s.address AS school_address,
	e.title AS event_title,
	e.category AS category
	FROM entries t
	JOIN schools s ON s.id = t.school_id
	JOIN events e ON e.id = t.event_id
	WHERE t.id = $1`
	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	if scope.UserID != nil && role == rbac.UserRoleSchoolAdmin {
		baseQuery += ` AND s.school_admin = $2`
		args = append(args, *scope.UserID)
	}
	err := sqlx.SelectContext(
		ctx, db,
		&entries,
		baseQuery, args...,
	)
	if err != nil {
		return nil, entry.ErrInternal
	}
	if len(entries) == 0 {
		return nil, entry.ErrEntryNotFound
	}
	return &entries[0], nil
}
func (e *eventRepository) CreateEntry(ctx context.Context, db sqlx.ExtContext, entryToCreate *entry.Entry) error {
	query := `
		INSERT INTO entries (
			event_id,
			school_id
		)
		VALUES (
			:event_id,
			:school_id
		)
		RETURNING id,chest_number, created_at
	`

	rows, err := sqlx.NamedQueryContext(ctx, db, query, entryToCreate)
	if err != nil {
		return errorMapper(err)
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&entryToCreate.ID, &entryToCreate.ChestNumber, &entryToCreate.CreatedAt)
		return nil
	}

	return entry.ErrInternal
}
func (e *eventRepository) DeleteEntry(ctx context.Context, db sqlx.ExtContext, entryID int64) error {
	query := `DELETE FROM entries WHERE id = $1`
	result, err := db.ExecContext(ctx, query, entryID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entry.ErrEntryNotFound
	}
	return nil
}

// entry members
func (e *eventRepository) GetEntryMembers(ctx context.Context, db sqlx.ExtContext, entryId int64) ([]entry.EntryMember, error) {
	entryMembers := []entry.EntryMember{}
	baseQuery := `SELECT
	s.name,tm.entry_id,tm.student_id,tm.created_at
	FROM entry_members tm
	JOIN students s ON s.id = tm.student_id
	WHERE tm.entry_id = $1`
	err := sqlx.SelectContext(
		ctx, db,
		&entryMembers,
		baseQuery, entryId,
	)
	if err != nil {
		return nil, err
	}
	return entryMembers, nil
}
func (e *eventRepository) CreateEntryMember(ctx context.Context, db sqlx.ExtContext, entryMemberToCreate *entry.EntryMember) error {
	query := `
		INSERT INTO entry_members (
			entry_id,
			student_id
		)
		VALUES (
			:entry_id,
			:student_id
		)
		RETURNING created_at
	`

	rows, err := sqlx.NamedQueryContext(ctx, db, query, entryMemberToCreate)
	if err != nil {
		return errorMapper(err)
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&entryMemberToCreate.CreatedAt)
		return nil
	}

	return entry.ErrInternal
}
func (e *eventRepository) DeleteEntryMember(ctx context.Context, db sqlx.ExtContext, entryID int64, studentID int64) error {
	query := `DELETE FROM entry_members
	WHERE entry_id = $1
	AND student_id = $2`
	result, err := db.ExecContext(ctx, query, entryID, studentID)
	if err != nil {
		return errorMapper(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entry.ErrEntryMemberNotFound
	}
	return nil
}
