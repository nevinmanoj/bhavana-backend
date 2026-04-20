package event

import (
	"context"
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/event"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

type eventRepository struct {
}

func NewEventWriteRepository() event.EventWriteRepository {
	return &eventRepository{}
}
func NewEventReadRepository() event.EventReadRepository {
	return &eventRepository{}
}

// events
func (r *eventRepository) GetAllEvents(ctx context.Context, db sqlx.ExtContext, filter event.EventFilter) ([]event.Event, error) {
	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	events := []event.Event{}
	baseQuery := `SELECT e.* FROM events e`
	args := []any{}

	if scope.UserID != nil && role == rbac.UserRoleJudge {
		baseQuery += " JOIN event_judges ej ON ej.event_id = e.id WHERE ej.user_id = $1"
		args = append(args, *scope.UserID)
	}

	finalQuery, finalArgs, err := buildEventQuery(baseQuery, args, filter)
	err = sqlx.SelectContext(
		ctx, db,
		&events,
		finalQuery, finalArgs...,
	)
	if err != nil {
		return nil, err
	}
	return events, nil
}
func (r *eventRepository) GetEventByID(ctx context.Context, db sqlx.ExtContext, id int64) (*event.Event, error) {
	events := []event.Event{}
	err := sqlx.SelectContext(
		ctx, db,
		&events,
		`SELECT * FROM events
		 WHERE id = $1`,
		id,
	)

	if err != nil {
		log.Println("Error fetching event by id:", err)
		return nil, event.ErrInternal
	}

	if len(events) == 0 {
		return nil, event.ErrNotFound
	}
	event := events[0]
	return &event, nil

}
func (r *eventRepository) CreateEvent(ctx context.Context, db sqlx.ExtContext, eventToCreate *event.Event) error {

	query := `
		INSERT INTO events (
			title,
			description,
			min_team_size,
			max_team_size,
			max_teams_per_school,
			status,
			category,
			created_at
		)
		VALUES (
			:title,
			:description,
			:min_team_size,
			:max_team_size,
			:max_teams_per_school,
			:status,
			:category,
			:created_at
		)
		RETURNING id, created_at
	`

	rows, err := sqlx.NamedQueryContext(ctx, db, query, eventToCreate)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&eventToCreate.ID, &eventToCreate.CreatedAt)
		return nil
	}

	return sql.ErrNoRows
}
func (r *eventRepository) UpdateEvent(ctx context.Context, db sqlx.ExtContext, eventToUpdate *event.Event) error {

	query := `
		UPDATE events
		SET title = :title,
			description = :description,
			min_team_size = :min_team_size,
			max_team_size = :max_team_size,
			max_teams_per_school = :max_teams_per_school,
			category = :category
		WHERE id = :id
	`
	result, err := sqlx.NamedExecContext(ctx, db, query, eventToUpdate)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return event.ErrInternal
	}
	return nil
}
func (r *eventRepository) UpdateEventStatus(ctx context.Context, db sqlx.ExtContext, status *core.EventStatus, eventID int64) error {
	query := `UPDATE events SET status = $1 WHERE id = $2`
	result, err := db.ExecContext(ctx, query, status, eventID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return event.ErrNotFound
	}

	return nil
}
func (r *eventRepository) DeleteEvent(ctx context.Context, db sqlx.ExtContext, eventID int64) error {
	query := `
		DELETE FROM events
		WHERE id = $1
	`
	_, err := db.ExecContext(ctx, query, eventID)
	if err != nil {
		return err
	}
	return nil
}

// event judges
func (r *eventRepository) GetEventJudges(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]event.EventJudge, error) {
	judges := []event.EventJudge{}
	err := sqlx.SelectContext(
		ctx, db,
		&judges,
		`SELECT u.name, ej.event_id, ej.user_id 
		FROM event_judges ej 
		JOIN users u ON ej.user_id = u.id 
		WHERE ej.event_id = $1`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	return judges, nil
}
func (r *eventRepository) CreateEventJudge(ctx context.Context, db sqlx.ExtContext, judgeToCreate *event.EventJudge) error {
	query := `
		INSERT INTO event_judges (
			event_id,
			user_id
		)
		VALUES (
			:event_id,
			:user_id
		)
		RETURNING event_id, user_id					
	`
	rows, err := sqlx.NamedQueryContext(ctx, db, query, judgeToCreate)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&judgeToCreate.EventID, &judgeToCreate.UserID)
		return nil
	}
	return sql.ErrNoRows
}
func (r *eventRepository) DeleteEventJudge(ctx context.Context, db sqlx.ExtContext, eventID int64, userID int64) error {
	query := `
		DELETE FROM event_judges
		WHERE event_id = $1 AND user_id = $2
	`
	_, err := db.ExecContext(ctx, query, eventID, userID)
	if err != nil {
		return err
	}
	return nil
}

// event criteria
func (r *eventRepository) GetEventCriteria(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]event.EventCriteria, error) {
	criteria := []event.EventCriteria{}
	err := sqlx.SelectContext(
		ctx, db,
		&criteria,
		`SELECT * 
		FROM event_criteria
		WHERE event_id = $1`,
		eventID,
	)
	if err != nil {
		return nil, err
	}
	return criteria, nil
}
func (r *eventRepository) CreateEventCriteria(ctx context.Context, db sqlx.ExtContext, criteriaToCreate *event.EventCriteria) error {
	query := `
		INSERT INTO event_criteria (
			event_id,
			title,
			max_score
		)
		VALUES (
			:event_id,
			:title,
			:max_score
		)
		RETURNING id, created_at					
	`
	rows, err := sqlx.NamedQueryContext(ctx, db, query, criteriaToCreate)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&criteriaToCreate.ID, &criteriaToCreate.CreatedAt)
		return nil
	}

	return sql.ErrNoRows
}
func (r *eventRepository) DeleteEventCriteria(ctx context.Context, db sqlx.ExtContext, criteriaID int64) error {
	query := `
		DELETE FROM event_criteria
		WHERE id = $1
	`
	_, err := db.ExecContext(ctx, query, criteriaID)
	if err != nil {
		return err
	}
	return nil
}
