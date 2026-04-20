package rbac

type Scope struct {
	UserID *int64
}

func ResolveScope(userid int64, role UserRole) Scope {
	switch role {
	case UserRoleAdmin:
		return Scope{}
	case UserRoleJudge:
		return Scope{UserID: &userid}
	case UserRoleSchoolAdmin:
		return Scope{UserID: &userid}
	default:
		return Scope{UserID: &userid}
	}
}
