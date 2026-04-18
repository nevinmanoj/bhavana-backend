package event

import (
	core "github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/event"
)

// requests
type CreateEventRequest struct {
	Title             string                 `json:"title" validate:"required"`
	Description       string                 `json:"description"`
	MinTeamSize       int64                  `json:"min_team_size" validate:"required"`
	MaxTeamSize       int64                  `json:"max_team_size" validate:"required"`
	MaxTeamsPerSchool int64                  `json:"max_teams_per_school" validate:"required"`
	Status            core.EventStatus       `json:"status" validate:"required,event_status"`
	Category          core.Category          `json:"category" validate:"required,category"`
	Judges            []EventJudgeRequest    `json:"judges"`
	Criteria          []EventCriteriaRequest `json:"criteria"`
}
type UpdateEventRequest struct {
	ID int64 `json:"id" validate:"required"`
	CreateEventRequest
}
type UpdatEventStatusRequest struct {
	ID     int64            `json:"id" validate:"required"`
	Status core.EventStatus `json:"status" validate:"required,event_status"`
}

type EventJudgeRequest struct {
	UserId int64 `json:"user_id" validate:"required"`
}
type EventCriteriaRequest struct {
	ID       int64   `json:"id"`
	Title    string  `json:"title" validate:"required"`
	MaxScore float64 `json:"max_score" validate:"required"`
}

// responses
type EventResponse struct {
	ID                int64            `json:"id"`
	Title             string           `json:"title"`
	Description       string           `json:"description"`
	MinTeamSize       int64            `json:"min_team_size"`
	MaxTeamSize       int64            `json:"max_team_size"`
	MaxTeamsPerSchool int64            `json:"max_teams_per_school"`
	Status            core.EventStatus `json:"status"`
	Category          core.Category    `json:"category"`
	CreatedAt         string           `json:"created_at"`
}
type EventDetailsResponse struct {
	EventResponse
	Judges   []EventJudgeResponse    `json:"judges"`
	Criteria []EventCriteriaResponse `json:"criteria"`
}
type EventJudgeResponse struct {
	Name   string `json:"name"`
	UserId int64  `json:"user_id"`
}
type EventCriteriaResponse struct {
	ID       int64   `json:"id"`
	Title    string  `json:"title"`
	MaxScore float64 `json:"max_score"`
}

func ToEventDetailsResponse(details *event.EventDetails) EventDetailsResponse {
	judges := make([]EventJudgeResponse, len(details.Judges))
	for i, judge := range details.Judges {
		judges[i] = EventJudgeResponse{
			Name:   judge.Name,
			UserId: judge.UserID,
		}
	}
	criteria := make([]EventCriteriaResponse, len(details.Criteria))
	for i, c := range details.Criteria {
		criteria[i] = EventCriteriaResponse{
			ID:       c.ID,
			Title:    c.Title,
			MaxScore: c.MaxScore,
		}
	}
	return EventDetailsResponse{
		EventResponse: ToEventResponse(&details.Event),
		Judges:        judges,
		Criteria:      criteria,
	}
}

func ToEventResponse(e *event.Event) EventResponse {
	return EventResponse{
		ID:                e.ID,
		Title:             e.Title,
		Description:       e.Description,
		MinTeamSize:       e.MinTeamSize,
		MaxTeamSize:       e.MaxTeamSize,
		MaxTeamsPerSchool: e.MaxTeamsPerSchool,
		Status:            e.Status,
		Category:          e.Category,
		CreatedAt:         e.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
