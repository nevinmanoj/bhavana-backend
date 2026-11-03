package user

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
)

type UserWriteRepository interface {
	UserReadRepository
	CreateUser(ctx context.Context, db sqlx.ExtContext, password string, userToCreate *User) error
	DeleteUser(ctx context.Context, db sqlx.ExtContext, id int64) error
	SetPasswordHash(ctx context.Context, db sqlx.ExtContext, id int64, hash string) error
	// UpdateEmail returns ErrAlreadyExists if another account has the email.
	UpdateEmail(ctx context.Context, db sqlx.ExtContext, id int64, email string) error

	CreateResetToken(ctx context.Context, db sqlx.ExtContext, userID int64, hash []byte, expiresAt time.Time) error
	HasRecentResetToken(ctx context.Context, db sqlx.ExtContext, userID int64, since time.Time) (bool, error)
	// ConsumeResetToken marks an open, unexpired token used and returns its
	// user, or ErrInvalidResetToken. One statement, so a token can't be used twice.
	ConsumeResetToken(ctx context.Context, db sqlx.ExtContext, hash []byte) (int64, error)
	// InvalidateResetTokens marks every open token of the user used.
	InvalidateResetTokens(ctx context.Context, db sqlx.ExtContext, userID int64) error
}
type UserReadRepository interface {
	GetUserByEmail(ctx context.Context, db sqlx.ExtContext, email string) (*User, error)
	GetUserByID(ctx context.Context, db sqlx.ExtContext, id int64) (*User, error)
	GetAllUsers(ctx context.Context, db sqlx.ExtContext, filter UserFilter) ([]User, error)
	ExistsAsJudge(ctx context.Context, db sqlx.ExtContext, userID int64) (bool, error)
	// GetSchoolStatusForAdmin returns the status of the school userID administers, or nil if none.
	GetSchoolStatusForAdmin(ctx context.Context, db sqlx.ExtContext, userID int64) (*core.SchoolStatus, error)
}
