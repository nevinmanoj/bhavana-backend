package rbac

import "slices"

var RolePermissions = map[UserRole][]Permission{
	UserRoleAdmin: {
		PermViewUser,
		PermCreateEvent, PermUpdateEvent, PermDeleteEvent, PermViewEvent, PermUpdateEventStatus,
		PermCreateSchool, PermUpdateSchool, PermDeleteSchool, PermViewSchool,
		PermCreateStudent, PermUpdateStudent, PermDeleteStudent, PermViewStudent,
		PermCreateTeam, PermUpdateTeam, PermViewTeam, PermDeleteTeam,
	},
	UserRoleJudge: {
		PermViewEvent, PermViewTeam,
	},
}

func HasPermission(role UserRole, perm Permission) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	return slices.Contains(perms, perm)
}
