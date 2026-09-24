package school

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
)

// appendStudentConditions turns the filter's own fields into WHERE
// conditions, shared between the data query (buildStudentQuery) and the
// count query (buildStudentCountQuery) so the two never drift out of sync.
func appendStudentConditions(conditions []string, args []any, f school.StudentFilter) ([]string, []any) {
	if f.SchoolID != nil {
		conditions = append(conditions, "s.school_id = ?")
		args = append(args, *f.SchoolID)
	}
	if f.Category != nil {
		conditions = append(conditions, "s.category = ?")
		args = append(args, *f.Category)
	}
	return conditions, args
}

func buildStudentQuery(baseQuery string, args []any, conditions []string, f school.StudentFilter) (string, []any, error) {
	conditions, args = appendStudentConditions(conditions, args, f)

	// Apply WHERE
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Ordering (always deterministic)
	baseQuery += " ORDER BY created_at DESC"

	// Pagination is opt-in: Limit/Offset are 0 for every caller that never
	// asked for a page (see StudentFilter), so this stays a no-op for them.
	if f.Limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, f.Limit)
	}

	if f.Offset > 0 {
		baseQuery += " OFFSET ?"
		args = append(args, f.Offset)
	}

	// Expand IN clauses
	query, finalArgs, err := sqlx.In(baseQuery, args...)
	if err != nil {
		return "", nil, err
	}

	// Rebind for postgres ($1, $2...)
	query = sqlx.Rebind(sqlx.DOLLAR, query)

	return query, finalArgs, nil
}

// buildStudentCountQuery mirrors buildStudentQuery's WHERE clause exactly
// (via appendStudentConditions) but with no ORDER BY/LIMIT/OFFSET, for the
// total count a paginated list needs alongside its current page. baseQuery
// must already be a `SELECT COUNT(*) FROM students s ...` with the same
// JOINs/scope conditions as the data query.
func buildStudentCountQuery(baseQuery string, args []any, conditions []string, f school.StudentFilter) (string, []any, error) {
	conditions, args = appendStudentConditions(conditions, args, f)

	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	query, finalArgs, err := sqlx.In(baseQuery, args...)
	if err != nil {
		return "", nil, err
	}

	query = sqlx.Rebind(sqlx.DOLLAR, query)

	return query, finalArgs, nil
}
