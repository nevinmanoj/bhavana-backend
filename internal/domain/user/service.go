package user

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	auth "github.com/nevinmanoj/bhavana-backend/internal/auth"
	core "github.com/nevinmanoj/bhavana-backend/internal/core"
)

type UserService interface {
	CreateUser(ctx context.Context, user *User, password, jwtToken string) error
	LoginUser(ctx context.Context, email, password string) (string, *User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	GetAllUsers(ctx context.Context, filter UserFilter) ([]User, error)
}

type userService struct {
	repo      UserWriteRepository
	jwtSecret []byte
	db        *sqlx.DB
}

func NewUserService(repo UserWriteRepository, jwtSecret []byte, db *sqlx.DB) UserService {
	return &userService{repo: repo, jwtSecret: jwtSecret, db: db}
}

func (s *userService) CreateUser(ctx context.Context, user *User, password, jwtToken string) error {
	if user.Role == core.UserRoleAdmin {
		if jwtToken == "" {
			return fmt.Errorf("Unauthorized: JWT token is required to create admin users")
		}
		claims, err := auth.ParseToken(jwtToken, s.jwtSecret)
		if err != nil {
			return fmt.Errorf("Unauthorized")
		}
		if claims.Role != core.UserRoleAdmin {
			return fmt.Errorf("Forbidden: Only admins can create admin users")
		}
	}
	return s.repo.CreateUser(ctx, s.db, password, user)
}
func (s *userService) LoginUser(ctx context.Context, email, password string) (string, *User, error) {
	// Implementation for user login
	user, err := s.repo.GetUserByEmail(ctx, s.db, email)
	if err != nil {
		return "", nil, err
	}
	err = auth.CheckPassword(password, user.PasswordHash)
	if err != nil {
		return "", nil, fmt.Errorf("Invalid credentials")
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
