package school

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
)

type SchoolWriteRepository interface {
	SchoolReadRepository
	CreateSchool(ctx context.Context, db sqlx.ExtContext, SchoolToCreate *School) error
	UpdateSchool(ctx context.Context, db sqlx.ExtContext, SchoolToUpdate *School) error
	UpdateSchoolStatus(ctx context.Context, db sqlx.ExtContext, id int64, status core.SchoolStatus) error
	DeleteSchool(ctx context.Context, db sqlx.ExtContext, id int64) error
	// LockSchoolForDelete row-locks the school and returns what the delete
	// path needs to decide whether its school admin account goes too.
	LockSchoolForDelete(ctx context.Context, db sqlx.ExtContext, id int64) (adminID *int64, status core.SchoolStatus, err error)
	SchoolExists(ctx context.Context, db sqlx.ExtContext, schoolID int64) (bool, error)
	SetRegistrationOpen(ctx context.Context, db sqlx.ExtContext, open bool) error

	CreateStudent(ctx context.Context, db sqlx.ExtContext, StudentToCreate *Student) error
	UpdateStudent(ctx context.Context, db sqlx.ExtContext, StudentToUpdate *Student) error
	DeleteStudent(ctx context.Context, db sqlx.ExtContext, id int64) error
}
type SchoolReadRepository interface {
	GetSchoolByID(ctx context.Context, db sqlx.ExtContext, id int64) (*School, error)
	GetAllSchools(ctx context.Context, db sqlx.ExtContext) ([]School, error)
	IsRegistrationOpen(ctx context.Context, db sqlx.ExtContext) (bool, error)

	GetAllStudents(ctx context.Context, db sqlx.ExtContext, filter StudentFilter) ([]Student, error)
	// CountStudents returns the total row count for filter, ignoring
	// filter.Limit/Offset - the total a paginated list needs alongside its page.
	CountStudents(ctx context.Context, db sqlx.ExtContext, filter StudentFilter) (int64, error)
	GetStudentByID(ctx context.Context, db sqlx.ExtContext, id int64) (*Student, error)
	StudentExists(ctx context.Context, db sqlx.ExtContext, studentID int64) (bool, error)
}
