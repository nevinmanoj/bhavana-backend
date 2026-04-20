package score

import (
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/score"
)

func buildScoreQuery(baseQuery string, args []any, f score.ScoreFilter) (string, []any, error) {
	var (
		conditions []string
	)

	if f.TeamID != nil {
		conditions = append(conditions, "s.team_id = ?")
		args = append(args, *f.EventID)
	}
	if f.EventID != nil {
		baseQuery += " JOIN event_criteria ec ON ec.id = s.criteria_id"
		conditions = append(conditions, "ec.event_id = ?")
		args = append(args, *f.EventID)
	}
	if f.JudgeID != nil {
		conditions = append(conditions, "s.judge_id = ?")
		args = append(args, *f.JudgeID)
	}

	// Apply WHERE
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Ordering (always deterministic)
	baseQuery += " ORDER BY e.created_at DESC"

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
