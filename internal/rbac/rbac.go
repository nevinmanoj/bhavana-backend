package rbac

import "slices"

var RolePermissions = map[UserRole][]Permission{
	UserRoleAdmin: {
		PermCreateEvent, PermUpdateEvent, PermDeleteEvent, PermViewEvent,
	},
	UserRoleJudge: {
		PermViewEvent,
	},
}

func HasPermission(role UserRole, perm Permission) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	return slices.Contains(perms, perm)
}
