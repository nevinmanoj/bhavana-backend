package event

import (
	"testing"

	"github.com/nevinmanoj/bhavana-backend/internal/core"
)

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
