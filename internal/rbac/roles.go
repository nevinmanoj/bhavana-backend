package rbac

import (
	"fmt"
	"strings"
)

type UserRole string
type Permission string

const (
	UserRoleAdmin       UserRole = "admin"
	UserRoleJudge       UserRole = "judge"
	UserRoleSchoolAdmin UserRole = "school_admin"
)

func ParseUserRole(v string) (UserRole, error) {
	switch strings.ToLower(v) {
	case "admin":
		return UserRoleAdmin, nil
	case "judge":
		return UserRoleJudge, nil
	case "school_admin":
		return UserRoleSchoolAdmin, nil
	default:
		return "", fmt.Errorf("Invalid role, must be ['admin','judge','school_admin'] ")
	}
}
func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleAdmin, UserRoleJudge, UserRoleSchoolAdmin:
		return true
	}
	return false
}
