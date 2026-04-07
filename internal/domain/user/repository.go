package user

import (
	"context"

	core "github.com/nevinmanoj/bhavana-backend/internal/core"
)

type UserWriteRepository interface {
	UserReadRepository
	CreateUser(ctx context.Context, email, password, name string, role core.UserRole) (*User, error)
}
type UserReadRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	GetAllUsers(ctx context.Context, filter UserFilter) ([]User, error)
}
