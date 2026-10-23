package school

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/access"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/user"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

type SchoolService interface {
	GetSchoolByID(ctx context.Context, id int64) (*School, error)
	GetAllSchools(ctx context.Context) ([]School, error)
	CreateSchool(ctx context.Context, school *School) error
	// RegisterSchool is the public sign-up: it creates the school admin user and
	// the school (in_review) in one transaction, if registration is open.
	RegisterSchool(ctx context.Context, school *School, admin *user.User, password string) error
	UpdateSchoolStatus(ctx context.Context, id int64, status core.SchoolStatus) error
	GetRegistrationOpen(ctx context.Context) (bool, error)
	SetRegistrationOpen(ctx context.Context, open bool) error
	UpdateSchool(ctx context.Context, school *School) error
	DeleteSchool(ctx context.Context, id int64) error

	CreateStudent(ctx context.Context, student *Student) error
	UpdateStudent(ctx context.Context, student *Student) error
	DeleteStudent(ctx context.Context, id int64) error
	// GetAllStudents returns the matching page (per filter.Limit/Offset, or
	// every row if unset) plus the total count for that same filter.
	GetAllStudents(ctx context.Context, filter StudentFilter) ([]Student, int64, error)
}

type schoolService struct {
	db            *sqlx.DB
	accessService access.AccessService
	repo          SchoolWriteRepository
	userRepo      user.UserWriteRepository
}

func NewSchoolService(db *sqlx.DB, accessService access.AccessService, repo SchoolWriteRepository, userRepo user.UserWriteRepository) SchoolService {
	return &schoolService{db: db, accessService: accessService, repo: repo, userRepo: userRepo}
}

// schools
func (s *schoolService) GetAllSchools(ctx context.Context) ([]School, error) {
	return s.repo.GetAllSchools(ctx, s.db)
}
func (s *schoolService) GetSchoolByID(ctx context.Context, id int64) (*School, error) {
	return s.repo.GetSchoolByID(ctx, s.db, id)
}
func (s *schoolService) CreateSchool(ctx context.Context, school *School) error {
	// admin-created schools skip review
	school.Status = core.SchoolStatusAccepted
	return s.repo.CreateSchool(ctx, s.db, school)
}
func (s *schoolService) UpdateSchool(ctx context.Context, school *School) error {
	return s.repo.UpdateSchool(ctx, s.db, school)
}
func (s *schoolService) DeleteSchool(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return ErrInternal
	}
	defer tx.Rollback()

	adminID, status, err := s.repo.LockSchoolForDelete(ctx, tx, id)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteSchool(ctx, tx, id); err != nil {
		return err
	}
	// rejecting an in-review registration must not leave its login behind
	if status == core.SchoolStatusInReview && adminID != nil {
		if err := s.userRepo.DeleteUser(ctx, tx, *adminID); err != nil {
			return ErrInternal
		}
	}
	if err := tx.Commit(); err != nil {
		return ErrInternal
	}
	return nil
}

func (s *schoolService) RegisterSchool(ctx context.Context, school *School, admin *user.User, password string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return ErrInternal
	}
	defer tx.Rollback()

	open, err := s.repo.IsRegistrationOpen(ctx, tx)
	if err != nil {
		return ErrInternal
	}
	if !open {
		return ErrRegistrationClosed
	}

	admin.Role = rbac.UserRoleSchoolAdmin
	if err := s.userRepo.CreateUser(ctx, tx, password, admin); err != nil {
		if err == user.ErrAlreadyExists {
			return ErrEmailTaken
		}
		return err
	}

	school.SchoolAdmin = &admin.ID
	school.Status = core.SchoolStatusInReview
	if err := s.repo.CreateSchool(ctx, tx, school); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return ErrInternal
	}
	return nil
}

func (s *schoolService) UpdateSchoolStatus(ctx context.Context, id int64, status core.SchoolStatus) error {
	return s.repo.UpdateSchoolStatus(ctx, s.db, id, status)
}
func (s *schoolService) GetRegistrationOpen(ctx context.Context) (bool, error) {
	return s.repo.IsRegistrationOpen(ctx, s.db)
}
func (s *schoolService) SetRegistrationOpen(ctx context.Context, open bool) error {
	return s.repo.SetRegistrationOpen(ctx, s.db, open)
}

// students
func (s *schoolService) GetAllStudents(ctx context.Context, filter StudentFilter) ([]Student, int64, error) {
	students, err := s.repo.GetAllStudents(ctx, s.db, filter)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountStudents(ctx, s.db, filter)
	if err != nil {
		return nil, 0, err
	}
	return students, total, nil
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
	exists, err := s.repo.SchoolExists(ctx, s.db, student.SchoolID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrSchoolNotFound
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
