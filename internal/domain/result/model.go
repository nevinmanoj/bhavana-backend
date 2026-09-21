package result

import (
	"time"
)

// Result is a single persisted row of the immutable results snapshot,
// written once per team when an event is finalized.
type Result struct {
	ID         int64     `db:"id"`
	EventID    int64     `db:"event_id"`
	TeamID     int64     `db:"team_id"`
	SchoolID   int64     `db:"school_id"`
	Position   *int64    `db:"position"` // NULL = unscored, excluded from ranking
	Points     *float64  `db:"points"`   // NULL = unscored, or no mapping for this position
	TotalScore float64   `db:"total_score"`
	CreatedAt  time.Time `db:"created_at"`
}

// TeamTotal is a team's aggregate score for an event, the input to Rank.
// HasScores distinguishes "never scored" (excluded from ranking) from a
// legitimate total of 0.00 (every judge scored it 0, which IS ranked).
type TeamTotal struct {
	TeamID      int64   `db:"team_id"`
	SchoolID    int64   `db:"school_id"`
	ChestNumber *int64  `db:"chest_number"`
	TotalScore  float64 `db:"total_score"`
	HasScores   bool    `db:"has_scores"`
}

// response / aggregate models

// EventResult is one row of a finalized event's results, joined with team
// and school identity for display. Returned directly as the HTTP response
// body (same convention as score.EventScoresDetailed), hence the json tags
// alongside the db tags used to scan the join.
type EventResult struct {
	TeamID      int64    `db:"team_id" json:"team_id"`
	ChestNumber *int64   `db:"chest_number" json:"chest_number"`
	SchoolID    int64    `db:"school_id" json:"school_id"`
	SchoolName  string   `db:"school_name" json:"school_name"`
	TotalScore  float64  `db:"total_score" json:"total_score"`
	Position    *int64   `db:"position" json:"position"`
	Points      *float64 `db:"points" json:"points"`
}

// LeaderboardRow is one team's mapped result joined with its school and
// event identity, the raw material the service groups into per-school rows.
type LeaderboardRow struct {
	SchoolID      int64   `db:"school_id"`
	SchoolName    string  `db:"school_name"`
	SchoolAddress string  `db:"school_address"`
	EventID       int64   `db:"event_id"`
	EventName     string  `db:"event_name"`
	TeamID        int64   `db:"team_id"`
	ChestNumber   *int64  `db:"chest_number"`
	Position      int64   `db:"position"`
	Points        float64 `db:"points"`
}

// LeaderboardEventBreakdown is one team's placing within one event, nested
// under its school in the leaderboard response.
type LeaderboardEventBreakdown struct {
	EventID     int64   `json:"event_id"`
	EventName   string  `json:"event_name"`
	TeamID      int64   `json:"team_id"`
	ChestNumber *int64  `json:"chest_number"`
	Position    int64   `json:"position"`
	Points      float64 `json:"points"`
}

// LeaderboardSchoolScore is one school's row on the leaderboard: its total
// points across every event it placed in, and the per-team breakdown.
type LeaderboardSchoolScore struct {
	SchoolID        int64                       `json:"school_id"`
	SchoolName      string                      `json:"school_name"`
	SchoolAddress   string                      `json:"school_address"`
	TotalPoints     float64                     `json:"total_points"`
	EventCount      int64                       `json:"event_count"`
	EventBreakdowns []LeaderboardEventBreakdown `json:"event_breakdowns"`
}

// UnscoredTeam is a team with zero score rows for the event, used by the
// finalize-readiness check.
type UnscoredTeam struct {
	TeamID      int64  `db:"team_id" json:"team_id"`
	ChestNumber *int64 `db:"chest_number" json:"chest_number"`
	SchoolName  string `db:"school_name" json:"school_name"`
}

// JudgeGap is a judge assigned to the event who has not yet scored every
// team x criteria combination.
type JudgeGap struct {
	UserID  int64  `db:"id" json:"user_id"`
	Name    string `db:"name" json:"name"`
	Missing int64  `db:"missing" json:"missing"`
}

// FinalizeReadiness summarises whether an event is ready to finalize. It is
// advisory only — the service never blocks finalize on judging gaps, only
// on HasStandings being false (enforced by the event service, not here).
type FinalizeReadiness struct {
	TotalTeams    int64          `json:"total_teams"`
	UnscoredTeams []UnscoredTeam `json:"unscored_teams"`
	JudgeGaps     []JudgeGap     `json:"judge_gaps"`
	HasStandings  bool           `json:"has_standings"`
}
