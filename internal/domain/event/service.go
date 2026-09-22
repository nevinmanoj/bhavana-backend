package event

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/user"
)

// ResultGenerator is implemented by the result domain's service. Declared here
// (consumer side) so this package never imports domain/result and there is no
// import cycle; app.go wires the concrete implementation in.
type ResultGenerator interface {
	GenerateForEvent(ctx context.Context, tx *sqlx.Tx, eventID int64) error
}

type EventService interface {
	GetEventByID(ctx context.Context, id int64) (*EventDetails, error)
	GetAllEvents(ctx context.Context, filter EventFilter) ([]Event, error)
	CreateEvent(ctx context.Context, event *EventDetails) error
	UpdateEvent(ctx context.Context, event *EventDetails) error
	UpdateEventStatus(ctx context.Context, eventID int64, status core.EventStatus) error
	DeleteEvent(ctx context.Context, eventID int64) error
}

type eventService struct {
	db              *sqlx.DB
	repo            EventWriteRepository
	userRepo        user.UserReadRepository
	resultGenerator ResultGenerator
}

func NewEventService(db *sqlx.DB, repo EventWriteRepository, userReadRepo user.UserReadRepository, resultGenerator ResultGenerator) EventService {
	return &eventService{db: db, repo: repo, userRepo: userReadRepo, resultGenerator: resultGenerator}
}
func (s *eventService) GetEventByID(ctx context.Context, id int64) (*EventDetails, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

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

	standings, err := s.repo.GetEventStandings(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	return &EventDetails{
		Event:     *event,
		Judges:    judges,
		Criteria:  criteria,
		Standings: standings,
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
	if len(event.Standings) == 0 {
		return ErrStandingsRequired
	}
	if err := checkDuplicateStandingPositions(event.Standings); err != nil {
		return err
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	//create event
	err = s.repo.CreateEvent(ctx, tx, &event.Event)
	if err != nil {
		return err
	}
	//create judges
	if err := s.syncEventJudges(ctx, tx, event); err != nil {
		return err
	}
	//create criteria
	if err := s.syncEventCriterias(ctx, tx, event); err != nil {
		return err
	}
	//create standings
	if err := s.syncEventStandings(ctx, tx, event); err != nil {
		return err
	}

	return tx.Commit()
}
func (s *eventService) UpdateEvent(ctx context.Context, event *EventDetails) error {
	if len(event.Standings) == 0 {
		return ErrStandingsRequired
	}
	if err := checkDuplicateStandingPositions(event.Standings); err != nil {
		return err
	}

	existingEvent, err := s.repo.GetEventByID(ctx, s.db, event.Event.ID)
	if err != nil {
		return fmt.Errorf("error fetching event: %w", err)
	}
	//check if finalized
	if existingEvent.Status == core.EventStatusFinalized {
		return ErrEventFinalized
	}
	//check if status is being updated to draft from open or closed
	if existingEvent.Status != core.EventStatusDraft && event.Event.Status == core.EventStatusDraft {
		return ErrInvalidStatusChange
	}
	//check if we are editing core fields when not in draft status
	if existingEvent.Status != core.EventStatusDraft {
		//checking for core field changes
		if existingEvent.Title != event.Event.Title ||
			existingEvent.Description != event.Event.Description ||
			existingEvent.Category != event.Event.Category ||
			existingEvent.MinMembers != event.Event.MinMembers ||
			existingEvent.MaxMembers != event.Event.MaxMembers ||
			existingEvent.MaxEntriesPerSchool != event.Event.MaxEntriesPerSchool {
			return ErrEventNotDraft
		}
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return ErrInternal
	}
	defer tx.Rollback()
	event.Event.CreatedAt = existingEvent.CreatedAt

	//update the core fields
	err = s.repo.UpdateEvent(ctx, tx, &event.Event)
	if err != nil {
		return err
	}

	//sync judges
	if err := s.syncEventJudges(ctx, tx, event); err != nil {
		return err
	}
	//sync criteria
	if err := s.syncEventCriterias(ctx, tx, event); err != nil {
		return err
	}
	//sync standings
	if err := s.syncEventStandings(ctx, tx, event); err != nil {
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
		return ErrEventFinalized
	}
	//the lifecycle graph itself is enforced by the events BEFORE UPDATE trigger
	//finalizing has extra preconditions and a side effect: computing results
	if status == core.EventStatusFinalized {
		if existingEvent.Status != core.EventStatusClosed {
			return ErrEventNotClosed
		}
		standings, err := s.repo.GetEventStandings(ctx, s.db, eventID)
		if err != nil {
			return err
		}
		if len(standings) == 0 {
			return ErrStandingsRequired
		}
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	//update the status
	err = s.repo.UpdateEventStatus(ctx, tx, &status, eventID)
	if err != nil {
		return err
	}

	//registration has just ended: draw chest numbers for the whole event in one
	//shot, so no two schools race for the same number during registration
	if existingEvent.Status == core.EventStatusRegistrationClosed && status == core.EventStatusPreparing {
		if _, err := s.repo.AssignChestNumbers(ctx, tx, eventID); err != nil {
			return err
		}
	}

	//compute and persist results once the event reads as finalized in this tx
	if status == core.EventStatusFinalized {
		if err := s.resultGenerator.GenerateForEvent(ctx, tx, eventID); err != nil {
			return err
		}
	}

	return tx.Commit()
}
func (s *eventService) DeleteEvent(ctx context.Context, eventID int64) error {
	existingEvent, err := s.repo.GetEventByID(ctx, s.db, eventID)
	if err != nil {
		return fmt.Errorf("error fetching event: %w", err)
	}
	if existingEvent.Status == core.EventStatusFinalized {
		return ErrEventFinalized
	}
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
			return err
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
				return ErrInvalidJudgeRemoval
			}
			if err := s.repo.DeleteEventJudge(ctx, tx, eventID, j.UserID); err != nil {
				return err
			}
		}
	}

	// insert new judges
	for _, j := range requested {
		if !existingMap[j.UserID] {
			if event.Event.Status == core.EventStatusFinalized {
				return ErrInvalidJudgeAssignment
			}
			// Check if the user exists as a judge
			exists, err := s.userRepo.ExistsAsJudge(ctx, tx, j.UserID)
			if err != nil {
				return ErrInternal
			}
			if !exists {
				return ErrInvalidJudge
			}
			if err := s.repo.CreateEventJudge(ctx, tx, &EventJudge{
				EventID: eventID,
				UserID:  j.UserID,
			}); err != nil {
				return ErrInternal
			}
		}
	}
	createdJudges, err := s.repo.GetEventJudges(ctx, tx, event.Event.ID)
	if err != nil {
		return ErrInternal
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
			return ErrInternal
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
				return ErrInvalidCriteriaDeletion
			}
			if err := s.repo.DeleteEventCriteria(ctx, tx, c.ID); err != nil {
				return ErrInternal
			}
			criteriaUpdated = true
		}
	}
	// add new criterias
	for _, c := range requested {
		_, exists := existingMap[c.ID]
		if !exists {
			if event.Event.Status != core.EventStatusDraft {
				return ErrInvalidCriteriaAddition
			}
			// insert new criteria
			if err := s.repo.CreateEventCriteria(ctx, tx, &EventCriteria{
				EventID:  eventID,
				Title:    c.Title,
				MaxScore: c.MaxScore,
			}); err != nil {
				return ErrInternal
			}
			criteriaUpdated = true
		}
	}

	if criteriaUpdated {
		createdCriteria, err := s.repo.GetEventCriteria(ctx, tx, event.Event.ID)
		if err != nil {
			return ErrInternal
		}
		event.Criteria = createdCriteria
	}

	return nil
}

