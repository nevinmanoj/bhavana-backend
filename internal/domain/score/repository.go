package score

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type ScoreWriteRepository interface {
	ScoreReadRepository
	CreateScore(ctx context.Context, db sqlx.ExtContext, scoreToCreate *Score) error
	UpdateScore(ctx context.Context, db sqlx.ExtContext, scoreToUpdate *Score) error
	DeleteScore(ctx context.Context, db sqlx.ExtContext, scoreID int64) error
}
type ScoreReadRepository interface {
	GetAllScores(ctx context.Context, db sqlx.ExtContext, filter ScoreFilter) ([]Score, error)
	GetScoresByEventID(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]EventScoreRow, error)
	GetScoreByID(ctx context.Context, db sqlx.ExtContext, id int64) (*Score, error)
}
