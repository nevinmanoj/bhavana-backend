package entry

import (
	"testing"

	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/event"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

func soloEvent() *event.Event {
	return &event.Event{MinMembers: 1, MaxMembers: 1, MaxEntriesPerSchool: 10}
}

func groupEvent() *event.Event {
	return &event.Event{MinMembers: 2, MaxMembers: 4, MaxEntriesPerSchool: 10}
}

func batch(memberGroups ...[]int64) []*EntryFull {
	entries := make([]*EntryFull, len(memberGroups))
	for i, group := range memberGroups {
		members := make([]EntryMember, len(group))
		for j, studentID := range group {
			members[j] = EntryMember{StudentID: studentID}
		}
		entries[i] = &EntryFull{Members: members}
	}
	return entries
}

func TestValidateBatch(t *testing.T) {
	tests := []struct {
		name    string
		event   *event.Event
		entries []*EntryFull
		want    error
	}{
		{
			name:    "solo batch, one member each",
			event:   soloEvent(),
			entries: batch([]int64{1}, []int64{2}, []int64{3}),
			want:    nil,
		},
		{
			name:    "group batch within range",
			event:   groupEvent(),
			entries: batch([]int64{1, 2}, []int64{3, 4, 5, 6}),
			want:    nil,
		},
		{
			name:    "empty batch",
			event:   soloEvent(),
			entries: batch(),
			want:    ErrEmptyBatch,
		},
		{
			name:    "an entry over max members",
			event:   soloEvent(),
			entries: batch([]int64{1}, []int64{2, 3}),
			want:    ErrMemberCountExceedsLimit,
		},
		{
			name:    "an entry under min members",
			event:   groupEvent(),
			entries: batch([]int64{1, 2}, []int64{3}),
			want:    ErrMemberCountBelowMinimum,
		},
		{
			name:    "same student twice across the batch",
			event:   soloEvent(),
			entries: batch([]int64{1}, []int64{2}, []int64{1}),
			want:    ErrStudentAlreadyInEntry,
		},
		{
			name:    "same student twice inside one group entry",
			event:   groupEvent(),
			entries: batch([]int64{7, 7}),
			want:    ErrStudentAlreadyInEntry,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateBatch(tt.event, tt.entries); got != tt.want {
				t.Errorf("validateBatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntryRegistrationAllowed(t *testing.T) {
	// school admins register during the registration window only; admins may
	// add late entries right up to the end of scoring.
	tests := []struct {
		status      core.EventStatus
		admin       bool
		schoolAdmin bool
	}{
		{core.EventStatusDraft, false, false},
		{core.EventStatusRegistrationOpen, true, true},
		{core.EventStatusRegistrationClosed, true, false},
		{core.EventStatusPreparing, true, false},
		{core.EventStatusOpen, true, false},
		{core.EventStatusClosed, true, false},
		{core.EventStatusFinalized, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := entryRegistrationAllowed(rbac.UserRoleAdmin, tt.status); got != tt.admin {
				t.Errorf("admin in %s = %v, want %v", tt.status, got, tt.admin)
			}
			if got := entryRegistrationAllowed(rbac.UserRoleSchoolAdmin, tt.status); got != tt.schoolAdmin {
				t.Errorf("school_admin in %s = %v, want %v", tt.status, got, tt.schoolAdmin)
			}
			if got := entryRegistrationAllowed(rbac.UserRoleJudge, tt.status); got != tt.schoolAdmin {
				t.Errorf("judge in %s = %v, want %v", tt.status, got, tt.schoolAdmin)
			}
		})
	}
}

func TestReconcileEntryUpdate(t *testing.T) {
	chest := 7
	stored := func() *Entry {
		return &Entry{ID: 5, EventID: 10, SchoolID: 20, ChestNumber: &chest}
	}
	members := []EntryMember{{StudentID: 1}, {StudentID: 2}}

	tests := []struct {
		name     string
		eventID  int64
		schoolID int64
		want     error
	}{
		{"matching event and school", 10, 20, nil},
		{"body names another event", 11, 20, ErrInvalidEntryUpdate},
		{"body names another school", 10, 21, ErrInvalidEntryUpdate},
		{"body names neither", 0, 0, ErrInvalidEntryUpdate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &EntryFull{
				Entry:   Entry{ID: 5, EventID: tt.eventID, SchoolID: tt.schoolID},
				Members: members,
			}
			err := reconcileEntryUpdate(stored(), req)
			if err != tt.want {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
			if err != nil {
				return
			}
			if req.ChestNumber == nil || *req.ChestNumber != chest {
				t.Errorf("chest number not taken from the stored entry: %v", req.ChestNumber)
			}
			if len(req.Members) != len(members) {
				t.Errorf("members were not preserved from the request: %v", req.Members)
			}
		})
	}
}
