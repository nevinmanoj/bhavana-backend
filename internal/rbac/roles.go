package rbac

import (
	"fmt"
	"strings"
)

type UserRole string
type Permission string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleJudge UserRole = "judge"
)

func ParseUserRole(v string) (UserRole, error) {
	switch strings.ToLower(v) {
	case "admin":
		return UserRoleAdmin, nil
	case "judge":
		return UserRoleJudge, nil
	default:
		return "", fmt.Errorf("Invalid role, must be ['admin','judge'] ")
	}
}
func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleAdmin, UserRoleJudge:
		return true
	}
	return false
}
