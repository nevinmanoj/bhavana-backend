package event

import (
	"context"

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
	conditions := []string{}
	if scope.UserID != nil && role == rbac.UserRoleJudge {
		baseQuery += " JOIN event_judges ej ON ej.event_id = e.id"
		conditions = append(conditions, "ej.user_id = ?")
		conditions = append(conditions, "e.status = 'open'")
		args = append(args, *scope.UserID)
	}

	finalQuery, finalArgs, err := buildEventQuery(baseQuery, args, conditions, filter)
	err = sqlx.SelectContext(
		ctx, db,
		&events,
		finalQuery, finalArgs...,
	)
	if err != nil {
		return nil, event.ErrInternal
	}
	return events, nil
}
func (r *eventRepository) GetEventByID(ctx context.Context, db sqlx.ExtContext, id int64) (*event.Event, error) {
	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	events := []event.Event{}

	query := `SELECT e.* FROM events e WHERE e.id = $1`
	args := []any{id}

	if scope.UserID != nil && role == rbac.UserRoleJudge {
		// judges can only fetch events they are assigned to judge, matching the scoping in GetAllEvents.
		query = `SELECT e.* FROM events e
			JOIN event_judges ej ON ej.event_id = e.id
			WHERE e.id = $1 AND ej.user_id = $2`
		args = append(args, *scope.UserID)
	}

	err := sqlx.SelectContext(ctx, db, &events, query, args...)

	if err != nil {
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
			min_members,
			max_members,
			max_entries_per_school,
			status,
			category,
			created_at
		)
		VALUES (
			:title,
			:description,
			:min_members,
			:max_members,
			:max_entries_per_school,
			:status,
			:category,
			:created_at
		)
		RETURNING id, created_at
	`

	rows, err := sqlx.NamedQueryContext(ctx, db, query, eventToCreate)
	if err != nil {
		return errorMapper(err)
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&eventToCreate.ID, &eventToCreate.CreatedAt)
		return nil
	}

	return event.ErrInternal
}
func (r *eventRepository) UpdateEvent(ctx context.Context, db sqlx.ExtContext, eventToUpdate *event.Event) error {

	query := `
		UPDATE events
		SET title = :title,
			description = :description,
			min_members = :min_members,
			max_members = :max_members,
			max_entries_per_school = :max_entries_per_school,
			category = :category,
			status = :status
		WHERE id = :id
	`
	result, err := sqlx.NamedExecContext(ctx, db, query, eventToUpdate)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errorMapper(err)
	}
	if rows == 0 {
		return event.ErrNotFound
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
		return errorMapper(err)
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
		return errorMapper(err)
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
		return nil, event.ErrInternal
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
		return errorMapper(err)
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&judgeToCreate.EventID, &judgeToCreate.UserID)
		return nil
	}
	return event.ErrInternal
}
func (r *eventRepository) DeleteEventJudge(ctx context.Context, db sqlx.ExtContext, eventID int64, userID int64) error {
	query := `
		DELETE FROM event_judges
		WHERE event_id = $1 AND user_id = $2
	`
	_, err := db.ExecContext(ctx, query, eventID, userID)
	if err != nil {
		return errorMapper(err)
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
		return nil, event.ErrInternal
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
		return errorMapper(err)
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&criteriaToCreate.ID, &criteriaToCreate.CreatedAt)
		return nil
	}

	return event.ErrInternal
}
func (r *eventRepository) DeleteEventCriteria(ctx context.Context, db sqlx.ExtContext, criteriaID int64) error {
	query := `
		DELETE FROM event_criteria
		WHERE id = $1
	`
	_, err := db.ExecContext(ctx, query, criteriaID)
	if err != nil {
		return errorMapper(err)
	}
	return nil
}

// event standings
func (r *eventRepository) GetEventStandings(ctx context.Context, db sqlx.ExtContext, eventID int64) ([]event.EventStanding, error) {
	standings := []event.EventStanding{}
	err := sqlx.SelectContext(
		ctx, db,
		&standings,
		`SELECT *
		FROM event_standings
		WHERE event_id = $1
		ORDER BY position ASC`,
		eventID,
	)
	if err != nil {
		return nil, event.ErrInternal
	}
	return standings, nil
}
func (r *eventRepository) CreateEventStanding(ctx context.Context, db sqlx.ExtContext, standingToCreate *event.EventStanding) error {
	query := `
		INSERT INTO event_standings (
			event_id,
			position,
			points
		)
		VALUES (
			:event_id,
			:position,
			:points
		)
		RETURNING id, created_at
	`
	rows, err := sqlx.NamedQueryContext(ctx, db, query, standingToCreate)
	if err != nil {
		return errorMapper(err)
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&standingToCreate.ID, &standingToCreate.CreatedAt)
		return nil
	}

	return event.ErrInternal
}
func (r *eventRepository) UpdateEventStanding(ctx context.Context, db sqlx.ExtContext, standingToUpdate *event.EventStanding) error {
	query := `
		UPDATE event_standings
		SET position = :position,
			points = :points
		WHERE id = :id
	`
	result, err := sqlx.NamedExecContext(ctx, db, query, standingToUpdate)
	if err != nil {
		return errorMapper(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return errorMapper(err)
	}
	if rows == 0 {
		return event.ErrNotFound
	}
	return nil
}
func (r *eventRepository) DeleteEventStanding(ctx context.Context, db sqlx.ExtContext, standingID int64) error {
	query := `
		DELETE FROM event_standings
		WHERE id = $1
	`
	_, err := db.ExecContext(ctx, query, standingID)
	if err != nil {
		return errorMapper(err)
	}
	return nil
}
