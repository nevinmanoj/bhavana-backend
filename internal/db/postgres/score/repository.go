package score

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/score"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

type scoreRepository struct {
}

func NewScoreWriteRepository() score.ScoreWriteRepository {
	return &scoreRepository{}
}
func NewScoreReadRepository() score.ScoreReadRepository {
	return &scoreRepository{}
}

func (r *scoreRepository) GetAllScores(ctx context.Context, db sqlx.ExtContext, filter score.ScoreFilter) ([]score.Score, error) {
	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	scores := []score.Score{}
	baseQuery := `SELECT s.* FROM scores s`
	args := []any{}

	if scope.UserID != nil && role == rbac.UserRoleJudge {
		filter.JudgeID = scope.UserID
	}

	finalQuery, finalArgs, err := buildScoreQuery(baseQuery, args, filter)
	err = sqlx.SelectContext(
		ctx, db,
		&scores,
		finalQuery, finalArgs...,
	)
	if err != nil {
		return nil, err
	}
	return scores, nil
}

func (r *scoreRepository) GetScoresByEventID(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]score.EventScoreRow, error) {
	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)

	isJudge := scope.UserID != nil && role == rbac.UserRoleJudge

	args := []any{eventID}

	query := `
		SELECT
			t.id            AS team_id,
			t.chest_number,
			ec.id           AS criteria_id,
			ec.title        AS criteria_title,
			ec.max_score,
			s.judge_id,
			u.name          AS judge_name,
			s.score
	`

	if !isJudge {
		query += `, sc.name AS school_name `
	}

	query += `
		FROM teams t
		JOIN event_criteria ec   ON ec.event_id = t.event_id
		LEFT JOIN scores s       ON s.team_id = t.id
								AND s.criteria_id = ec.id
		LEFT JOIN users u        ON u.id = s.judge_id
	`

	if !isJudge {
		query += ` JOIN schools sc ON sc.id = t.school_id `
	}

	query += ` WHERE t.event_id = $1 `

	if isJudge {
		args = append(args, scope.UserID)
		query += ` AND s.judge_id = $2 `
	}

	query += ` ORDER BY t.id, ec.id, s.judge_id `
	scores := []score.EventScoreRow{}
	err := sqlx.SelectContext(
		ctx, db,
		&scores,
		query, args...,
	)
	if err != nil {
		return nil, err
	}
	return scores, nil
}
func (s *scoreRepository) GetScoreByID(ctx context.Context, db sqlx.ExtContext, id int64) (*score.Score, error) {
	var scoreFromDB score.Score
	baseQuery := `SELECT * FROM scores WHERE id = $1`
	args := []any{id}

	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)

	if role == rbac.UserRoleJudge && scope.UserID != nil {
		baseQuery += ` AND judge_id = $2`
		args = append(args, scope.UserID)
	}

	err := sqlx.GetContext(ctx, db, &scoreFromDB, baseQuery, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, score.ErrScoreNotFound
		}
		return nil, err
	}
	return &scoreFromDB, nil
}
func (e *scoreRepository) CreateScore(ctx context.Context, db sqlx.ExtContext, scoreToCreate *score.Score) error {
	query := `
		INSERT INTO scores (
			team_id,
			judge_id,
			criteria_id,
			score
		)
		VALUES (
			:team_id,
			:judge_id,
			:criteria_id,
			:score
		)
		RETURNING id, created_at
	`

	rows, err := sqlx.NamedQueryContext(ctx, db, query, scoreToCreate)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&scoreToCreate.ID, &scoreToCreate.CreatedAt)
		return nil
	}

	return sql.ErrNoRows
}
func (s *scoreRepository) UpdateScore(ctx context.Context, db sqlx.ExtContext, scoreToUpdate *score.Score) error {
	query := `
		UPDATE scores
		SET score = :score
		WHERE id = :id
	`
	result, err := sqlx.NamedExecContext(ctx, db, query, scoreToUpdate)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return score.ErrScoreNotFound
	}
	return nil
}
func (r *scoreRepository) DeleteScore(ctx context.Context, db sqlx.ExtContext, scoreID int64) error {
	query := `
		DELETE FROM scores
		WHERE id = $1
	`
	_, err := db.ExecContext(ctx, query, scoreID)
	if err != nil {
		return err
	}
	return nil
}