func (s *eventService) syncEventStandings(ctx context.Context, tx *sqlx.Tx, event *EventDetails) error {
	eventID := event.Event.ID
	existing := []EventStanding{}
	var err error
	if eventID != 0 {
		existing, err = s.repo.GetEventStandings(ctx, tx, eventID)
		if err != nil {
			return ErrInternal
		}
	}
	requested := event.Standings
	existingMap := make(map[int64]EventStanding)
	for _, st := range existing {
		existingMap[st.ID] = st
	}

	requestedMap := make(map[int64]bool)
	for _, st := range requested {
		if st.ID != 0 {
			requestedMap[st.ID] = true
		}
	}

	standingsUpdated := false

	// delete standings removed from the request
	for _, st := range existing {
		if !requestedMap[st.ID] {
			if event.Event.Status != core.EventStatusDraft {
				return ErrInvalidStandingModification
			}
			if err := s.repo.DeleteEventStanding(ctx, tx, st.ID); err != nil {
				return err
			}
			standingsUpdated = true
		}
	}

	// add new standings, and update existing ones whose position/points changed
	for _, st := range requested {
		existingStanding, exists := existingMap[st.ID]
		if !exists {
			if event.Event.Status != core.EventStatusDraft {
				return ErrInvalidStandingModification
			}
			if err := s.repo.CreateEventStanding(ctx, tx, &EventStanding{
				EventID:  eventID,
				Position: st.Position,
				Points:   st.Points,
			}); err != nil {
				return err
			}
			standingsUpdated = true
			continue
		}
		if existingStanding.Position != st.Position || existingStanding.Points != st.Points {
			if event.Event.Status != core.EventStatusDraft {
				return ErrInvalidStandingModification
			}
			if err := s.repo.UpdateEventStanding(ctx, tx, &EventStanding{
				ID:       st.ID,
				EventID:  eventID,
				Position: st.Position,
				Points:   st.Points,
			}); err != nil {
				return err
			}
			standingsUpdated = true
		}
	}

	if standingsUpdated {
		createdStandings, err := s.repo.GetEventStandings(ctx, tx, eventID)
		if err != nil {
			return ErrInternal
		}
		event.Standings = createdStandings
	}

	return nil
}

// checkDuplicateStandingPositions rejects a request containing two standings
// for the same position before it ever reaches the DB's unique constraint.
func checkDuplicateStandingPositions(standings []EventStanding) error {
	seen := make(map[int64]bool)
	for _, st := range standings {
		if seen[st.Position] {
			return ErrDuplicateStandingPosition
		}
		seen[st.Position] = true
	}
	return nil
}
