package access

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

type AccessService interface {
	CanModifyStudent(ctx context.Context, schoolID int64) (bool, error)
	CanModifyScore(ctx context.Context, scoreID int64) (bool, error)
	CanCreateStudent(ctx context.Context, schoolID int64) (bool, error)
	CanCreateTeam(ctx context.Context, schoolID int64) (bool, error)
}

type accessService struct {
	db   *sqlx.DB
	repo AccessRepository
}

func NewAccessService(db *sqlx.DB, repo AccessRepository) AccessService {
	return &accessService{db: db, repo: repo}
}

// schools,students
func (s *accessService) CanCreateStudent(ctx context.Context, schoolID int64) (bool, error) {
	userID := ctx.Value(middleware.ContextUserID).(int64)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	if role == rbac.UserRoleAdmin {
		return true, nil
	}
	canAccess, err := s.repo.HasSchoolAccess(ctx, s.db, schoolID, userID)
	if err != nil {
		return false, err
	}
	return canAccess, nil
}
func (s *accessService) CanModifyStudent(ctx context.Context, studentID int64) (bool, error) {
	userID := ctx.Value(middleware.ContextUserID).(int64)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	if role == rbac.UserRoleAdmin {
		return true, nil
	}
	canAccess, err := s.repo.HasStudentAccess(ctx, s.db, studentID, userID)
	if err != nil {
		return false, err
	}
	return canAccess, nil
}

// teams
func (s *accessService) CanCreateTeam(ctx context.Context, schoolID int64) (bool, error) {
	userID := ctx.Value(middleware.ContextUserID).(int64)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	if role == rbac.UserRoleAdmin {
		return true, nil
	}
	canAccess, err := s.repo.HasSchoolAccess(ctx, s.db, schoolID, userID)
	if err != nil {
		return false, err
	}
	return canAccess, nil
}
func (s *accessService) CanModifyTeam(ctx context.Context, teamID int64) (bool, error) {
	userID := ctx.Value(middleware.ContextUserID).(int64)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	if role == rbac.UserRoleAdmin {
		return true, nil
	}
	canAccess, err := s.repo.HasTeamAccess(ctx, s.db, teamID, userID)
	if err != nil {
		return false, err
	}
	return canAccess, nil
}

// scores
func (s *accessService) CanModifyScore(ctx context.Context, scoreID int64) (bool, error) {
	userID := ctx.Value(middleware.ContextUserID).(int64)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	if role == rbac.UserRoleAdmin {
		return true, nil
	}
	canAccess, err := s.repo.HasScoreAccess(ctx, s.db, scoreID, userID)
	if err != nil {
		return false, err
	}
	return canAccess, nil
}
