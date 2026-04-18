package school

import (
	"context"
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
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
	baseQuery := `SELECT * FROM schools`
	err := sqlx.SelectContext(
		ctx, db,
		&schools,
		baseQuery,
	)
	if err != nil {
		return nil, err
	}
	return schools, nil
}
func (s *schoolRepository) GetSchoolByID(ctx context.Context, db sqlx.ExtContext, id int64) (*school.School, error) {
	schools := []school.School{}
	err := sqlx.SelectContext(
		ctx, db,
		&schools,
		`SELECT * FROM schools
		 WHERE id = $1`,
		id,
	)

	if err != nil {
		log.Println("Error fetching school by id:", err)
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
			address = :address,
			contact_name = :contact_name,
			contact_email = :contact_email,
			contact_phone = :contact_phone
		WHERE id = :id
		RETURNING created_at
	`
	rows, err := sqlx.NamedQueryContext(ctx, db, query, schoolToUpdate)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&schoolToUpdate.CreatedAt); err != nil {
			return err
		}

		return nil
	}
	return sql.ErrNoRows
}

// students
func (s *schoolRepository) UpdateStudent(ctx context.Context, db sqlx.ExtContext, studentToUpdate *school.Student) error {
	query := `
		UPDATE students
		SET name = :name
		WHERE id = :id
		RETURNING created_at
	`
	rows, err := sqlx.NamedQueryContext(ctx, db, query, studentToUpdate)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&studentToUpdate.CreatedAt); err != nil {
			return err
		}

		return nil
	}
	return sql.ErrNoRows
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
	baseQuery := `SELECT * FROM students s`
	finalQuery, args, err := buildStudentQuery(baseQuery, filter)
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
	err := sqlx.SelectContext(
		ctx, db,
		&students,
		`SELECT * FROM students
		 WHERE id = $1`,
		id,
	)

	if err != nil {
		log.Println("Error fetching school by id:", err)
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
