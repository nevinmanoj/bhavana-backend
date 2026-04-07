package user

import core "github.com/nevinmanoj/bhavana-backend/internal/core"

type UserFilter struct {
	Name  *string
	Roles []core.UserRole
}
