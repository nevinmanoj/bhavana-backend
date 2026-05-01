package team

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/app/errmap"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/team"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
	"github.com/nevinmanoj/bhavana-backend/internal/util"
)

type TeamHandler struct {
	service   team.TeamService
	validator *validator.Validate
}

func NewTeamHandler(s team.TeamService, v *validator.Validate) *TeamHandler {
	return &TeamHandler{service: s, validator: v}
}

func (h *TeamHandler) GetTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	var resp any
	q := r.URL.Query()
	filter, errresp := parseTeamFilter(q)
	if errresp != nil {
		json.NewEncoder(w).Encode(errresp)
		return
	}

	teams, err := h.service.GetTeams(ctx, filter)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
	} else {
		role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
		switch role {
		case rbac.UserRoleJudge:
			resp = GetAllResponsePage[TeamResponseJudge]{
				StatusCode: 200,
				Message:    "Teams fetched successfully",
				Data:       mapTeams(teams, ToTeamResponseJudge),
			}
		default:
			resp = GetAllResponsePage[TeamFullResponse]{
				StatusCode: 200,
				Message:    "Teams fetched successfully",
				Data:       mapTeams(teams, ToTeamFullResponse),
			}
		}
	}
	json.NewEncoder(w).Encode(resp)
}
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamIdStr := chi.URLParam(r, "teamId")
	w.Header().Set("Content-Type", "application/json")
	var resp any
	teamId, err := strconv.ParseInt(teamIdStr, 10, 64)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
		json.NewEncoder(w).Encode(resp)
		return
	}
	result, err := h.service.GetTeamsByID(ctx, teamId)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
	} else {
		role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
		switch role {
		case rbac.UserRoleJudge:
			resp = GetResponsePage[TeamResponseJudge]{
				StatusCode: 200,
				Message:    "Teams fetched successfully",
				Data:       ToTeamResponseJudge(result),
			}
		default:
			resp = GetResponsePage[TeamFullResponse]{
				StatusCode: 200,
				Message:    "Teams fetched successfully",
				Data:       ToTeamFullResponse(result),
			}
		}

	}

	json.NewEncoder(w).Encode(resp)
}
func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateTeamRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body0 " + err.Error(),
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
	teamToCreate := team.TeamFull{
		Team: team.Team{
			EventID:  req.EventID,
			SchoolID: req.SchoolID,
		},
		Members: pareseTeamMemberReqs(req.Members),
	}
	err := h.service.CreateTeam(ctx, &teamToCreate)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	teamResponse := ToTeamFullResponse(&teamToCreate)
	json.NewEncoder(w).Encode(PostResponsePage[TeamFullResponse]{
		Message:    "Team created successfully",
		Data:       teamResponse,
		StatusCode: http.StatusCreated,
	})
}
func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req UpdateTeamRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body" + err.Error(),
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
	teamToUpdate := team.TeamFull{
		Team: team.Team{
			ID:       req.ID,
			EventID:  req.EventID,
			SchoolID: req.SchoolID,
		},
		Members: pareseTeamMemberReqs(req.Members),
	}

	teamIdStr := chi.URLParam(r, "teamId")
	teamId, err := util.ParseStrToInt64(teamIdStr)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid team ID in URL parameter",
		})
		return
	}
	if req.ID != *teamId {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "team ID in request body does not match URL parameter",
		})
		return
	}
	err = h.service.UpdateTeam(ctx, &teamToUpdate)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	teamResponse := ToTeamFullResponse(&teamToUpdate)
	json.NewEncoder(w).Encode(PostResponsePage[TeamFullResponse]{
		Message:    "Team updated successfully",
		Data:       teamResponse,
		StatusCode: http.StatusCreated,
	})
}
func (h *TeamHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teamIdstr := chi.URLParam(r, "teamId")
	w.Header().Set("Content-Type", "application/json")
	var resp any
	teamID, err := strconv.ParseInt(teamIdstr, 10, 64)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
		json.NewEncoder(w).Encode(resp)
		return
	}
	err = h.service.DeleteTeam(ctx, teamID)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
	} else {
		resp = DeleteResponsePage{
			StatusCode: http.StatusNoContent,
			Message:    "Team deleted successfully",
		}
	}

	json.NewEncoder(w).Encode(resp)
}

// helpers
func pareseTeamMemberReqs(membersRequests []TeamMemberRequest) []team.TeamMember {
	members := make([]team.TeamMember, len(membersRequests))
	for i, membersRequest := range membersRequests {
		members[i] = team.TeamMember{
			StudentID: membersRequest.StudentID,
		}
	}
	return members
}
func mapTeams[T any](teams []team.TeamFull, mapper func(*team.TeamFull) T) []T {
	result := make([]T, len(teams))
	for i, t := range teams {
		result[i] = mapper(&t)
	}
	return result
}
