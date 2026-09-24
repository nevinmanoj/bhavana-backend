package event

import (
	"errors"
)

var (
	ErrNotFound            = errors.New("Event not found")
	ErrInternal            = errors.New("Internal error")
	ErrUnauthorized        = errors.New("Unauthorized")
	ErrAlreadyExists       = errors.New("Event already exists")
	ErrEventFinalized      = errors.New("Event is finalized and cannot be modified")
	ErrInvalidStatusChange = errors.New("Invalid event status transition. Allowed: draft -> registration_open <-> registration_closed -> preparing -> open <-> closed -> finalized")
	ErrEventNotDraft       = errors.New("Event is not in draft status, cannot edit fields")
	ErrInvalidMemberRange  = errors.New("Invalid member range")

	ErrInvalidJudge           = errors.New("User is invalid or not a judge")
	ErrInvalidJudgeAssignment = errors.New("Judges can only be added when event is DRAFT, PREPARING, OPEN or CLOSED. Current status does not allow this operation")
	ErrInvalidJudgeRemoval    = errors.New("Judges can only be removed when event is DRAFT or PREPARING. Current status does not allow this operation")

	ErrInvalidCriteriaDeletion = errors.New("Event criteria can only be deleted when event is DRAFT or PREPARING. Current status does not allow this operation")
	ErrInvalidCriteriaAddition = errors.New("Event criteria can only be added when event is DRAFT or PREPARING. Current status does not allow this operation")
	ErrInvalidCriteriaEdit     = errors.New("Event criteria can only be edited when event is DRAFT or PREPARING. Current status does not allow this operation")
	ErrInvalidCriteriaMove     = errors.New("Event criteria can only be moved when event is DRAFT or PREPARING. Current status does not allow this operation")

	ErrStandingsRequired           = errors.New("At least one position-points mapping is required")
	ErrInvalidStandingModification = errors.New("Event standings can only be modified when event is DRAFT or PREPARING. Current status does not allow this operation")
	ErrDuplicateStandingPosition   = errors.New("Duplicate position in standings, each position must be unique")
	ErrEventNotClosed              = errors.New("Event must be CLOSED before it can be finalized")
	ErrCriteriaRequired             = errors.New("At least one scoring criterion is required before the event can be opened for judging")
	ErrChestNumbersNotDrawn        = errors.New("Chest numbers could not be drawn for this event")
)
