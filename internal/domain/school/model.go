package school

import (
	"time"

	"github.com/nevinmanoj/bhavana-backend/internal/core"
)

type School struct {
	ID           int64     `db:"id"`
	Address      string    `db:"address"`
	Name         string    `db:"name"`
	ContactName  string    `db:"contact_name"`
	ContactEmail string    `db:"contact_email"`
	ContactPhone string    `db:"contact_phone"`
	CreatedAt    time.Time `db:"created_at"`
}

type Student struct {
	ID        int64         `db:"id"`
	SchoolID  int64         `db:"school_id"`
	Name      string        `db:"name"`
	Age       int           `db:"age"`
	Category  core.Category `db:"category"`
	CreatedAt time.Time     `db:"created_at"`
}
