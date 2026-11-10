package school

import (
	"testing"

	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

func id(n int64) *int64 { return &n }

func TestStudentsLocked(t *testing.T) {
	tests := []struct {
		name string
		role rbac.UserRole
		open bool
		want bool
	}{
		{"school admin, closed", rbac.UserRoleSchoolAdmin, false, true},
		{"school admin, open", rbac.UserRoleSchoolAdmin, true, false},
		{"admin, closed", rbac.UserRoleAdmin, false, false},
		{"admin, open", rbac.UserRoleAdmin, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := studentsLocked(tt.role, &Settings{StudentRegistrationOpen: tt.open}); got != tt.want {
				t.Errorf("studentsLocked() = %v, want %v", got, tt.want)
			}
		})
	}
}

// While in review, a contact email edit is also the admin's login change, so
// acceptance emails the corrected address. Never otherwise.
func TestSyncsAdminEmail(t *testing.T) {
	locked := func(status core.SchoolStatus, admin *int64) *LockedSchool {
		return &LockedSchool{SchoolAdmin: admin, Status: status, ContactEmail: "old@example.com"}
	}
	updated := func(admin *int64, email string) *School {
		return &School{SchoolAdmin: admin, ContactEmail: email}
	}

	tests := []struct {
		name    string
		locked  *LockedSchool
		updated *School
		want    bool
	}{
		{"in review, email changed", locked(core.SchoolStatusInReview, id(7)), updated(id(7), "new@example.com"), true},
		{"in review, email unchanged", locked(core.SchoolStatusInReview, id(7)), updated(id(7), "old@example.com"), false},
		{"accepted, email changed", locked(core.SchoolStatusAccepted, id(7)), updated(id(7), "new@example.com"), false},
		{"in review, admin reassigned", locked(core.SchoolStatusInReview, id(7)), updated(id(8), "new@example.com"), false},
		{"in review, admin unlinked", locked(core.SchoolStatusInReview, id(7)), updated(nil, "new@example.com"), false},
		{"in review, no admin", locked(core.SchoolStatusInReview, nil), updated(nil, "new@example.com"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := syncsAdminEmail(tt.locked, tt.updated); got != tt.want {
				t.Errorf("syncsAdminEmail = %v, want %v", got, tt.want)
			}
		})
	}
}

// After acceptance, a school with an admin changes its email only through
// SendLoginDetails, so the sign-in email can't silently drift from it.
func TestContactEmailLocked(t *testing.T) {
	locked := func(status core.SchoolStatus, admin *int64) *LockedSchool {
		return &LockedSchool{SchoolAdmin: admin, Status: status, ContactEmail: "old@example.com"}
	}

	tests := []struct {
		name   string
		locked *LockedSchool
		email  string
		want   bool
	}{
		{"accepted with admin, email changed", locked(core.SchoolStatusAccepted, id(7)), "new@example.com", true},
		{"accepted with admin, email unchanged", locked(core.SchoolStatusAccepted, id(7)), "old@example.com", false},
		{"accepted without admin", locked(core.SchoolStatusAccepted, nil), "new@example.com", false},
		{"in review with admin", locked(core.SchoolStatusInReview, id(7)), "new@example.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contactEmailLocked(tt.locked, &School{ContactEmail: tt.email}); got != tt.want {
				t.Errorf("contactEmailLocked = %v, want %v", got, tt.want)
			}
		})
	}
}

// A school admin edits only the school they're linked to, whatever its status.
func TestOwnsSchool(t *testing.T) {
	tests := []struct {
		name   string
		admin  *int64
		userID int64
		want   bool
	}{
		{"linked admin", id(7), 7, true},
		{"another school's admin", id(7), 8, false},
		{"school with no admin", nil, 7, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ownsSchool(&LockedSchool{SchoolAdmin: tt.admin}, tt.userID); got != tt.want {
				t.Errorf("ownsSchool = %v, want %v", got, tt.want)
			}
		})
	}
}
