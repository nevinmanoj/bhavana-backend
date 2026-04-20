package score

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/app/errmap"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/score"
)

type ScoreHandler struct {
	service   score.ScoreService
	validator *validator.Validate
}

func NewSchoolHandler(s score.ScoreService, v *validator.Validate) *ScoreHandler {
	return &ScoreHandler{service: s, validator: v}
}

func (h *ScoreHandler) GetScoresByEventID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventIdStr := chi.URLParam(r, "eventId")
	log.Println("GetScoresByEventID::Fetching all score for event" + eventIdStr)
	w.Header().Set("Content-Type", "application/json")
	var resp any
	eventID, err := strconv.ParseInt(eventIdStr, 10, 64)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
		json.NewEncoder(w).Encode(resp)
		return
	}

	scores, err := h.service.GetEventScoresDetailed(ctx, eventID)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
	} else {
		resp = GetResponsePage[score.EventScoresDetailed]{
			StatusCode: 200,
			Message:    "Scores fetched successfully for event " + eventIdStr,
			Data:       *scores,
		}
	}
	json.NewEncoder(w).Encode(resp)
}
func (h *ScoreHandler) GetScore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Println("GetScore::Fetching score with ID:", chi.URLParam(r, "scoreId"))
	w.Header().Set("Content-Type", "application/json")
	var resp any
	scoreIdStr := chi.URLParam(r, "scoreId")
	scoreId, err := strconv.ParseInt(scoreIdStr, 10, 64)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
		json.NewEncoder(w).Encode(resp)
		return
	}
	scoreFromDB, err := h.service.GetScoretByID(ctx, scoreId)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
	} else {
		resp = GetResponsePage[score.Score]{
			StatusCode: 200,
			Message:    "Score fetched successfully",
			Data:       *scoreFromDB,
		}
	}
	json.NewEncoder(w).Encode(resp)
}
func (h *ScoreHandler) CreateScores(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateScoreRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body",
		})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	scoresToCreate := parseCreateScoreReq(req.Scores, req.TeamID, req.JudgeID)
	err := h.service.CreateScores(ctx, scoresToCreate)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	scoresResponse := ToCreateUpdateScoreResponse(scoresToCreate)
	json.NewEncoder(w).Encode(PostResponsePage[CreateUpdateScoreResponse]{
		Message:    "Scores created successfully",
		Data:       scoresResponse,
		StatusCode: http.StatusCreated,
	})
}
func (h *ScoreHandler) UpdateScores(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Println("UpdateScores::Updating scores")
	var req UpdateScoresRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body",
		})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	scores := parseUpdateScoreReq(req.Scores)

	err := h.service.UpdateScores(ctx, scores)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	scoresResponses := ToCreateUpdateScoreResponse(scores)
	json.NewEncoder(w).Encode(PutResponsePage[CreateUpdateScoreResponse]{
		Message:    "Scores were updated successfully",
		Data:       scoresResponses,
		StatusCode: http.StatusOK,
	})
}
func (h *ScoreHandler) DeleteScore(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	scoreIdStr := chi.URLParam(r, "scoreId")
	log.Println("DeleteScore::Deleting score with ID:", scoreIdStr)
	w.Header().Set("Content-Type", "application/json")
	var resp any
	scoreId, err := strconv.ParseInt(scoreIdStr, 10, 64)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
		json.NewEncoder(w).Encode(resp)
		return
	}
	err = h.service.DeleteScore(ctx, scoreId)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
	} else {
		resp = DeleteResponsePage{
			StatusCode: http.StatusNoContent,
			Message:    "Score deleted successfully",
		}
	}

	json.NewEncoder(w).Encode(resp)
}

// helpers
func parseCreateScoreReq(scoreReqs []ScoreRequest, teamID, judgeID int64) []score.Score {
	scores := make([]score.Score, len(scoreReqs))
	for i, scoreReq := range scoreReqs {
		scores[i] = score.Score{
			JudgeID:    judgeID,
			TeamID:     teamID,
			CriteriaID: scoreReq.CriteriaID,
			Score:      scoreReq.Score,
		}
	}
	return scores
}
func parseUpdateScoreReq(scoreReqs []UpdateScoreRequest) []score.Score {
	scores := make([]score.Score, len(scoreReqs))
	for i, scoreReq := range scoreReqs {
		scores[i] = score.Score{
			ID:    scoreReq.ID,
			Score: scoreReq.Score,
		}
	}
	return scores
}
