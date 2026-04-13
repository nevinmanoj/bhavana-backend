package user

import "github.com/nevinmanoj/bhavana-backend/internal/rbac"

type UserFilter struct {
	Roles []rbac.UserRole
}
