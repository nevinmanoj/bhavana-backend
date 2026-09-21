package result

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type ResultWriteRepository interface {
	ResultReadRepository
	// CreateResults bulk-inserts the ranked snapshot for an event. Called once,
	// inside the finalize transaction; the DB triggers enforce that the event
	// already reads 'finalized' and that results are never updated afterwards.
	CreateResults(ctx context.Context, db sqlx.ExtContext, results []Result) error
}

type ResultReadRepository interface {
	// GetEntryTotals returns every entry's aggregate score for the event, in
	// ranking order (has_scores DESC, total_score DESC, chest_number ASC).
	GetEntryTotals(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]EntryTotal, error)
	// GetStandingsMap returns the event's position -> points mapping.
	GetStandingsMap(ctx context.Context, db sqlx.ExtContext, eventID int64) (map[int64]float64, error)
	// HasStandings reports whether the event has at least one standing defined.
	HasStandings(ctx context.Context, db sqlx.ExtContext, eventID int64) (bool, error)

	GetResultsByEventID(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]EventResult, error)
	GetLeaderboard(ctx context.Context, db sqlx.ExtContext, filter LeaderboardFilter) ([]LeaderboardRow, error)

	GetUnscoredEntries(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]UnscoredEntry, error)
	GetJudgeGaps(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]JudgeGap, error)
	GetTotalEntriesCount(ctx context.Context, db sqlx.ExtContext, eventID int64) (int64, error)
}
