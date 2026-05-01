package school

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

type schoolRepository struct {
}

func NewSchoolWriteRepository() school.SchoolWriteRepository {
	return &schoolRepository{}
}
func NewSchoolReadRepository() school.SchoolReadRepository {
	return &schoolRepository{}
}

// schools
func (s *schoolRepository) CreateSchool(ctx context.Context, db sqlx.ExtContext, schoolToCreate *school.School) error {
	query := `
		INSERT INTO schools (
			name,
			address,
			contact_name,
			contact_email,
			contact_phone
		)
		VALUES (
			:name,
			:address,
			:contact_name,
			:contact_email,
			:contact_phone
		)
		RETURNING id, created_at
	`

	rows, err := sqlx.NamedQueryContext(ctx, db, query, schoolToCreate)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&schoolToCreate.ID, &schoolToCreate.CreatedAt)
		if err != nil {
			return err
		}
		return nil
	}

	return sql.ErrNoRows
}
func (s *schoolRepository) DeleteSchool(ctx context.Context, db sqlx.ExtContext, id int64) error {
	query := `DELETE FROM schools WHERE id = $1`
	_, err := db.ExecContext(ctx, query, id)
	return err
}
func (s *schoolRepository) GetAllSchools(ctx context.Context, db sqlx.ExtContext) ([]school.School, error) {
	schools := []school.School{}
	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	baseQuery := `SELECT * FROM schools`
	args := []any{}
	if scope.UserID != nil && role == rbac.UserRoleSchoolAdmin {
		baseQuery += " WHERE school_admin = $1"
		args = append(args, scope.UserID)
	}
	err := sqlx.SelectContext(
		ctx, db,
		&schools,
		baseQuery, args...,
	)
	if err != nil {
		return nil, err
	}
	return schools, nil
}
func (s *schoolRepository) GetSchoolByID(ctx context.Context, db sqlx.ExtContext, id int64) (*school.School, error) {
	schools := []school.School{}
	baseQuery := `SELECT * FROM schools
		 			WHERE id = $1`
	args := []any{id}
	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	if scope.UserID != nil && role == rbac.UserRoleSchoolAdmin {
		baseQuery += " AND school_admin = $2"
		args = append(args, scope.UserID)
	}
	err := sqlx.SelectContext(
		ctx, db,
		&schools,
		baseQuery,
		args...,
	)

	if err != nil {
		return nil, school.ErrInternal
	}

	if len(schools) == 0 {
		return nil, school.ErrSchoolNotFound
	}
	school := schools[0]
	return &school, nil
}
func (s *schoolRepository) UpdateSchool(ctx context.Context, db sqlx.ExtContext, schoolToUpdate *school.School) error {
	query := `
		UPDATE schools
		SET name = :name,
			school_admin = :school_admin,
			address = :address,
			contact_name = :contact_name,
			contact_email = :contact_email,
			contact_phone = :contact_phone
		WHERE id = :id
	`
	result, err := sqlx.NamedExecContext(ctx, db, query, schoolToUpdate)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return school.ErrSchoolNotFound
	}
	return nil
}

// students
func (s *schoolRepository) UpdateStudent(ctx context.Context, db sqlx.ExtContext, studentToUpdate *school.Student) error {
	query := `
		UPDATE students
		SET name = :name
		WHERE id = :id
	`
	result, err := sqlx.NamedExecContext(ctx, db, query, studentToUpdate)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return school.ErrStudentNotFound
	}
	return nil
}
func (s *schoolRepository) CreateStudent(ctx context.Context, db sqlx.ExtContext, studentToCreate *school.Student) error {
	query := `
		INSERT INTO students (
			name,
			school_id,
			age,
			category
		)
		VALUES (
			:name,
			:school_id,
			:age,
			:category
		)
		RETURNING id, created_at
	`

	rows, err := sqlx.NamedQueryContext(ctx, db, query, studentToCreate)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&studentToCreate.ID, &studentToCreate.CreatedAt)
		return nil
	}

	return sql.ErrNoRows
}
func (s *schoolRepository) DeleteStudent(ctx context.Context, db sqlx.ExtContext, id int64) error {
	query := `DELETE FROM students WHERE id = $1`
	_, err := db.ExecContext(ctx, query, id)
	return err
}
func (s *schoolRepository) GetAllStudents(ctx context.Context, db sqlx.ExtContext, filter school.StudentFilter) ([]school.Student, error) {
	students := []school.Student{}
	baseQuery := `SELECT s.* FROM students s`
	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	args := []any{}
	conditions := []string{}
	if scope.UserID != nil && role == rbac.UserRoleSchoolAdmin {
		baseQuery += " JOIN schools sc ON s.school_id = sc.id"
		conditions = append(conditions, "sc.school_admin = ?")
		args = append(args, scope.UserID)
	}
	finalQuery, args, err := buildStudentQuery(baseQuery, args, conditions, filter)
	if err != nil {
		return nil, err
	}
	err = sqlx.SelectContext(
		ctx, db,
		&students,
		finalQuery, args...,
	)
	if err != nil {
		return nil, err
	}
	return students, nil
}
func (s *schoolRepository) GetStudentByID(ctx context.Context, db sqlx.ExtContext, id int64) (*school.Student, error) {
	students := []school.Student{}
	args := []any{id}
	conditions := []string{"s.id = ?"}
	baseQuery := `SELECT s.* FROM students s`
	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	if scope.UserID != nil && role == rbac.UserRoleSchoolAdmin {
		baseQuery += " JOIN schools sc ON sc.id = s.school_id "
		conditions = append(conditions, "sc.school_admin = ?")
		args = append(args, scope.UserID)
	}
	finalQuery, finalArgs, err := buildStudentQuery(baseQuery, args, conditions, school.StudentFilter{})
	if err != nil {
		return nil, err
	}
	err = sqlx.SelectContext(
		ctx, db,
		&students,
		finalQuery,
		finalArgs...,
	)

	if err != nil {
		return nil, school.ErrInternal
	}

	if len(students) == 0 {
		return nil, school.ErrStudentNotFound
	}
	student := students[0]
	return &student, nil
}
func (s *schoolRepository) StudentExists(ctx context.Context, db sqlx.ExtContext, studentID int64) (bool, error) {
	var count int
	err := db.QueryRowxContext(ctx,
		`SELECT COUNT(*) FROM students WHERE id = $1`, studentID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
