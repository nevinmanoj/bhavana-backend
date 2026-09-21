package result

import "testing"

func TestRank_Empty(t *testing.T) {
	got := Rank(nil, map[int64]float64{1: 10})
	if len(got) != 0 {
		t.Fatalf("expected 0 results, got %d", len(got))
	}
}

func TestRank_SimpleOrdering(t *testing.T) {
	teams := []TeamTotal{
		{TeamID: 1, TotalScore: 90, HasScores: true},
		{TeamID: 2, TotalScore: 80, HasScores: true},
		{TeamID: 3, TotalScore: 70, HasScores: true},
	}
	standings := map[int64]float64{1: 10, 2: 7, 3: 5}

	got := Rank(teams, standings)

	want := []struct {
		teamID   int64
		position int64
		points   float64
	}{
		{1, 1, 10},
		{2, 2, 7},
		{3, 3, 5},
	}
	for i, w := range want {
		r := got[i]
		if r.Position == nil || *r.Position != w.position {
			t.Fatalf("team %d: expected position %d, got %v", w.teamID, w.position, r.Position)
		}
		if r.Points == nil || *r.Points != w.points {
			t.Fatalf("team %d: expected points %v, got %v", w.teamID, w.points, r.Points)
		}
	}
}

func TestRank_TieSharesPositionAndFullPoints(t *testing.T) {
	// Two teams tie for 1st (both 90.00): both get position 1 and the FULL
	// 1st-place points (not split). The next distinct score takes position 2,
	// not 3 — dense ranking, no gap.
	teams := []TeamTotal{
		{TeamID: 1, TotalScore: 90, HasScores: true},
		{TeamID: 2, TotalScore: 90, HasScores: true},
		{TeamID: 3, TotalScore: 80, HasScores: true},
	}
	standings := map[int64]float64{1: 10, 2: 7, 3: 5}

	got := Rank(teams, standings)

	for _, r := range got[:2] {
		if r.Position == nil || *r.Position != 1 {
			t.Fatalf("tied team %d: expected position 1, got %v", r.TeamID, r.Position)
		}
		if r.Points == nil || *r.Points != 10 {
			t.Fatalf("tied team %d: expected full points 10, got %v", r.TeamID, r.Points)
		}
	}
	third := got[2]
	if third.Position == nil || *third.Position != 2 {
		t.Fatalf("team after tie: expected position 2 (dense rank, no skip), got %v", third.Position)
	}
	if third.Points == nil || *third.Points != 7 {
		t.Fatalf("team after tie: expected points 7, got %v", third.Points)
	}
}

func TestRank_UnmappedPositionGetsNilPoints(t *testing.T) {
	teams := []TeamTotal{
		{TeamID: 1, TotalScore: 90, HasScores: true},
		{TeamID: 2, TotalScore: 80, HasScores: true},
	}
	// only position 1 has a mapping
	standings := map[int64]float64{1: 10}

	got := Rank(teams, standings)

	if got[1].Position == nil || *got[1].Position != 2 {
		t.Fatalf("expected team 2 ranked at position 2, got %v", got[1].Position)
	}
	if got[1].Points != nil {
		t.Fatalf("expected nil points for unmapped position 2, got %v", *got[1].Points)
	}
}

func TestRank_UnscoredTeamsExcludedFromRanking(t *testing.T) {
	// Two unscored teams must NOT tie with each other for a position, and
	// must NOT tie with a team that was genuinely scored 0.00 by every judge.
	teams := []TeamTotal{
		{TeamID: 1, TotalScore: 90, HasScores: true},
		{TeamID: 2, TotalScore: 0, HasScores: true}, // scored zero by every judge — still ranked
		{TeamID: 3, TotalScore: 0, HasScores: false},
		{TeamID: 4, TotalScore: 0, HasScores: false},
	}
	standings := map[int64]float64{1: 10, 2: 7}

	got := Rank(teams, standings)

	// team 1: position 1
	if got[0].Position == nil || *got[0].Position != 1 {
		t.Fatalf("team 1: expected position 1, got %v", got[0].Position)
	}
	// team 2: scored zero, but IS ranked — position 2, mapped points 7
	if got[1].Position == nil || *got[1].Position != 2 {
		t.Fatalf("team 2 (scored zero): expected position 2, got %v", got[1].Position)
	}
	if got[1].Points == nil || *got[1].Points != 7 {
		t.Fatalf("team 2 (scored zero): expected points 7, got %v", got[1].Points)
	}
	// teams 3 and 4: unscored, excluded from ranking entirely
	for _, r := range got[2:] {
		if r.Position != nil {
			t.Fatalf("unscored team %d: expected nil position, got %v", r.TeamID, *r.Position)
		}
		if r.Points != nil {
			t.Fatalf("unscored team %d: expected nil points, got %v", r.TeamID, *r.Points)
		}
	}
}

func TestRank_RoundingMakesNearEqualTotalsTie(t *testing.T) {
	// Averaging across judges produces repeating decimals; two totals that
	// land in the same cent (both round to 81.00) must tie even though the
	// raw floats differ, rather than drifting apart on float noise.
	teams := []TeamTotal{
		{TeamID: 1, TotalScore: 81.001, HasScores: true},
		{TeamID: 2, TotalScore: 81.004, HasScores: true},
	}
	got := Rank(teams, map[int64]float64{1: 10})

	if *got[0].Position != *got[1].Position {
		t.Fatalf("expected both teams tied at the same position after rounding, got %v and %v", *got[0].Position, *got[1].Position)
	}
}
