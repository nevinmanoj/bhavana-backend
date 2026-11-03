package user

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	auth "github.com/nevinmanoj/bhavana-backend/internal/auth"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/mail"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
	"github.com/nevinmanoj/bhavana-backend/internal/util"
)

const (
	resetTokenTTL         = 24 * time.Hour
	resetRequestThrottle  = 2 * time.Minute
	backgroundMailTimeout = 30 * time.Second
)

type UserService interface {
	CreateUser(ctx context.Context, user *User, password, jwtToken string) error
	LoginUser(ctx context.Context, email, password string) (string, *User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	GetAllUsers(ctx context.Context, filter UserFilter) ([]User, error)
	// ForgotPassword emails a reset link in the background. It reports
	// nothing, so callers can't learn whether the account exists.
	ForgotPassword(email string)
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error
}

type userService struct {
	db        *sqlx.DB
	jwtSecret []byte
	repo      UserWriteRepository
	mailer    mail.Mailer
	appURL    string
}

func NewUserService(db *sqlx.DB, jwtSecret []byte, repo UserWriteRepository, mailer mail.Mailer, appURL string) UserService {
	return &userService{db: db, jwtSecret: jwtSecret, repo: repo, mailer: mailer, appURL: appURL}
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
	// No password yet: a registered school admin waiting for acceptance is
	// told so, since there is no password they could have typed right.
	if user.PasswordHash == "" {
		if user.Role == rbac.UserRoleSchoolAdmin {
			status, err := s.repo.GetSchoolStatusForAdmin(ctx, s.db, user.ID)
			if err != nil {
				return "", nil, ErrInternal
			}
			if status != nil && *status == core.SchoolStatusInReview {
				return "", nil, ErrRegistrationUnderReview
			}
		}
		return "", nil, ErrInvalidCredentials
	}
	err = auth.CheckPassword(password, user.PasswordHash)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// A school admin whose school was moved back to review can't sign in.
	// Checked after the password so this path never leaks registration state.
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

func (s *userService) ForgotPassword(email string) {
	email = util.NormalizeEmail(email)
	// Detached from the request: the response must not wait on (and so time
	// out differently for) accounts that exist.
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), backgroundMailTimeout)
		defer cancel()
		if err := s.sendPasswordReset(ctx, email); err != nil {
			log.Printf("forgot password: %v", err)
		}
	}()
}

func (s *userService) sendPasswordReset(ctx context.Context, email string) error {
	u, err := s.repo.GetUserByEmail(ctx, s.db, email)
	if err == ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	// no password yet means a registration still in review: acceptance, not
	// a reset link, gives them their first password
	if u.PasswordHash == "" {
		return nil
	}
	recent, err := s.repo.HasRecentResetToken(ctx, s.db, u.ID, time.Now().Add(-resetRequestThrottle))
	if err != nil {
		return err
	}
	if recent {
		return nil
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.repo.InvalidateResetTokens(ctx, tx, u.ID); err != nil {
		return err
	}
	token, hash, err := auth.NewResetToken()
	if err != nil {
		return err
	}
	if err := s.repo.CreateResetToken(ctx, tx, u.ID, hash, time.Now().Add(resetTokenTTL)); err != nil {
		return err
	}
	if err := s.mailer.Send(ctx, mail.PasswordResetEmail(s.appURL, u.Email, u.Name, token)); err != nil {
		return fmt.Errorf("send reset email to user %d: %w", u.ID, err)
	}
	return tx.Commit()
}

func (s *userService) ResetPassword(ctx context.Context, token, newPassword string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return ErrInternal
	}
	defer tx.Rollback()

	userID, err := s.repo.ConsumeResetToken(ctx, tx, auth.HashResetToken(token))
	if err != nil {
		return err
	}
	if err := s.setPassword(ctx, tx, userID, newPassword); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return ErrInternal
	}
	return nil
}

func (s *userService) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return ErrInternal
	}
	defer tx.Rollback()

	u, err := s.repo.GetUserByID(ctx, tx, userID)
	if err != nil {
		return err
	}
	if auth.CheckPassword(currentPassword, u.PasswordHash) != nil {
		return ErrWrongCurrentPassword
	}
	if err := s.setPassword(ctx, tx, userID, newPassword); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return ErrInternal
	}
	return nil
}

// setPassword stores the new password's hash and kills any open reset links,
// which would otherwise still let someone replace it.
func (s *userService) setPassword(ctx context.Context, tx sqlx.ExtContext, userID int64, password string) error {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return ErrInternal
	}
	if err := s.repo.SetPasswordHash(ctx, tx, userID, hash); err != nil {
		return err
	}
	if err := s.repo.InvalidateResetTokens(ctx, tx, userID); err != nil {
		return ErrInternal
	}
	return nil
}

func (s *userService) GetAllUsers(ctx context.Context, filter UserFilter) ([]User, error) {
	users, err := s.repo.GetAllUsers(ctx, s.db, filter)
	if err != nil || len(users) == 0 || users == nil {
		return []User{}, err
	}

	return users, nil
}
