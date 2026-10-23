package score

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
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
	// GetScoreEventStatus returns the status of the event a score belongs to.
	GetScoreEventStatus(ctx context.Context, db sqlx.ExtContext, scoreID int64) (core.EventStatus, error)
}
