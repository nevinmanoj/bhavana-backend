package school

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
)

// LockedSchool is the slice of a row-locked school that the service decides on.
type LockedSchool struct {
	SchoolAdmin  *int64            `db:"school_admin"`
	Status       core.SchoolStatus `db:"status"`
	Name         string            `db:"name"`
	ContactEmail string            `db:"contact_email"`
}

type SchoolWriteRepository interface {
	SchoolReadRepository
	CreateSchool(ctx context.Context, db sqlx.ExtContext, SchoolToCreate *School) error
	UpdateSchool(ctx context.Context, db sqlx.ExtContext, SchoolToUpdate *School) error
	// UpdateSchoolDetails writes what a school admin may change: every field
	// except name, contact email, school admin and status.
	UpdateSchoolDetails(ctx context.Context, db sqlx.ExtContext, SchoolToUpdate *School) error
	UpdateSchoolStatus(ctx context.Context, db sqlx.ExtContext, id int64, status core.SchoolStatus) error
	DeleteSchool(ctx context.Context, db sqlx.ExtContext, id int64) error
	// LockSchool row-locks the school for the rest of the transaction and
	// returns what delete, update, accept and resend decide on.
	LockSchool(ctx context.Context, db sqlx.ExtContext, id int64) (*LockedSchool, error)
	UpdateContactEmail(ctx context.Context, db sqlx.ExtContext, id int64, email string) error
	SchoolExists(ctx context.Context, db sqlx.ExtContext, schoolID int64) (bool, error)
	UpdateSettings(ctx context.Context, db sqlx.ExtContext, settings *Settings) error

	CreateStudent(ctx context.Context, db sqlx.ExtContext, StudentToCreate *Student) error
	UpdateStudent(ctx context.Context, db sqlx.ExtContext, StudentToUpdate *Student) error
	DeleteStudent(ctx context.Context, db sqlx.ExtContext, id int64) error
}
type SchoolReadRepository interface {
	GetSchoolByID(ctx context.Context, db sqlx.ExtContext, id int64) (*School, error)
	GetAllSchools(ctx context.Context, db sqlx.ExtContext) ([]School, error)
	GetSettings(ctx context.Context, db sqlx.ExtContext) (*Settings, error)

	GetAllStudents(ctx context.Context, db sqlx.ExtContext, filter StudentFilter) ([]Student, error)
	// CountStudents returns the total row count for filter, ignoring
	// filter.Limit/Offset - the total a paginated list needs alongside its page.
	CountStudents(ctx context.Context, db sqlx.ExtContext, filter StudentFilter) (int64, error)
	GetStudentByID(ctx context.Context, db sqlx.ExtContext, id int64) (*Student, error)
	StudentExists(ctx context.Context, db sqlx.ExtContext, studentID int64) (bool, error)
}
