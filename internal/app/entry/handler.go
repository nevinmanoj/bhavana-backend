package entry

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/entry"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
	"github.com/nevinmanoj/bhavana-backend/internal/util"
)

type EntryHandler struct {
	service   entry.EntryService
	validator *validator.Validate
}

func NewEntryHandler(s entry.EntryService, v *validator.Validate) *EntryHandler {
	return &EntryHandler{service: s, validator: v}
}

func (h *EntryHandler) GetEntries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	var resp any
	q := r.URL.Query()
	filter, errresp := parseEntryFilter(q)
	if errresp != nil {
		json.NewEncoder(w).Encode(errresp)
		return
	}

	entries, err := h.service.GetEntries(ctx, filter)
	if err != nil {
		resp = GetEntryDomainErrorResponse(err)
	} else {
		role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
		switch role {
		case rbac.UserRoleJudge:
			resp = GetAllResponsePage[EntryResponseJudge]{
				StatusCode: 200,
				Message:    "Entries fetched successfully",
				Data:       mapEntries(entries, ToEntryResponseJudge),
			}
		default:
			resp = GetAllResponsePage[EntryFullResponse]{
				StatusCode: 200,
				Message:    "Entries fetched successfully",
				Data:       mapEntries(entries, ToEntryFullResponse),
			}
		}
	}
	json.NewEncoder(w).Encode(resp)
}
func (h *EntryHandler) GetEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	entryIdStr := chi.URLParam(r, "entryId")
	w.Header().Set("Content-Type", "application/json")
	var resp any
	entryId, err := strconv.ParseInt(entryIdStr, 10, 64)
	if err != nil {
		resp = ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid entry ID in URL parameter",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	result, err := h.service.GetEntryByID(ctx, entryId)
	if err != nil {
		resp = GetEntryDomainErrorResponse(err)
	} else {
		role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
		switch role {
		case rbac.UserRoleJudge:
			resp = GetResponsePage[EntryResponseJudge]{
				StatusCode: 200,
				Message:    "Entries fetched successfully",
				Data:       ToEntryResponseJudge(result),
			}
		default:
			resp = GetResponsePage[EntryFullResponse]{
				StatusCode: 200,
				Message:    "Entries fetched successfully",
				Data:       ToEntryFullResponse(result),
			}
		}

	}

	json.NewEncoder(w).Encode(resp)
}
func (h *EntryHandler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateEntryRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body " + err.Error(),
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
	entryToCreate := entry.EntryFull{
		Entry: entry.Entry{
			EventID:  req.EventID,
			SchoolID: req.SchoolID,
		},
		Members: parseEntryMemberReqs(req.Members),
	}
	err := h.service.CreateEntry(ctx, &entryToCreate)
	if err != nil {
		json.NewEncoder(w).Encode(GetEntryDomainErrorResponse(err))
		return
	}
	entryResponse := ToEntryFullResponse(&entryToCreate)
	json.NewEncoder(w).Encode(PostResponsePage[EntryFullResponse]{
		Message:    "Entry created successfully",
		Data:       entryResponse,
		StatusCode: http.StatusCreated,
	})
}
func (h *EntryHandler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req UpdateEntryRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body " + err.Error(),
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
	entryToUpdate := entry.EntryFull{
		Entry: entry.Entry{
			ID:       req.ID,
			EventID:  req.EventID,
			SchoolID: req.SchoolID,
		},
		Members: parseEntryMemberReqs(req.Members),
	}

	entryIdStr := chi.URLParam(r, "entryId")
	entryId, err := util.ParseStrToInt64(entryIdStr)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid entry ID in URL parameter",
		})
		return
	}
	if req.ID != *entryId {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "entry ID in request body does not match URL parameter",
		})
		return
	}
	err = h.service.UpdateEntry(ctx, &entryToUpdate)
	if err != nil {
		json.NewEncoder(w).Encode(GetEntryDomainErrorResponse(err))
		return
	}
	entryResponse := ToEntryFullResponse(&entryToUpdate)
	json.NewEncoder(w).Encode(PostResponsePage[EntryFullResponse]{
		Message:    "Entry updated successfully",
		Data:       entryResponse,
		StatusCode: http.StatusCreated,
	})
}
func (h *EntryHandler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	entryIdstr := chi.URLParam(r, "entryId")
	w.Header().Set("Content-Type", "application/json")
	var resp any
	entryID, err := strconv.ParseInt(entryIdstr, 10, 64)
	if err != nil {
		resp = ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid entry ID in URL parameter",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	err = h.service.DeleteEntry(ctx, entryID)
	if err != nil {
		resp = GetEntryDomainErrorResponse(err)
	} else {
		resp = DeleteResponsePage{
			StatusCode: http.StatusNoContent,
			Message:    "Entry deleted successfully",
		}
	}

	json.NewEncoder(w).Encode(resp)
}

// helpers
func parseEntryMemberReqs(membersRequests []EntryMemberRequest) []entry.EntryMember {
	members := make([]entry.EntryMember, len(membersRequests))
	for i, membersRequest := range membersRequests {
		members[i] = entry.EntryMember{
			StudentID: membersRequest.StudentID,
		}
	}
	return members
}
func mapEntries[T any](entries []entry.EntryFull, mapper func(*entry.EntryFull) T) []T {
	result := make([]T, len(entries))
	for i, t := range entries {
		result[i] = mapper(&t)
	}
	return result
}
