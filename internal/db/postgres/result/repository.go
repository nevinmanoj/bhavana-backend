package result

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/result"
)

type resultRepository struct {
}

func NewResultWriteRepository() result.ResultWriteRepository {
	return &resultRepository{}
}
func NewResultReadRepository() result.ResultReadRepository {
	return &resultRepository{}
}

// GetTeamTotals aggregates every team's score into a single total: the sum,
// over criteria, of the average score across judges — the same aggregate
// score.GetEventScoresDetailed already shows in the UI's Scores tab.
// has_scores distinguishes "no judge ever scored this team" (excluded from
// ranking) from a legitimate total of 0.00.
func (r *resultRepository) GetTeamTotals(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]result.TeamTotal, error) {
	totals := []result.TeamTotal{}
	query := `
		SELECT t.id AS team_id, t.school_id, t.chest_number,
		       COALESCE(SUM(ca.avg_score), 0) AS total_score,
		       COUNT(ca.team_id) > 0          AS has_scores
		FROM teams t
		LEFT JOIN (
		    SELECT s.team_id, s.criteria_id, AVG(s.score) AS avg_score
		    FROM scores s
		    JOIN event_criteria ec ON ec.id = s.criteria_id
		    WHERE ec.event_id = $1
		    GROUP BY s.team_id, s.criteria_id
		) ca ON ca.team_id = t.id
		WHERE t.event_id = $1
		GROUP BY t.id, t.school_id, t.chest_number
		ORDER BY has_scores DESC, total_score DESC, t.chest_number ASC
	`
	err := sqlx.SelectContext(ctx, db, &totals, query, eventID)
	if err != nil {
		return nil, result.ErrInternal
	}
	return totals, nil
}

func (r *resultRepository) GetStandingsMap(ctx context.Context, db sqlx.ExtContext, eventID int64) (map[int64]float64, error) {
	type row struct {
		Position int64   `db:"position"`
		Points   float64 `db:"points"`
	}
	rows := []row{}
	err := sqlx.SelectContext(ctx, db, &rows,
		`SELECT position, points FROM event_standings WHERE event_id = $1`,
		eventID,
	)
	if err != nil {
		return nil, result.ErrInternal
	}
	m := make(map[int64]float64, len(rows))
	for _, rw := range rows {
		m[rw.Position] = rw.Points
	}
	return m, nil
}

func (r *resultRepository) HasStandings(ctx context.Context, db sqlx.ExtContext, eventID int64) (bool, error) {
	var count int64
	err := sqlx.GetContext(ctx, db, &count,
		`SELECT COUNT(*) FROM event_standings WHERE event_id = $1`,
		eventID,
	)
	if err != nil {
		return false, result.ErrInternal
	}
	return count > 0, nil
}

func (r *resultRepository) CreateResults(ctx context.Context, db sqlx.ExtContext, results []result.Result) error {
	query := `
		INSERT INTO results (event_id, team_id, school_id, position, points, total_score)
		VALUES (:event_id, :team_id, :school_id, :position, :points, :total_score)
	`
	_, err := sqlx.NamedExecContext(ctx, db, query, results)
	if err != nil {
		return errorMapper(err)
	}
	return nil
}

func (r *resultRepository) GetResultsByEventID(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]result.EventResult, error) {
	results := []result.EventResult{}
	query := `
		SELECT r.team_id, t.chest_number, r.school_id, sc.name AS school_name,
		       r.total_score, r.position, r.points
		FROM results r
		JOIN teams t     ON t.id = r.team_id
		JOIN schools sc  ON sc.id = r.school_id
		WHERE r.event_id = $1
		ORDER BY r.position ASC NULLS LAST, t.chest_number ASC
	`
	err := sqlx.SelectContext(ctx, db, &results, query, eventID)
	if err != nil {
		return nil, result.ErrInternal
	}
	return results, nil
}

func (r *resultRepository) GetLeaderboard(ctx context.Context, db sqlx.ExtContext, filter result.LeaderboardFilter) ([]result.LeaderboardRow, error) {
	rows := []result.LeaderboardRow{}
	query := `
		SELECT sc.id AS school_id, sc.name AS school_name, sc.address AS school_address,
		       e.id AS event_id, e.title AS event_name,
		       r.team_id, t.chest_number, r.position, r.points
		FROM results r
		JOIN events  e  ON e.id  = r.event_id
		JOIN schools sc ON sc.id = r.school_id
		JOIN teams   t  ON t.id  = r.team_id
		WHERE r.points IS NOT NULL
	`
	args := []any{}
	if filter.Category != nil {
		query += ` AND e.category = $1`
		args = append(args, *filter.Category)
	}
	query += ` ORDER BY sc.id, e.id, r.position`

	err := sqlx.SelectContext(ctx, db, &rows, query, args...)
	if err != nil {
		return nil, result.ErrInternal
	}
	return rows, nil
}

// GetUnscoredTeams and GetJudgeGaps power the pre-finalize readiness check.
func (r *resultRepository) GetUnscoredTeams(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]result.UnscoredTeam, error) {
	teams := []result.UnscoredTeam{}
	query := `
		SELECT t.id AS team_id, t.chest_number, sc.name AS school_name
		FROM teams t
		JOIN schools sc ON sc.id = t.school_id
		WHERE t.event_id = $1
		  AND NOT EXISTS (
		      SELECT 1 FROM scores s
		      JOIN event_criteria ec ON ec.id = s.criteria_id
		      WHERE s.team_id = t.id AND ec.event_id = $1
		  )
		ORDER BY t.chest_number ASC
	`
	err := sqlx.SelectContext(ctx, db, &teams, query, eventID)
	if err != nil {
		return nil, result.ErrInternal
	}
	return teams, nil
}

func (r *resultRepository) GetJudgeGaps(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]result.JudgeGap, error) {
	gaps := []result.JudgeGap{}
	query := `
		SELECT u.id, u.name,
		       (SELECT COUNT(*) FROM teams WHERE event_id = $1)
		     * (SELECT COUNT(*) FROM event_criteria WHERE event_id = $1)
		     - COUNT(s.id) AS missing
		FROM event_judges ej
		JOIN users u ON u.id = ej.user_id
		LEFT JOIN scores s ON s.judge_id = ej.user_id
		    AND s.team_id IN (SELECT id FROM teams WHERE event_id = $1)
		WHERE ej.event_id = $1
		GROUP BY u.id, u.name
		HAVING COUNT(s.id) < (SELECT COUNT(*) FROM teams WHERE event_id = $1)
		                   * (SELECT COUNT(*) FROM event_criteria WHERE event_id = $1)
		ORDER BY u.name ASC
	`
	err := sqlx.SelectContext(ctx, db, &gaps, query, eventID)
	if err != nil {
		return nil, result.ErrInternal
	}
	return gaps, nil
}

func (r *resultRepository) GetTotalTeamsCount(ctx context.Context, db sqlx.ExtContext, eventID int64) (int64, error) {
	var count int64
	err := sqlx.GetContext(ctx, db, &count, `SELECT COUNT(*) FROM teams WHERE event_id = $1`, eventID)
	if err != nil {
		return 0, result.ErrInternal
	}
	return count, nil
}
