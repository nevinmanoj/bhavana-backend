package event

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/event"
)

func buildEventQuery(baseQuery string, args []any, f event.EventFilter) (string, []any, error) {
	var (
		conditions []string
	)

	if f.Status != nil {
		conditions = append(conditions, "e.status = ?")
		args = append(args, *f.Status)
	}

	if f.Category != nil {
		conditions = append(conditions, "e.category = ?")
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
