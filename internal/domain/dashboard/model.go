package dashboard

import (
	"errors"
	"time"

	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/result"
)

var ErrInternal = errors.New("internal error")

// Totals are the headline counts across the whole competition.
type Totals struct {
	Schools  int64 `db:"schools" json:"schools"`
	Students int64 `db:"students" json:"students"`
	Events   int64 `db:"events" json:"events"`
	Entries  int64 `db:"entries" json:"entries"`
}

// SchoolInReview is a registered school still waiting for an admin to accept it.
type SchoolInReview struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// IncompleteSchool is an accepted school that has added no students, or has
// students but no entries yet.
type IncompleteSchool struct {
	ID       int64  `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Students int64  `db:"students" json:"students"`
	Entries  int64  `db:"entries" json:"entries"`
}

type StatusCount struct {
	Status core.EventStatus `db:"status" json:"status"`
	Count  int64            `db:"count" json:"count"`
}

// EventFacts is the raw per-event material the service derives the setup and
// scoring sections from. Never returned to the client.
type EventFacts struct {
	ID           int64            `db:"id"`
	Title        string           `db:"title"`
	Status       core.EventStatus `db:"status"`
	Judges       int64            `db:"judges"`
	Criteria     int64            `db:"criteria"`
	Entries      int64            `db:"entries"`
	HasStandings bool             `db:"has_standings"`
}

// JudgeScored is how many score rows one assigned judge has entered for one
// event that is open or closed.
type JudgeScored struct {
	EventID int64  `db:"event_id"`
	UserID  int64  `db:"id"`
	Name    string `db:"name"`
	Scored  int64  `db:"scored"`
}

// EventSetup is an event heading towards judging that still lacks judges,
// criteria or a points table.
type EventSetup struct {
	ID           int64            `json:"id"`
	Title        string           `json:"title"`
	Status       core.EventStatus `json:"status"`
	HasJudges    bool             `json:"has_judges"`
	HasCriteria  bool             `json:"has_criteria"`
	HasStandings bool             `json:"has_standings"`
}

// EventScoring is the judging progress of an open or closed event. Expected is
// entries x criteria x assigned judges; Entered counts only scores from
// currently assigned judges, so it never exceeds Expected.
type EventScoring struct {
	EventID      int64             `json:"event_id"`
	Title        string            `json:"title"`
	Status       core.EventStatus  `json:"status"`
	Entries      int64             `json:"entries"`
	Expected     int64             `json:"expected"`
	Entered      int64             `json:"entered"`
	HasStandings bool              `json:"has_standings"`
	JudgeGaps    []result.JudgeGap `json:"judge_gaps"`
}

type Dashboard struct {
	Totals             Totals             `json:"totals"`
	SchoolsInReview    []SchoolInReview   `json:"schools_in_review"`
	IncompleteSchools  []IncompleteSchool `json:"incomplete_schools"`
	EventsByStatus     []StatusCount      `json:"events_by_status"`
	EventsNeedingSetup []EventSetup       `json:"events_needing_setup"`
	Scoring            []EventScoring     `json:"scoring"`
}
