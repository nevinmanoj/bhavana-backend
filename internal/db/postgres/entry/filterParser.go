package entry

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/entry"
)

// appendEntryConditions turns the filter's own fields into WHERE conditions,
// shared between the data query (buildEntryQuery) and the count query
// (buildEntryCountQuery) so the two never drift out of sync with each other.
func appendEntryConditions(conditions []string, args []any, f entry.EntryFilter) ([]string, []any) {
	if f.SchoolID != nil {
		conditions = append(conditions, "t.school_id = ?")
		args = append(args, *f.SchoolID)
	}
	if f.EventID != nil {
		conditions = append(conditions, "t.event_id = ?")
		args = append(args, *f.EventID)
	}
	if f.Category != nil {
		// category lives on the joined events table (aliased e), not entries (t)
		conditions = append(conditions, "e.category = ?")
		args = append(args, *f.Category)
	}
	return conditions, args
}

func buildEntryQuery(baseQuery string, conditions []string, args []any, f entry.EntryFilter) (string, []any, error) {
	conditions, args = appendEntryConditions(conditions, args, f)

	// Apply WHERE
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Ordering (always deterministic)
	baseQuery += " ORDER BY t.created_at DESC"

	// Pagination is opt-in: Limit/Offset are 0 for every caller that never
	// asked for a page (see EntryFilter), so this stays a no-op for them.
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

// buildEntryCountQuery mirrors buildEntryQuery's WHERE clause exactly (via
// appendEntryConditions) but with no ORDER BY/LIMIT/OFFSET, for the total
// count a paginated list needs alongside its current page. baseQuery must
// already be a `SELECT COUNT(*) FROM entries t JOIN ...` with the same
// JOINs/scope conditions as the data query, so the two return counts that
// agree with each other.
func buildEntryCountQuery(baseQuery string, conditions []string, args []any, f entry.EntryFilter) (string, []any, error) {
	conditions, args = appendEntryConditions(conditions, args, f)

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
