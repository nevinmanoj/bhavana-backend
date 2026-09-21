package entry

import (
	"time"

	"github.com/nevinmanoj/bhavana-backend/internal/core"
)

type EntryFull struct {
	Entry         `db:"entry"`
	SchoolName    string        `db:"school_name"`
	SchoolAddress string        `db:"school_address"`
	EventTitle    string        `db:"event_title"`
	Category      core.Category `db:"category"`
	Members       []EntryMember `db:"-"`
}

type Entry struct {
	ID          int64     `db:"id"`
	EventID     int64     `db:"event_id"`
	SchoolID    int64     `db:"school_id"`
	ChestNumber int       `db:"chest_number"`
	CreatedAt   time.Time `db:"created_at"`
}

type EntryMember struct {
	Name      string    `db:"name"`
	EntryID   int64     `db:"entry_id"`
	StudentID int64     `db:"student_id"`
	CreatedAt time.Time `db:"created_at"`
}
