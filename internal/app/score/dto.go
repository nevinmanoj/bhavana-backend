package score

import (
	"time"

	"github.com/nevinmanoj/bhavana-backend/internal/domain/score"
)

type CreateScoreRequest struct {
	TeamID  int64 `json:"team_id" validate:"required"`
	JudgeID int64 `json:"judge_id"`
	Scores  []ScoreRequest
}
type ScoreRequest struct {
	CriteriaID int64   `json:"criteria_id" validate:"required"`
	Score      float64 `json:"score" validate:"required"`
}
type UpdateScoresRequest struct {
	Scores []UpdateScoreRequest `json:"scores" validate:"required,dive"`
}
type UpdateScoreRequest struct {
	ID    int64   `json:"id" validate:"required"`
	Score float64 `json:"score" validate:"required,min=0"`
}

type ScoreResponse struct {
	ID         int64     `json:"id"`
	TeamID     int64     `json:"team_id"`
	JudgeID    int64     `json:"judge_id"`
	CriteriaID int64     `json:"criteria_id"`
	Score      float64   `json:"score"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateUpdateScoreResponse struct {
	Scores []ScoreResponse `json:"scores"`
}

func ToScoreResponseResponse(sc *score.Score) ScoreResponse {
	return ScoreResponse{
		ID:         sc.ID,
		TeamID:     sc.TeamID,
		JudgeID:    sc.JudgeID,
		CriteriaID: sc.CriteriaID,
		Score:      sc.Score,
		CreatedAt:  sc.CreatedAt,
	}
}

func ToCreateUpdateScoreResponse(scores []score.Score) CreateUpdateScoreResponse {
	arr := []ScoreResponse{}
	for _, sc := range scores {
		arr = append(arr, ToScoreResponseResponse(&sc))
	}
	return CreateUpdateScoreResponse{
		Scores: arr,
	}
}
