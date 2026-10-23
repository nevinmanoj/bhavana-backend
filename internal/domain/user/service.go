package user

import (
	"context"

	"github.com/jmoiron/sqlx"
	auth "github.com/nevinmanoj/bhavana-backend/internal/auth"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
	"github.com/nevinmanoj/bhavana-backend/internal/util"
)

type UserService interface {
	CreateUser(ctx context.Context, user *User, password, jwtToken string) error
	LoginUser(ctx context.Context, email, password string) (string, *User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	GetAllUsers(ctx context.Context, filter UserFilter) ([]User, error)
}

type userService struct {
	db        *sqlx.DB
	jwtSecret []byte
	repo      UserWriteRepository
}

func NewUserService(db *sqlx.DB, jwtSecret []byte, repo UserWriteRepository) UserService {
	return &userService{db: db, jwtSecret: jwtSecret, repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, user *User, password, jwtToken string) error {
	user.Email = util.NormalizeEmail(user.Email)
	if user.Role == rbac.UserRoleAdmin || user.Role == rbac.UserRoleSchoolAdmin {
		if jwtToken == "" {
			return ErrUnauthorized
		}
		claims, err := auth.ParseToken(jwtToken, s.jwtSecret)
		if err != nil {
			return ErrUnauthorized
		}
		if claims.Role != rbac.UserRoleAdmin {
			return ErrUnauthorized
		}
	}
	return s.repo.CreateUser(ctx, s.db, password, user)
}
func (s *userService) LoginUser(ctx context.Context, email, password string) (string, *User, error) {
	// Implementation for user login
	user, err := s.repo.GetUserByEmail(ctx, s.db, util.NormalizeEmail(email))
	if err != nil {
		return "", nil, err
	}
	err = auth.CheckPassword(password, user.PasswordHash)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// A school admin whose school is still in review can't sign in yet. Checked
	// after the password so the response never leaks registration state.
	if user.Role == rbac.UserRoleSchoolAdmin {
		status, err := s.repo.GetSchoolStatusForAdmin(ctx, s.db, user.ID)
		if err != nil {
			return "", nil, ErrInternal
		}
		if status != nil && *status == core.SchoolStatusInReview {
			return "", nil, ErrRegistrationUnderReview
		}
	}

	// create and issue JWT token
	token, err := auth.GenerateToken(user.Role, user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetUserByEmail(ctx, s.db, email)
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*User, error) {
	return s.repo.GetUserByID(ctx, s.db, id)
}

func (s *userService) GetAllUsers(ctx context.Context, filter UserFilter) ([]User, error) {
	users, err := s.repo.GetAllUsers(ctx, s.db, filter)
	if err != nil || len(users) == 0 || users == nil {
		return []User{}, err
	}

	return users, nil
}
