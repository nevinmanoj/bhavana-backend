package access

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/access"
)

type accessRepository struct {
}

func NewAccessRepository() access.AccessRepository {
	return &accessRepository{}
}

func (r *accessRepository) HasStudentAccess(ctx context.Context, db sqlx.ExtContext, studentID, userID int64) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1
			FROM students s
			JOIN schools sc ON s.school_id = sc.id
			WHERE s.id = $1 AND sc.school_admin = $2
		)
	`

	args := []any{studentID, userID}
	return r.accessQueryExecutor(ctx, db, q, args)
}
func (r *accessRepository) HasSchoolAccess(ctx context.Context, db sqlx.ExtContext, schoolID, userID int64) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1
			FROM schools sc
			WHERE sc.id = $1 AND sc.school_admin = $2
		)
	`

	args := []any{schoolID, userID}
	return r.accessQueryExecutor(ctx, db, q, args)
}
func (r *accessRepository) HasScoreAccess(ctx context.Context, db sqlx.ExtContext, scoreID, userID int64) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1
			FROM scores s
			WHERE s.id = $1
			  AND s.judge_id = $2
		)
	`
	args := []any{scoreID, userID}
	return r.accessQueryExecutor(ctx, db, q, args)
}
func (r *accessRepository) HasTeamAccess(ctx context.Context, db sqlx.ExtContext, teamID, userID int64) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1
			FROM teams t
			JOIN schools sc ON t.school_id = sc.id
			WHERE t.id = $1 AND sc.school_admin = $2
		)
	`
	args := []any{teamID, userID}
	return r.accessQueryExecutor(ctx, db, q, args)
}

func (r *accessRepository) accessQueryExecutor(ctx context.Context, db sqlx.ExtContext, query string, args []any) (bool, error) {
	var exists bool
	err := sqlx.GetContext(ctx, db, &exists, query, args...)
	if err != nil {
		return false, err
	}

	return exists, nil
}
