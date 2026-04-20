package school

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/access"
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
	db            *sqlx.DB
	accessService access.AccessService
	repo          SchoolWriteRepository
}

func NewSchoolService(db *sqlx.DB, accessService access.AccessService, repo SchoolWriteRepository) SchoolService {
	return &schoolService{db: db, accessService: accessService, repo: repo}
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
	access, err := s.accessService.CanModifyStudent(ctx, student.ID)
	if err != nil {
		return err
	}
	if !access {
		return ErrUnauthorized
	}
	return s.repo.UpdateStudent(ctx, s.db, student)
}
func (s *schoolService) CreateStudent(ctx context.Context, student *Student) error {
	access, err := s.accessService.CanCreateStudent(ctx, student.SchoolID)
	if err != nil {
		return err
	}
	if !access {
		return ErrUnauthorized
	}
	return s.repo.CreateStudent(ctx, s.db, student)
}
func (s *schoolService) DeleteStudent(ctx context.Context, id int64) error {
	access, err := s.accessService.CanModifyStudent(ctx, id)
	if err != nil {
		return err
	}
	if !access {
		return ErrUnauthorized
	}
	return s.repo.DeleteStudent(ctx, s.db, id)
}
