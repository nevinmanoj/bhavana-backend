package dashboard

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/result"
)

type DashboardService interface {
	GetDashboard(ctx context.Context) (*Dashboard, error)
}

type dashboardService struct {
	db   *sqlx.DB
	repo DashboardReadRepository
}

func NewDashboardService(db *sqlx.DB, repo DashboardReadRepository) DashboardService {
	return &dashboardService{db: db, repo: repo}
}

func (s *dashboardService) GetDashboard(ctx context.Context) (*Dashboard, error) {
	totals, err := s.repo.GetTotals(ctx, s.db)
	if err != nil {
		return nil, err
	}
	inReview, err := s.repo.GetSchoolsInReview(ctx, s.db)
	if err != nil {
		return nil, err
	}
	incomplete, err := s.repo.GetIncompleteSchools(ctx, s.db)
	if err != nil {
		return nil, err
	}
	byStatus, err := s.repo.GetEventsByStatus(ctx, s.db)
	if err != nil {
		return nil, err
	}
	facts, err := s.repo.GetEventFacts(ctx, s.db)
	if err != nil {
		return nil, err
	}
	judgeScored, err := s.repo.GetJudgeScored(ctx, s.db)
	if err != nil {
		return nil, err
	}

	return &Dashboard{
		Totals:             totals,
		SchoolsInReview:    inReview,
		IncompleteSchools:  incomplete,
		EventsByStatus:     OrderByLifecycle(byStatus),
		EventsNeedingSetup: EventsNeedingSetup(facts),
		Scoring:            Scoring(facts, judgeScored),
	}, nil
}

// OrderByLifecycle sorts status counts into lifecycle order and drops statuses
// with no events.
func OrderByLifecycle(counts []StatusCount) []StatusCount {
	byStatus := make(map[core.EventStatus]int64, len(counts))
	for _, c := range counts {
		byStatus[c.Status] = c.Count
	}
	ordered := []StatusCount{}
	for _, status := range core.AllEventStatuses {
		if n := byStatus[status]; n > 0 {
			ordered = append(ordered, StatusCount{Status: status, Count: n})
		}
	}
	return ordered
}

// EventsNeedingSetup picks the events between registration opening and judging
// that lack judges, criteria or a points table. Closed events are left out:
// their blockers are reported with their scoring instead.
func EventsNeedingSetup(facts []EventFacts) []EventSetup {
	setups := []EventSetup{}
	for _, f := range facts {
		switch f.Status {
		case core.EventStatusRegistrationOpen, core.EventStatusRegistrationClosed,
			core.EventStatusPreparing, core.EventStatusOpen:
		default:
			continue
		}
		if f.Judges > 0 && f.Criteria > 0 && f.HasStandings {
			continue
		}
		setups = append(setups, EventSetup{
			ID:           f.ID,
			Title:        f.Title,
			Status:       f.Status,
			HasJudges:    f.Judges > 0,
			HasCriteria:  f.Criteria > 0,
			HasStandings: f.HasStandings,
		})
	}
	return setups
}

// Scoring builds the judging progress of every open or closed event from its
// facts and its judges' entered-score counts.
func Scoring(facts []EventFacts, judgeScored []JudgeScored) []EventScoring {
	byEvent := map[int64][]JudgeScored{}
	for _, js := range judgeScored {
		byEvent[js.EventID] = append(byEvent[js.EventID], js)
	}

	scoring := []EventScoring{}
	for _, f := range facts {
		if f.Status != core.EventStatusOpen && f.Status != core.EventStatusClosed {
			continue
		}
		perJudge := f.Entries * f.Criteria
		row := EventScoring{
			EventID:      f.ID,
			Title:        f.Title,
			Status:       f.Status,
			Entries:      f.Entries,
			Expected:     perJudge * f.Judges,
			HasStandings: f.HasStandings,
			JudgeGaps:    []result.JudgeGap{},
		}
		for _, js := range byEvent[f.ID] {
			scored := min(js.Scored, perJudge)
			row.Entered += scored
			if scored < perJudge {
				row.JudgeGaps = append(row.JudgeGaps, result.JudgeGap{
					UserID:  js.UserID,
					Name:    js.Name,
					Missing: perJudge - scored,
				})
			}
		}
		scoring = append(scoring, row)
	}
	return scoring
}
