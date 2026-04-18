package school

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SchoolService interface {
	GetSchoolByID(ctx context.Context, id int64) (*School, error)
	GetAllSchools(ctx context.Context) ([]School, error)
	CreateSchool(ctx context.Context, school *School) error
	UpdateSchool(ctx context.Context, school *School) error
	DeleteSchool(ctx context.Context, id int64) error

	CreateStudent(ctx context.Context, student *Student) error
	UpdateStudent(ctx context.Context, student *Student) error
	DeleteStudent(ctx context.Context, id int64) error
	GetAllStudents(ctx context.Context, filter StudentFilter) ([]Student, error)
}

type schoolService struct {
	repo SchoolWriteRepository
	db   *sqlx.DB
}

func NewSchoolService(repo SchoolWriteRepository, db *sqlx.DB) SchoolService {
	return &schoolService{repo: repo, db: db}
}

// schools
func (s *schoolService) GetAllSchools(ctx context.Context) ([]School, error) {
	return s.repo.GetAllSchools(ctx, s.db)
}
func (s *schoolService) GetSchoolByID(ctx context.Context, id int64) (*School, error) {
	return s.repo.GetSchoolByID(ctx, s.db, id)
}
func (s *schoolService) CreateSchool(ctx context.Context, school *School) error {
	return s.repo.CreateSchool(ctx, s.db, school)
}
func (s *schoolService) UpdateSchool(ctx context.Context, school *School) error {
	return s.repo.UpdateSchool(ctx, s.db, school)
}
func (s *schoolService) DeleteSchool(ctx context.Context, id int64) error {
	return s.repo.DeleteSchool(ctx, s.db, id)
}

// students
func (s *schoolService) GetAllStudents(ctx context.Context, filter StudentFilter) ([]Student, error) {
	return s.repo.GetAllStudents(ctx, s.db, filter)
}
func (s *schoolService) UpdateStudent(ctx context.Context, student *Student) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()
	err = s.repo.UpdateStudent(ctx, tx, student)
	if err != nil {
		return err
	}
	student, err = s.repo.GetStudentByID(ctx, tx, student.ID)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *schoolService) CreateStudent(ctx context.Context, student *Student) error {
	return s.repo.CreateStudent(ctx, s.db, student)
}
func (s *schoolService) DeleteStudent(ctx context.Context, id int64) error {
	return s.repo.DeleteStudent(ctx, s.db, id)
}
