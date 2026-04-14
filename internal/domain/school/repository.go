package school

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type SchoolWriteRepository interface {
	SchoolReadRepository
	CreateSchool(ctx context.Context, db sqlx.ExtContext, SchoolToCreate *School) error
	UpdateSchool(ctx context.Context, db sqlx.ExtContext, SchoolToUpdate *School) error
	DeleteSchool(ctx context.Context, db sqlx.ExtContext, id int64) error

	CreateStudent(ctx context.Context, db sqlx.ExtContext, StudentToCreate *Student) error
	UpdateStudent(ctx context.Context, db sqlx.ExtContext, StudentToUpdate *Student) error
	DeleteStudent(ctx context.Context, db sqlx.ExtContext, id int64) error
}
type SchoolReadRepository interface {
	GetSchoolByID(ctx context.Context, db sqlx.ExtContext, id int64) (*School, error)
	GetAllSchools(ctx context.Context, db sqlx.ExtContext) ([]School, error)

	GetAllStudents(ctx context.Context, db sqlx.ExtContext, filter StudentFilter) ([]Student, error)
}
