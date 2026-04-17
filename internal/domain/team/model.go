package team

import (
	"time"

	"github.com/nevinmanoj/bhavana-backend/internal/core"
)

type TeamFull struct {
	Team          `db:"team"`
	SchoolName    string        `db:"school_name"`
	SchoolAddress string        `db:"school_address"`
	EventTitle    string        `db:"event_title"`
	Category      core.Category `db:"category"`
	Members       []TeamMember  `db:"-"`
}

type Team struct {
	ID          int64     `db:"id"`
	EventID     int64     `db:"event_id"`
	SchoolID    int64     `db:"school_id"`
	ChestNumber int       `db:"chest_number"`
	CreatedAt   time.Time `db:"created_at"`
}

type TeamMember struct {
	Name      string    `db:"name"`
	TeamID    int64     `db:"team_id"`
	StudentID int64     `db:"student_id"`
	CreatedAt time.Time `db:"created_at"`
}
