package event

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/app/errmap"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	event "github.com/nevinmanoj/bhavana-backend/internal/domain/event"
	"github.com/nevinmanoj/bhavana-backend/internal/util"
)

type EventHandler struct {
	service   event.EventService
	validator *validator.Validate
}

func NewEventHandler(s event.EventService, v *validator.Validate) *EventHandler {
	return &EventHandler{service: s, validator: v}
}

func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.Println("HandlerGetEvents::Fetching all events")
	w.Header().Set("Content-Type", "application/json")
	var resp any
	q := r.URL.Query()

	filter, errresp := parseEventFilter(q)
	if errresp != nil {
		json.NewEncoder(w).Encode(errresp)
		return
	}
	events, err := h.service.GetAllEvents(ctx, filter)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
	} else {

		eventResponses := make([]EventResponse, len(events))
		for i, e := range events {
			eventResponses[i] = ToEventResponse(&e)
		}
		resp = GetAllResponsePage[EventResponse]{
			StatusCode: 200,
			Message:    "Events fetched successfully",
			Data:       eventResponses,
		}
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventIdStr := chi.URLParam(r, "eventId")
	log.Println("HandlerGetEvent::Fetching event with ID:", eventIdStr)
	w.Header().Set("Content-Type", "application/json")
	var resp any
	eventId, err := strconv.ParseInt(eventIdStr, 10, 64)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
		json.NewEncoder(w).Encode(resp)
		return
	}
	result, err := h.service.GetEventByID(ctx, eventId)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
	} else {
		eventResponse := ToEventDetailsResponse(result)
		resp = GetResponsePage[EventDetailsResponse]{
			StatusCode: 200,
			Message:    "Event fetched successfully",
			Data:       eventResponse,
		}
	}

	json.NewEncoder(w).Encode(resp)
}
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateEventRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		fmt.Print(err)
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

	eventToCreate := event.EventDetails{
		Event: event.Event{
			Title:             req.Title,
			Description:       req.Description,
			MinTeamSize:       req.MinTeamSize,
			MaxTeamSize:       req.MaxTeamSize,
			MaxTeamsPerSchool: req.MaxTeamsPerSchool,
			Status:            req.Status,
			Category:          req.Category,
		},
		Judges:   paresejudgeReqs(req.Judges),
		Criteria: parseCriteriaReqs(req.Criteria),
	}
	err := h.service.CreateEvent(ctx, &eventToCreate)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	eventResponse := ToEventDetailsResponse(&eventToCreate)
	json.NewEncoder(w).Encode(PostResponsePage[EventDetailsResponse]{
		Message:    "Event created successfully",
		Data:       eventResponse,
		StatusCode: http.StatusCreated,
	})
}
func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventIdStr := chi.URLParam(r, "eventId")
	log.Println("HandlerUpdateEvent::Updating event with ID:", eventIdStr)
	var req UpdateEventRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		fmt.Print(err)
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

	eventToUpdate := event.EventDetails{
		Event: event.Event{
			ID:                req.ID,
			Title:             req.Title,
			Description:       req.Description,
			MinTeamSize:       req.MinTeamSize,
			MaxTeamSize:       req.MaxTeamSize,
			MaxTeamsPerSchool: req.MaxTeamsPerSchool,
			Status:            req.Status,
			Category:          req.Category,
		},
		Judges:   paresejudgeReqs(req.Judges),
		Criteria: parseCriteriaReqs(req.Criteria),
	}

	eventId, err := util.ParseStrToInt64(eventIdStr)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid event ID in URL parameter",
		})
		return
	}
	if eventToUpdate.Event.ID != *eventId {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Event ID in request body does not match URL parameter",
		})
		return
	}

	err = h.service.UpdateEvent(ctx, &eventToUpdate)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	eventResponse := ToEventDetailsResponse(&eventToUpdate)
	json.NewEncoder(w).Encode(PutResponsePage[EventDetailsResponse]{
		Message:    "Event updated successfully",
		Data:       eventResponse,
		StatusCode: http.StatusOK,
	})
}
func (h *EventHandler) UpdateEventStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventIdStr := chi.URLParam(r, "eventId")
	log.Println("HandlerUpdateEventStatus::Updating event status with ID:", eventIdStr)
	var req UpdatEventStatusRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		fmt.Print(err)
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
	eventId, err := util.ParseStrToInt64(eventIdStr)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid event ID in URL parameter",
		})
		return
	}
	if req.ID != *eventId {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Event ID in request body does not match URL parameter",
		})
		return
	}
	err = h.service.UpdateEventStatus(ctx, *eventId, req.Status)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	json.NewEncoder(w).Encode(PutResponsePage[core.EventStatus]{
		Message:    "Event status changed successfully",
		Data:       req.Status,
		StatusCode: http.StatusOK,
	})
}
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventIdStr := chi.URLParam(r, "eventId")
	log.Println("HandlerDeleteEvent::Deleting event with ID:", eventIdStr)
	w.Header().Set("Content-Type", "application/json")
	var resp any
	eventId, err := strconv.ParseInt(eventIdStr, 10, 64)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
		json.NewEncoder(w).Encode(resp)
		return
	}
	err = h.service.DeleteEvent(ctx, eventId)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
	} else {
		resp = DeleteResponsePage{
			StatusCode: http.StatusNoContent,
			Message:    "Event deleted successfully",
		}
	}

	json.NewEncoder(w).Encode(resp)
}

func paresejudgeReqs(judgeReqs []EventJudgeRequest) []event.EventJudge {
	judges := make([]event.EventJudge, len(judgeReqs))
	for i, judgeReq := range judgeReqs {
		judges[i] = event.EventJudge{
			UserID: judgeReq.UserId,
		}
	}
	return judges
}
func parseCriteriaReqs(criteriaReqs []EventCriteriaRequest) []event.EventCriteria {
	criteria := make([]event.EventCriteria, len(criteriaReqs))
	for i, criteriaReq := range criteriaReqs {
		criteria[i] = event.EventCriteria{
			ID:       criteriaReq.ID,
			Title:    criteriaReq.Title,
			MaxScore: criteriaReq.MaxScore,
		}
	}
	return criteria
}
