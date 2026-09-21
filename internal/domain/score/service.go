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
	entryOrder := []int64{}
	seenCriteria := map[int64]bool{}
	seenEntries := map[int64]bool{}

	criteriaMap := map[int64]CriteriaSummary{}
	entryMap := map[int64]*EntryScore{}

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

		// entries
		if !seenEntries[row.EntryID] {
			seenEntries[row.EntryID] = true
			entryOrder = append(entryOrder, row.EntryID)

			t := &EntryScore{
				ID:          row.EntryID,
				ChestNumber: row.ChestNumber,
				Scores:      map[int64]CriteriaScore{},
			}
			if !isJudge && row.SchoolName != nil {
				t.School = *row.SchoolName
			}
			entryMap[row.EntryID] = t
		}

		// scores
		if row.Score != nil && row.JudgeID != nil {
			t := entryMap[row.EntryID]
			cs := t.Scores[row.CriteriaID]
			cs.Judges = append(cs.Judges, JudgeScore{
				ScoreID:   *row.ScoreID,
				JudgeID:   *row.JudgeID,
				JudgeName: *row.JudgeName,
				Score:     *row.Score,
			})
			t.Scores[row.CriteriaID] = cs
		}
	}

	// compute averages and totals
	for _, t := range entryMap {
		var entryTotal float64
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
			entryTotal += cs.Avg
		}
		t.Total = entryTotal
		if !isJudge {
			t.EntryTotal = entryTotal
		}
	}

	// assemble ordered slices
	criteriaSlice := make([]CriteriaSummary, 0, len(criteriaOrder))
	for _, id := range criteriaOrder {
		criteriaSlice = append(criteriaSlice, criteriaMap[id])
	}

	entrySlice := make([]*EntryScore, 0, len(entryOrder))
	for _, id := range entryOrder {
		entrySlice = append(entrySlice, entryMap[id])
	}

	return &EventScoresDetailed{
		EventID:  eventID,
		Criteria: criteriaSlice,
		Entries:  entrySlice,
	}, nil
}
func (s *scoreService) CreateScores(ctx context.Context, scoresToCreate []Score) error {
	judgeID := ctx.Value(middleware.ContextUserID).(int64)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()
	for _, scoreToCreate := range scoresToCreate {
		// judge_id always comes from the authenticated caller, never the request body,
		// so a judge cannot submit scores attributed to another judge.
		scoreToCreate.JudgeID = judgeID
		err := s.repo.CreateScore(ctx, tx, &scoreToCreate)
		if err != nil {
			return err
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
			return err
		}
	}
	return tx.Commit()
}
func (s *scoreService) DeleteScore(ctx context.Context, scoreID int64) error {
	access, err := s.accessService.CanModifyScore(ctx, scoreID)
	if err != nil {
		return err
	}
	if !access {
		return ErrUnauthorized
	}
	return s.repo.DeleteScore(ctx, s.db, scoreID)
}
