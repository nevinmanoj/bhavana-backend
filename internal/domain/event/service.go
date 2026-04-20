package event

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/user"
)

type EventService interface {
	GetEventByID(ctx context.Context, id int64) (*EventDetails, error)
	GetAllEvents(ctx context.Context, filter EventFilter) ([]Event, error)
	CreateEvent(ctx context.Context, event *EventDetails) error
	UpdateEvent(ctx context.Context, event *EventDetails) error
	UpdateEventStatus(ctx context.Context, eventID int64, status core.EventStatus) error
	DeleteEvent(ctx context.Context, eventID int64) error
}

type eventService struct {
	repo     EventWriteRepository
	userRepo user.UserReadRepository
	db       *sqlx.DB
}

func NewEventService(repo EventWriteRepository, userReadRepo user.UserReadRepository, db *sqlx.DB) EventService {
	return &eventService{repo: repo, userRepo: userReadRepo, db: db}
}
func (s *eventService) GetEventByID(ctx context.Context, id int64) (*EventDetails, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	event, err := s.repo.GetEventByID(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	judges, err := s.repo.GetEventJudges(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	criteria, err := s.repo.GetEventCriteria(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	return &EventDetails{
		Event:    *event,
		Judges:   judges,
		Criteria: criteria,
	}, nil
}
func (s *eventService) GetAllEvents(ctx context.Context, filter EventFilter) ([]Event, error) {
	events, err := s.repo.GetAllEvents(ctx, s.db, filter)
	if err != nil || events == nil || len(events) == 0 {
		return []Event{}, err
	}
	return events, nil
}
func (s *eventService) CreateEvent(ctx context.Context, event *EventDetails) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	//create event
	err = s.repo.CreateEvent(ctx, tx, &event.Event)
	if err != nil {
		return fmt.Errorf("error creating event: %w", err)
	}
	//create judges
	if err := s.syncEventJudges(ctx, tx, event); err != nil {
		return err
	}
	//create criteria
	if err := s.syncEventCriterias(ctx, tx, event); err != nil {
		return err
	}

	return tx.Commit()
}
func (s *eventService) UpdateEvent(ctx context.Context, event *EventDetails) error {
	existingEvent, err := s.repo.GetEventByID(ctx, s.db, event.Event.ID)
	if err != nil {
		return fmt.Errorf("error fetching event: %w", err)
	}
	//check if finalized
	if existingEvent.Status == core.EventStatusFinalized {
		return fmt.Errorf("finalized events cannot be updated")
	}
	//check if status is being updated to draft from open or closed
	if existingEvent.Status != core.EventStatusDraft && event.Event.Status == core.EventStatusDraft {
		return fmt.Errorf("cannot change event status back to draft")
	}
	//check if we are editing core fields when not in draft status
	if existingEvent.Status != core.EventStatusDraft {
		//checking for core field changes
		if existingEvent.Title != event.Event.Title ||
			existingEvent.Description != event.Event.Description ||
			existingEvent.Category != event.Event.Category ||
			existingEvent.MinTeamSize != event.Event.MinTeamSize ||
			existingEvent.MaxTeamSize != event.Event.MaxTeamSize ||
			existingEvent.MaxTeamsPerSchool != event.Event.MaxTeamsPerSchool {
			return fmt.Errorf("cannot edit core fields of a non-draft event")
		}
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()
	event.Event.CreatedAt = existingEvent.CreatedAt

	//update the core fields
	err = s.repo.UpdateEvent(ctx, tx, &event.Event)
	if err != nil {
		return fmt.Errorf("error updating event: %w", err)
	}

	//sync judges
	if err := s.syncEventJudges(ctx, tx, event); err != nil {
		return err
	}
	//sync criteria
	if err := s.syncEventCriterias(ctx, tx, event); err != nil {
		return err
	}

	return tx.Commit()
}
func (s *eventService) UpdateEventStatus(ctx context.Context, eventID int64, status core.EventStatus) error {
	existingEvent, err := s.repo.GetEventByID(ctx, s.db, eventID)
	if err != nil {
		return fmt.Errorf("error fetching event: %w", err)
	}
	//check if finalized
	if existingEvent.Status == core.EventStatusFinalized {
		return fmt.Errorf("finalized events cannot be updated")
	}
	//check if status is being updated to draft from open or closed
	if existingEvent.Status != core.EventStatusDraft && status == core.EventStatusDraft {
		return fmt.Errorf("cannot change event status back to draft")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	//update the status
	err = s.repo.UpdateEventStatus(ctx, tx, &status, eventID)
	if err != nil {
		return fmt.Errorf("error updating event Status: %w", err)
	}

	return tx.Commit()
}
func (s *eventService) DeleteEvent(ctx context.Context, eventID int64) error {
	return s.repo.DeleteEvent(ctx, s.db, eventID)
}

// helper functions
func (s *eventService) syncEventJudges(ctx context.Context, tx *sqlx.Tx, event *EventDetails) error {
	eventID := event.Event.ID
	existing := []EventJudge{}
	var err error
	if eventID != 0 {
		existing, err = s.repo.GetEventJudges(ctx, tx, event.Event.ID)
		if err != nil {
			return fmt.Errorf("error fetching event judges: %w", err)
		}
	}
	requested := event.Judges
	existingMap := make(map[int64]bool)
	for _, j := range existing {
		existingMap[j.UserID] = true
	}

	requestedMap := make(map[int64]bool)
	for _, j := range requested {
		requestedMap[j.UserID] = true
	}

	// delete removed judges
	for _, j := range existing {
		if !requestedMap[j.UserID] {
			if event.Event.Status != core.EventStatusDraft {
				return fmt.Errorf("Judges can only be removed when the EventStatus:Draft")
			}
			if err := s.repo.DeleteEventJudge(ctx, tx, eventID, j.UserID); err != nil {
				return fmt.Errorf("error deleting judge: %w", err)
			}
		}
	}

	// insert new judges
	for _, j := range requested {
		if !existingMap[j.UserID] {
			if event.Event.Status != core.EventStatusFinalized {
				return fmt.Errorf("Judges cannot be added in EventStatus:Finalized")
			}
			// Check if the user exists as a judge
			exists, err := s.userRepo.ExistsAsJudge(ctx, tx, j.UserID)
			if err != nil {
				return fmt.Errorf("error checking if user exists as judge: %w", err)
			}
			if !exists {
				return fmt.Errorf("user with ID %d does not exist or is not a judge", j.UserID)
			}
			if err := s.repo.CreateEventJudge(ctx, tx, &EventJudge{
				EventID: eventID,
				UserID:  j.UserID,
			}); err != nil {
				return fmt.Errorf("error adding judge: %w", err)
			}
		}
	}
	createdJudges, err := s.repo.GetEventJudges(ctx, tx, event.Event.ID)
	if err != nil {
		return fmt.Errorf("error fetching created judges: %w", err)
	}
	event.Judges = createdJudges
	return nil
}

func (s *eventService) syncEventCriterias(ctx context.Context, tx *sqlx.Tx, event *EventDetails) error {
	eventID := event.Event.ID
	existing := []EventCriteria{}
	var err error
	if eventID != 0 {
		existing, err = s.repo.GetEventCriteria(ctx, tx, event.Event.ID)
		if err != nil {
			return fmt.Errorf("error fetching event criteria: %w", err)
		}
	}
	requested := event.Criteria
	existingMap := make(map[int64]EventCriteria)
	for _, c := range existing {
		existingMap[c.ID] = c
	}

	requestedMap := make(map[int64]bool)
	for _, c := range requested {
		if c.ID != 0 {
			requestedMap[c.ID] = true
		}
	}
	criteriaUpdated := false
	// delete removed criterias
	for _, c := range existing {
		if !requestedMap[c.ID] {
			if event.Event.Status != core.EventStatusDraft {
				return fmt.Errorf("Event Criterias can only be removed when EventStatus:Draft")
			}
			if err := s.repo.DeleteEventCriteria(ctx, tx, c.ID); err != nil {
				return fmt.Errorf("error deleting criteria: %w", err)
			}
			criteriaUpdated = true
		}
	}
	// add new criterias
	for _, c := range requested {
		_, exists := existingMap[c.ID]
		if !exists {
			if event.Event.Status != core.EventStatusDraft {
				return fmt.Errorf("Event Criterias can only be added when EventStatus:Draft")
			}
			// insert new criteria
			if err := s.repo.CreateEventCriteria(ctx, tx, &EventCriteria{
				EventID:  eventID,
				Title:    c.Title,
				MaxScore: c.MaxScore,
			}); err != nil {
				return fmt.Errorf("error adding criteria: %w", err)
			}
			criteriaUpdated = true
		}
	}

	if criteriaUpdated {
		createdCriteria, err := s.repo.GetEventCriteria(ctx, tx, event.Event.ID)
		if err != nil {
			return fmt.Errorf("error fetching created criteria: %w", err)
		}
		event.Criteria = createdCriteria
	}

	return nil
}
