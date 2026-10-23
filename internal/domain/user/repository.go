package user

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
)

type UserWriteRepository interface {
	UserReadRepository
	CreateUser(ctx context.Context, db sqlx.ExtContext, password string, userToCreate *User) error
	DeleteUser(ctx context.Context, db sqlx.ExtContext, id int64) error
}
type UserReadRepository interface {
	GetUserByEmail(ctx context.Context, db sqlx.ExtContext, email string) (*User, error)
	GetUserByID(ctx context.Context, db sqlx.ExtContext, id int64) (*User, error)
	GetAllUsers(ctx context.Context, db sqlx.ExtContext, filter UserFilter) ([]User, error)
	ExistsAsJudge(ctx context.Context, db sqlx.ExtContext, userID int64) (bool, error)
	// GetSchoolStatusForAdmin returns the status of the school userID administers, or nil if none.
	GetSchoolStatusForAdmin(ctx context.Context, db sqlx.ExtContext, userID int64) (*core.SchoolStatus, error)
}
