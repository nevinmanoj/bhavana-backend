package event

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
)

type EventWriteRepository interface {
	EventReadRepository
	CreateEvent(ctx context.Context, db sqlx.ExtContext, eventToCreate *Event) error
	UpdateEvent(ctx context.Context, db sqlx.ExtContext, eventToUpdate *Event) error
	UpdateEventStatus(ctx context.Context, db sqlx.ExtContext, status *core.EventStatus, eventID int64) error
	DeleteEvent(ctx context.Context, db sqlx.ExtContext, eventID int64) error

	CreateEventCriteria(ctx context.Context, db sqlx.ExtContext, criteria *EventCriteria) error
	DeleteEventCriteria(ctx context.Context, db sqlx.ExtContext, criteriaID int64) error

	CreateEventJudge(ctx context.Context, db sqlx.ExtContext, judge *EventJudge) error
	DeleteEventJudge(ctx context.Context, db sqlx.ExtContext, eventID int64, userID int64) error
}
type EventReadRepository interface {
	GetEventByID(ctx context.Context, db sqlx.ExtContext, id int64) (*Event, error)
	GetAllEvents(ctx context.Context, db sqlx.ExtContext, filter EventFilter) ([]Event, error)

	GetEventJudges(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]EventJudge, error)

	GetEventCriteria(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]EventCriteria, error)
}
