package user

import (
	"time"

	core "github.com/nevinmanoj/bhavana-backend/internal/core"
)

type User struct {
	ID           int64         `db:"id"`
	Name         string        `db:"name"`
	Email        string        `db:"email"`
	PasswordHash string        `db:"password_hash"`
	Role         core.UserRole `db:"role"`
	CreatedAt    time.Time     `db:"created_at"`
}
