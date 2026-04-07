package core

import (
	"fmt"
	"strings"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleJudge UserRole = "judge"
)

func ParseUserRole(v string) (UserRole, error) {
	switch strings.ToLower(v) {
	case "admin":
		return RoleAdmin, nil
	case "judge":
		return RoleJudge, nil
	default:
		return "", fmt.Errorf("Invalid role, must be ['admin','judge'] ")
	}
}
