package event

import (
	"testing"

	"github.com/nevinmanoj/bhavana-backend/internal/core"
)

// PUT /events/{id} must never move an event's status: finalizing that way would
// skip the preconditions and the results computation in UpdateEventStatus.
func TestCheckStatusUnchanged(t *testing.T) {
	tests := []struct {
		name      string
		stored    core.EventStatus
		requested core.EventStatus
		wantErr   error
	}{
		{"same status in draft", core.EventStatusDraft, core.EventStatusDraft, nil},
		{"same status in preparing", core.EventStatusPreparing, core.EventStatusPreparing, nil},
		{"finalize from closed", core.EventStatusClosed, core.EventStatusFinalized, ErrStatusChangeNotAllowed},
		{"forward move", core.EventStatusDraft, core.EventStatusRegistrationOpen, ErrStatusChangeNotAllowed},
		{"back to draft", core.EventStatusOpen, core.EventStatusDraft, ErrStatusChangeNotAllowed},
		{"reopen from closed", core.EventStatusClosed, core.EventStatusOpen, ErrStatusChangeNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkStatusUnchanged(tt.stored, tt.requested); got != tt.wantErr {
				t.Errorf("checkStatusUnchanged(%s, %s) = %v, want %v", tt.stored, tt.requested, got, tt.wantErr)
			}
		})
	}
}

// canEditEventSetup gates judge removal, and criteria/standings add-edit-delete,
// in syncEventJudges/syncEventCriterias/syncEventStandings. It must mirror the
// DRAFT/PREPARING window that check_event_judge_modification,
// check_event_criteria_modification and check_event_standing_modification
// enforce in the DB - and the frontend's constants/eventStatuses.ts helper of
// the same name.
func TestCanEditEventSetup(t *testing.T) {
	tests := []struct {
		status core.EventStatus
		want   bool
	}{
		{core.EventStatusDraft, true},
		{core.EventStatusPreparing, true},
		{core.EventStatusRegistrationOpen, false},
		{core.EventStatusRegistrationClosed, false},
		{core.EventStatusOpen, false},
		{core.EventStatusClosed, false},
		{core.EventStatusFinalized, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := canEditEventSetup(tt.status); got != tt.want {
				t.Errorf("canEditEventSetup(%s) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
