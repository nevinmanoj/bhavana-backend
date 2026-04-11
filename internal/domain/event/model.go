package event

import (
	"time"

	core "github.com/nevinmanoj/bhavana-backend/internal/core"
)

type EventDetails struct {
	Event    Event
	Judges   []EventJudge
	Criteria []EventCriteria
}

type Event struct {
	ID                int64            `db:"id"`
	Title             string           `db:"title"`
	Description       string           `db:"description"`
	MinTeamSize       int64            `db:"min_team_size"`
	MaxTeamSize       int64            `db:"max_team_size"`
	MaxTeamsPerSchool int64            `db:"max_teams_per_school"`
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
