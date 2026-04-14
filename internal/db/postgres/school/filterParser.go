package school

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
)

func buildStudentQuery(baseQuery string, f school.StudentFilter) (string, []any, error) {
	var (
		conditions []string
		args       []any
	)
	if f.SchoolID != nil {
		conditions = append(conditions, "s.school_id = ?")
		args = append(args, *f.SchoolID)
	}

	if f.Category != nil {
		conditions = append(conditions, "s.category = ?")
		args = append(args, *f.Category)
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
