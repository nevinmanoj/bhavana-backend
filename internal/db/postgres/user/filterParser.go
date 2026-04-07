package user

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/user"
)

func buildUserQuery(baseQuery string, f user.UserFilter) (string, []any, error) {
	var (
		conditions []string
		args       []any
	)

	if f.Name != nil {
		conditions = append(conditions, "u.name ILIKE ?")
		args = append(args, "%"+*f.Name+"%")
	}

	if len(f.Roles) > 0 {
		conditions = append(conditions, "u.role IN (?)")
		args = append(args, f.Roles)
	}

	// Apply WHERE
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Ordering (always deterministic)
	baseQuery += " ORDER BY created_at DESC"

	// Pagination
	// if f.Limit > 0 {
	// 	baseQuery += " LIMIT ?"
	// 	args = append(args, f.Limit)
	// }

	// if f.Offset > 0 {
	// 	baseQuery += " OFFSET ?"
	// 	args = append(args, f.Offset)
	// }

	// Expand IN clauses
	query, finalArgs, err := sqlx.In(baseQuery, args...)
	if err != nil {
		return "", nil, err
	}

	// Rebind for postgres ($1, $2...)
	query = sqlx.Rebind(sqlx.DOLLAR, query)

	return query, finalArgs, nil
}
