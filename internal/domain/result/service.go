package result

import (
	"context"
	"sort"

	"github.com/jmoiron/sqlx"
)

type ResultService interface {
	// GenerateForEvent ranks every team in the event from its aggregate score,
	// maps each rank onto the event's standings, and bulk-inserts the
	// immutable results snapshot. Runs inside the caller's finalize
	// transaction. Satisfies event.ResultGenerator.
	GenerateForEvent(ctx context.Context, tx *sqlx.Tx, eventID int64) error

	GetEventResults(ctx context.Context, eventID int64) ([]EventResult, error)
	GetReadiness(ctx context.Context, eventID int64) (*FinalizeReadiness, error)
	GetLeaderboard(ctx context.Context, filter LeaderboardFilter) ([]LeaderboardSchoolScore, error)
}

type resultService struct {
	db   *sqlx.DB
	repo ResultWriteRepository
}

func NewResultService(db *sqlx.DB, repo ResultWriteRepository) ResultService {
	return &resultService{db: db, repo: repo}
}

func (s *resultService) GenerateForEvent(ctx context.Context, tx *sqlx.Tx, eventID int64) error {
	teams, err := s.repo.GetTeamTotals(ctx, tx, eventID)
	if err != nil {
		return err
	}
	standings, err := s.repo.GetStandingsMap(ctx, tx, eventID)
	if err != nil {
		return err
	}

	ranked := Rank(teams, standings)
	for i := range ranked {
		ranked[i].EventID = eventID
	}

	if len(ranked) == 0 {
		return nil
	}

	return s.repo.CreateResults(ctx, tx, ranked)
}

func (s *resultService) GetEventResults(ctx context.Context, eventID int64) ([]EventResult, error) {
	results, err := s.repo.GetResultsByEventID(ctx, s.db, eventID)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return []EventResult{}, nil
	}
	return results, nil
}

func (s *resultService) GetReadiness(ctx context.Context, eventID int64) (*FinalizeReadiness, error) {
	totalTeams, err := s.repo.GetTotalTeamsCount(ctx, s.db, eventID)
	if err != nil {
		return nil, err
	}
	unscored, err := s.repo.GetUnscoredTeams(ctx, s.db, eventID)
	if err != nil {
		return nil, err
	}
	judgeGaps, err := s.repo.GetJudgeGaps(ctx, s.db, eventID)
	if err != nil {
		return nil, err
	}
	hasStandings, err := s.repo.HasStandings(ctx, s.db, eventID)
	if err != nil {
		return nil, err
	}

	if unscored == nil {
		unscored = []UnscoredTeam{}
	}
	if judgeGaps == nil {
		judgeGaps = []JudgeGap{}
	}

	return &FinalizeReadiness{
		TotalTeams:    totalTeams,
		UnscoredTeams: unscored,
		JudgeGaps:     judgeGaps,
		HasStandings:  hasStandings,
	}, nil
}

func (s *resultService) GetLeaderboard(ctx context.Context, filter LeaderboardFilter) ([]LeaderboardSchoolScore, error) {
	rows, err := s.repo.GetLeaderboard(ctx, s.db, filter)
	if err != nil {
		return nil, err
	}

	order := []int64{}
	bySchool := map[int64]*LeaderboardSchoolScore{}
	eventSeen := map[int64]map[int64]bool{}

	for _, r := range rows {
		school, ok := bySchool[r.SchoolID]
		if !ok {
			school = &LeaderboardSchoolScore{
				SchoolID:      r.SchoolID,
				SchoolName:    r.SchoolName,
				SchoolAddress: r.SchoolAddress,
			}
			bySchool[r.SchoolID] = school
			eventSeen[r.SchoolID] = map[int64]bool{}
			order = append(order, r.SchoolID)
		}

		school.TotalPoints += r.Points
		school.EventBreakdowns = append(school.EventBreakdowns, LeaderboardEventBreakdown{
			EventID:     r.EventID,
			EventName:   r.EventName,
			TeamID:      r.TeamID,
			ChestNumber: r.ChestNumber,
			Position:    r.Position,
			Points:      r.Points,
		})
		if !eventSeen[r.SchoolID][r.EventID] {
			eventSeen[r.SchoolID][r.EventID] = true
			school.EventCount++
		}
	}

	schools := make([]LeaderboardSchoolScore, 0, len(order))
	for _, id := range order {
		schools = append(schools, *bySchool[id])
	}

	// The UI derives rank from array index and does no sorting of its own, so
	// the API must return schools pre-sorted: total points desc, name asc as
	// a deterministic tiebreaker.
	sort.SliceStable(schools, func(i, j int) bool {
		if schools[i].TotalPoints != schools[j].TotalPoints {
			return schools[i].TotalPoints > schools[j].TotalPoints
		}
		return schools[i].SchoolName < schools[j].SchoolName
	})

	return schools, nil
}
