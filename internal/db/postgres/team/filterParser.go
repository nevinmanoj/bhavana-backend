package team

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/team"
)

func buildTeamQuery(baseQuery string, conditions []string, args []any, f team.TeamFilter) (string, []any, error) {

	if f.SchoolID != nil {
		conditions = append(conditions, "t.school_id = ?")
		args = append(args, *f.SchoolID)
	}
	if f.EventID != nil {
		conditions = append(conditions, "t.event_id = ?")
		args = append(args, *f.EventID)
	}
	if f.Category != nil {
		conditions = append(conditions, "t.category = ?")
		args = append(args, *f.Category)
	}

	// Apply WHERE
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Ordering (always deterministic)
	baseQuery += " ORDER BY t.created_at DESC"

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
