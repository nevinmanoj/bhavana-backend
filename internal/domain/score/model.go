package score

import (
	"time"
)

type Score struct {
	ID         int64     `db:"id"`
	EntryID    int64     `db:"entry_id"`
	JudgeID    int64     `db:"judge_id"`
	CriteriaID int64     `db:"criteria_id"`
	Score      float64   `db:"score"`
	CreatedAt  time.Time `db:"created_at"`
}
type EventScoreRow struct {
	EntryID       int64    `db:"entry_id"`
	ChestNumber   *int64   `db:"chest_number"`
	SchoolName    *string  `db:"school_name"`
	CriteriaID    int64    `db:"criteria_id"`
	CriteriaTitle string   `db:"criteria_title"`
	MaxScore      float64  `db:"max_score"`
	JudgeID       *int64   `db:"judge_id"`
	JudgeName     *string  `db:"judge_name"`
	Score         *float64 `db:"score"`
	ScoreID       *int64   `db:"score_id"`
}

// repsonse models
type EventScoresDetailed struct {
	EventID  int64             `json:"event_id"`
	Criteria []CriteriaSummary `json:"criteria"`
	Entries  []*EntryScore     `json:"entries"`
}
type CriteriaSummary struct {
	ID       int64   `json:"id"`
	Title    string  `json:"title"`
	MaxScore float64 `json:"max_score"`
}

type EntryScore struct {
	ID          int64                   `json:"id"`
	ChestNumber *int64                  `json:"chest_number"`
	School      string                  `json:"school,omitempty"`
	Scores      map[int64]CriteriaScore `json:"scores"`
	Total       float64                 `json:"total"`
	EntryTotal  float64                 `json:"entry_total,omitempty"`
}

type CriteriaScore struct {
	Avg    float64      `json:"avg"`
	Judges []JudgeScore `json:"judges"`
}

type JudgeScore struct {
	ScoreID   int64   `json:"score_id"`
	JudgeID   int64   `json:"judge_id"`
	JudgeName string  `json:"judge_name"`
	Score     float64 `json:"score"`
}
