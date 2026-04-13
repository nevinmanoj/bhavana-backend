package user

import (
	"net/url"
	"strings"

	errmap "github.com/nevinmanoj/bhavana-backend/internal/app/errmap"
	user "github.com/nevinmanoj/bhavana-backend/internal/domain/user"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

func parseUserFilter(q url.Values) (user.UserFilter, *errmap.BadRequestError) {
	var f user.UserFilter

	if v := q.Get("role"); v != "" {
		roles, err := parseUserRoleSlice(v)
		if err != nil {
			return f, &errmap.BadRequestError{
				Param:  "role",
				Reason: err.Error(),
			}
		}
		f.Roles = roles
	}

	// // Pagination defaults
	// f.Limit = 100
	// f.Offset = 0

	// if v := q.Get("limit"); v != "" {
	// 	limit, err := strconv.Atoi(v)
	// 	if err != nil {
	// 		return f, &errMap.BadRequestError{
	// 			Param:  "limit",
	// 			Reason: err.Error(),
	// 		}
	// 	} else if limit > 0 && limit < 100 {
	// 		f.Limit = limit
	// 	}
	// }

	// if v := q.Get("offset"); v != "" {
	// 	offset, err := strconv.Atoi(v)
	// 	if err != nil {
	// 		return f, &errMap.BadRequestError{
	// 			Param:  "offset",
	// 			Reason: err.Error(),
	// 		}
	// 	} else if offset > 0 {
	// 		f.Offset = offset
	// 	}
	// }

	return f, nil
}

func parseUserRoleSlice(v string) ([]rbac.UserRole, error) {
	parts := strings.Split(v, ",")
	out := make([]rbac.UserRole, 0, len(parts))

	for _, p := range parts {
		t, err := rbac.ParseUserRole(strings.TrimSpace(p))
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}

	return out, nil
}
