package event

import (
	"time"

	core "github.com/nevinmanoj/bhavana-backend/internal/core"
)

type EventDetails struct {
	Event     Event
	Judges    []EventJudge
	Criteria  []EventCriteria
	Standings []EventStanding
}

type Event struct {
	ID                int64            `db:"id"`
	Title             string           `db:"title"`
	Description       string           `db:"description"`
	MinMembers          int64          `db:"min_members"`
	MaxMembers          int64          `db:"max_members"`
	MaxEntriesPerSchool int64          `db:"max_entries_per_school"`
	Status            core.EventStatus `db:"status"`
	Category          core.Category    `db:"category"`
	CreatedAt         time.Time        `db:"created_at"`
}

type EventJudge struct {
	Name    string `db:"name"`
	EventID int64  `db:"event_id"`
	UserID  int64  `db:"user_id"`
}

type EventCriteria struct {
	ID        int64     `db:"id"`
	EventID   int64     `db:"event_id"`
	Title     string    `db:"title"`
	MaxScore  float64   `db:"max_score"`
	CreatedAt time.Time `db:"created_at"`
}

type EventStanding struct {
	ID        int64     `db:"id"`
	EventID   int64     `db:"event_id"`
	Position  int64     `db:"position"`
	Points    float64   `db:"points"`
	CreatedAt time.Time `db:"created_at"`
}
