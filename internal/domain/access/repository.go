package access

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type AccessRepository interface {
	HasStudentAccess(ctx context.Context, db sqlx.ExtContext, studentID, userID int64) (bool, error)
	HasScoreAccess(ctx context.Context, db sqlx.ExtContext, scoreID, userID int64) (bool, error)
	HasSchoolAccess(ctx context.Context, db sqlx.ExtContext, schoolID, userID int64) (bool, error)
	HasTeamAccess(ctx context.Context, db sqlx.ExtContext, teamID, userID int64) (bool, error)
}
