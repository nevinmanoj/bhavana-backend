package user

import (
	"time"

	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

type User struct {
	ID           int64         `db:"id"`
	Name         string        `db:"name"`
	Email        string        `db:"email"`
	PasswordHash string        `db:"password_hash"`
	Role         rbac.UserRole `db:"role"`
	CreatedAt    time.Time     `db:"created_at"`
}
