package dashboard

import (
	"reflect"
	"testing"

	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/result"
)

func TestOrderByLifecycle(t *testing.T) {
	got := OrderByLifecycle([]StatusCount{
		{Status: core.EventStatusFinalized, Count: 2},
		{Status: core.EventStatusDraft, Count: 1},
		{Status: core.EventStatusOpen, Count: 0},
	})
	want := []StatusCount{
		{Status: core.EventStatusDraft, Count: 1},
		{Status: core.EventStatusFinalized, Count: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestEventsNeedingSetup(t *testing.T) {
	facts := []EventFacts{
		{ID: 1, Title: "Ready", Status: core.EventStatusPreparing, Judges: 2, Criteria: 3, HasStandings: true},
		{ID: 2, Title: "No judges", Status: core.EventStatusRegistrationOpen, Criteria: 3, HasStandings: true},
		{ID: 3, Title: "No points table", Status: core.EventStatusOpen, Judges: 1, Criteria: 1},
		// Closed events report their blockers through scoring, not setup.
		{ID: 4, Title: "Closed, nothing set", Status: core.EventStatusClosed},
	}
	got := EventsNeedingSetup(facts)
	want := []EventSetup{
		{ID: 2, Title: "No judges", Status: core.EventStatusRegistrationOpen, HasCriteria: true, HasStandings: true},
		{ID: 3, Title: "No points table", Status: core.EventStatusOpen, HasJudges: true, HasCriteria: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestScoring(t *testing.T) {
	facts := []EventFacts{
		{ID: 1, Title: "Open", Status: core.EventStatusOpen, Judges: 2, Criteria: 3, Entries: 4, HasStandings: true},
		{ID: 2, Title: "Preparing", Status: core.EventStatusPreparing, Judges: 2, Criteria: 3, Entries: 4},
		{ID: 3, Title: "Closed, no judges", Status: core.EventStatusClosed, Criteria: 3, Entries: 4},
	}
	judgeScored := []JudgeScored{
		{EventID: 1, UserID: 10, Name: "Anu", Scored: 12},
		{EventID: 1, UserID: 11, Name: "Biju", Scored: 5},
	}
	got := Scoring(facts, judgeScored)
	want := []EventScoring{
		{
			EventID: 1, Title: "Open", Status: core.EventStatusOpen,
			Entries: 4, Expected: 24, Entered: 17, HasStandings: true,
			JudgeGaps: []result.JudgeGap{{UserID: 11, Name: "Biju", Missing: 7}},
		},
		{
			EventID: 3, Title: "Closed, no judges", Status: core.EventStatusClosed,
			Entries: 4, JudgeGaps: []result.JudgeGap{},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// A judge's count can exceed entries x criteria if criteria were removed after
// scoring; entered must still never exceed expected.
func TestScoringClampsEntered(t *testing.T) {
	facts := []EventFacts{{ID: 1, Status: core.EventStatusOpen, Judges: 1, Criteria: 1, Entries: 2}}
	got := Scoring(facts, []JudgeScored{{EventID: 1, UserID: 10, Name: "Anu", Scored: 5}})
	if got[0].Entered != 2 || got[0].Expected != 2 || len(got[0].JudgeGaps) != 0 {
		t.Errorf("got %+v", got[0])
	}
}
