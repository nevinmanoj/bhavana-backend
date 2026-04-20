package score

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/access"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

type ScoreService interface {
	GetScoretByID(ctx context.Context, id int64) (*Score, error)
	GetEventScoresDetailed(ctx context.Context, eventID int64) (*EventScoresDetailed, error)
	CreateScores(ctx context.Context, scoresToCreate []Score) error
	UpdateScores(ctx context.Context, scoresToUpdate []Score) error
	DeleteScore(ctx context.Context, eventID int64) error
}

type scoreService struct {
	db            *sqlx.DB
	accessService access.AccessService
	repo          ScoreWriteRepository
}

func NewScoreService(db *sqlx.DB, accessService access.AccessService, repo ScoreWriteRepository) ScoreService {
	return &scoreService{db: db, accessService: accessService, repo: repo}
}

func (s *scoreService) GetScoretByID(ctx context.Context, id int64) (*Score, error) {
	return s.repo.GetScoreByID(ctx, s.db, id)
}

func (s *scoreService) GetEventScoresDetailed(ctx context.Context, eventID int64) (*EventScoresDetailed, error) {
	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)

	isJudge := scope.UserID != nil && role == rbac.UserRoleJudge

	rows, err := s.repo.GetScoresByEventID(ctx, s.db, eventID)
	if err != nil {
		return nil, err
	}

	criteriaOrder := []int64{}
	teamOrder := []int64{}
	seenCriteria := map[int64]bool{}
	seenTeams := map[int64]bool{}

	criteriaMap := map[int64]CriteriaSummary{}
	teamMap := map[int64]*TeamScore{}

	for _, row := range rows {
		// criteria
		if !seenCriteria[row.CriteriaID] {
			seenCriteria[row.CriteriaID] = true
			criteriaOrder = append(criteriaOrder, row.CriteriaID)
			criteriaMap[row.CriteriaID] = CriteriaSummary{
				ID:       row.CriteriaID,
				Title:    row.CriteriaTitle,
				MaxScore: row.MaxScore,
			}
		}

		// teams
		if !seenTeams[row.TeamID] {
			seenTeams[row.TeamID] = true
			teamOrder = append(teamOrder, row.TeamID)

			t := &TeamScore{
				ID:          row.TeamID,
				ChestNumber: row.ChestNumber,
				Scores:      map[int64]CriteriaScore{},
			}
			if !isJudge && row.SchoolName != nil {
				t.School = *row.SchoolName
			}
			teamMap[row.TeamID] = t
		}

		// scores
		if row.Score != nil && row.JudgeID != nil {
			t := teamMap[row.TeamID]
			cs := t.Scores[row.CriteriaID]
			cs.Judges = append(cs.Judges, JudgeScore{
				JudgeID:   *row.JudgeID,
				JudgeName: *row.JudgeName,
				Score:     *row.Score,
			})
			t.Scores[row.CriteriaID] = cs
		}
	}

	// compute averages and totals
	for _, t := range teamMap {
		var teamTotal float64
		for id, cs := range t.Scores {
			if len(cs.Judges) == 0 {
				continue
			}
			var sum float64
			for _, j := range cs.Judges {
				sum += j.Score
			}
			cs.Avg = sum / float64(len(cs.Judges))
			t.Scores[id] = cs
			teamTotal += cs.Avg
		}
		t.Total = teamTotal
		if !isJudge {
			t.TeamTotal = teamTotal
		}
	}

	// assemble ordered slices
	criteriaSlice := make([]CriteriaSummary, 0, len(criteriaOrder))
	for _, id := range criteriaOrder {
		criteriaSlice = append(criteriaSlice, criteriaMap[id])
	}

	teamSlice := make([]*TeamScore, 0, len(teamOrder))
	for _, id := range teamOrder {
		teamSlice = append(teamSlice, teamMap[id])
	}

	return &EventScoresDetailed{
		EventID:  eventID,
		Criteria: criteriaSlice,
		Teams:    teamSlice,
	}, nil
}
func (s *scoreService) CreateScores(ctx context.Context, scoresToCreate []Score) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()
	for _, scoreToCreate := range scoresToCreate {
		err := s.repo.CreateScore(ctx, tx, &scoreToCreate)
		if err != nil {
			return mapScoreError(err)
		}
	}
	return tx.Commit()
}
func (s *scoreService) UpdateScores(ctx context.Context, scoresToUpdate []Score) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()
	for _, scoreToUpdate := range scoresToUpdate {
		access, err := s.accessService.CanModifyScore(ctx, scoreToUpdate.ID)
		if err != nil {
			return err
		}
		if !access {
			return ErrUnauthorized
		}
		err = s.repo.UpdateScore(ctx, tx, &scoreToUpdate)
		if err != nil {
			return mapScoreError(err)
		}
	}
	return tx.Commit()
}
func (s *scoreService) DeleteScore(ctx context.Context, scoreID int64) error {
	return s.repo.DeleteScore(ctx, s.db, scoreID)
}
