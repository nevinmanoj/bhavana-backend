package school

import (
	"time"

	"github.com/nevinmanoj/bhavana-backend/internal/core"
)

// Settings is the single row of admin-controlled switches.
type Settings struct {
	SchoolRegistrationOpen  bool `db:"school_registration_open"`
	StudentRegistrationOpen bool `db:"student_registration_open"`
}

type School struct {
	ID              int64             `db:"id"`
	SchoolAdmin     *int64            `db:"school_admin"`
	Address         string            `db:"address"`
	Name            string            `db:"name"`
	ContactName     string            `db:"contact_name"`
	ContactEmail    string            `db:"contact_email"`
	ContactPhone    string            `db:"contact_phone"`
	Status          core.SchoolStatus `db:"status"`
	PrincipalName   string            `db:"principal_name"`
	PrincipalPhone  string            `db:"principal_phone"`
	ParticipantsMC  int               `db:"participants_mc"`
	ParticipantsPC  int               `db:"participants_pc"`
	ParticipantsHI  int               `db:"participants_hi"`
	NonParticipants int               `db:"non_participants"`
	TotalStaff      int               `db:"total_staff"`
	TotalParents    int               `db:"total_parents"`
	CreatedAt       time.Time         `db:"created_at"`
}

type Student struct {
	ID        int64         `db:"id"`
	SchoolID  int64         `db:"school_id"`
	Name      string        `db:"name"`
	Age       int           `db:"age"`
	Category  core.Category `db:"category"`
	CreatedAt time.Time     `db:"created_at"`
}
