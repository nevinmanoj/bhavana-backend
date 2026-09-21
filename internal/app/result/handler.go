package result

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/app/errmap"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/result"
)

type ResultHandler struct {
	service   result.ResultService
	validator *validator.Validate
}

func NewResultHandler(s result.ResultService, v *validator.Validate) *ResultHandler {
	return &ResultHandler{service: s, validator: v}
}

func (h *ResultHandler) GetEventResults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventIdStr := chi.URLParam(r, "eventId")
	w.Header().Set("Content-Type", "application/json")
	var resp any
	eventID, err := strconv.ParseInt(eventIdStr, 10, 64)
	if err != nil {
		resp = ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid event ID",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	results, err := h.service.GetEventResults(ctx, eventID)
	if err != nil {
		resp = GetResultDomainErrorResponse(err)
	} else {
		resp = GetAllResponsePage[result.EventResult]{
			StatusCode: 200,
			Message:    "Results fetched successfully for event " + eventIdStr,
			Data:       results,
		}
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *ResultHandler) GetReadiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventIdStr := chi.URLParam(r, "eventId")
	w.Header().Set("Content-Type", "application/json")
	var resp any
	eventID, err := strconv.ParseInt(eventIdStr, 10, 64)
	if err != nil {
		resp = ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid event ID",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	readiness, err := h.service.GetReadiness(ctx, eventID)
	if err != nil {
		resp = GetResultDomainErrorResponse(err)
	} else {
		resp = GetResponsePage[result.FinalizeReadiness]{
			StatusCode: 200,
			Message:    "Finalize readiness fetched successfully for event " + eventIdStr,
			Data:       *readiness,
		}
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *ResultHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	var resp any
	q := r.URL.Query()

	filter, errresp := parseLeaderboardFilter(q)
	if errresp != nil {
		json.NewEncoder(w).Encode(errmap.GetHttpErrorResponse(errresp))
		return
	}

	leaderboard, err := h.service.GetLeaderboard(ctx, filter)
	if err != nil {
		resp = GetResultDomainErrorResponse(err)
	} else {
		resp = GetAllResponsePage[result.LeaderboardSchoolScore]{
			StatusCode: 200,
			Message:    "Leaderboard fetched successfully",
			Data:       leaderboard,
		}
	}
	json.NewEncoder(w).Encode(resp)
}
