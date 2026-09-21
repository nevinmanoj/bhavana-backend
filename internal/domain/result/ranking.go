package result

import "math"

// round2 rounds to 2 decimal places, matching the NUMERIC(10,2) precision
// scores are stored at. Comparing rounded totals (rather than raw floats) is
// what makes ties detectable: averaging produces repeating decimals that
// would otherwise almost never compare exactly equal.
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// Rank computes each entry's position and points from its aggregate score.
// It is a pure function: no DB, no context, so it is unit-testable on its
// own (see ranking_test.go).
//
// entries should already be sorted has_scores DESC, total_score DESC (the
// repository's GetEntryTotals query does this), but Rank does not depend on
// that beyond scored entries needing to be contiguous by equal total to
// dense-rank correctly — it checks HasScores explicitly on every element.
//
// Entries with HasScores == false are excluded from ranking entirely: they
// get Position == nil, Points == nil, so they can never tie with each other
// or with an entry that was genuinely scored 0 by every judge.
//
// Ranking is dense: tied entries (equal rounded total) share a position and
// each receives that position's full points; the next distinct total takes
// the next position number, with no gap.
func Rank(entries []EntryTotal, standings map[int64]float64) []Result {
	results := make([]Result, 0, len(entries))

	var position int64
	var prevTotal float64
	havePrev := false

	for _, t := range entries {
		if !t.HasScores {
			results = append(results, Result{
				EntryID:    t.EntryID,
				SchoolID:   t.SchoolID,
				TotalScore: 0,
				Position:   nil,
				Points:     nil,
			})
			continue
		}

		total := round2(t.TotalScore)
		if !havePrev || total != prevTotal {
			position++
			prevTotal = total
			havePrev = true
		}

		pos := position
		var points *float64
		if p, ok := standings[pos]; ok {
			pointsVal := p
			points = &pointsVal
		}

		results = append(results, Result{
			EntryID:    t.EntryID,
			SchoolID:   t.SchoolID,
			TotalScore: total,
			Position:   &pos,
			Points:     points,
		})
	}

	return results
}
