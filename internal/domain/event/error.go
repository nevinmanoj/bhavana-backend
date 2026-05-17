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
	ErrInvalidStatusChange = errors.New("Invalid status change, cannot move event back to draft")
	ErrEventNotDraft       = errors.New("Event is not in draft status, cannot edit fields")
	ErrinvalidTeamSize     = errors.New("Invalid team size")

	ErrInvalidJudge           = errors.New("User is invalid or not a judge")
	ErrInvalidJudgeAssignment = errors.New("Judges can only be addded when event is DRAFT, OPEN or CLOSED. Current status does not allow this operation")
	ErrInvalidJudgeRemoval    = errors.New("Judges can only be removed when event is DRAFT. Current status does not allow this operation")

	ErrInvalidCriteriaDeletion = errors.New("Event criteria can only be deleted when event is DRAFT. Current status does not allow this operation")
	ErrInvalidCriteriaAddition = errors.New("Event criteria can only be added when event is DRAFT. Current status does not allow this operation")
	ErrInvalidCriteriaEdit     = errors.New("Event criteria can only be edited when event is DRAFT. Current status does not allow this operation")
	ErrInvalidCriteriaMove     = errors.New("Event criteria can only be moved when event is DRAFT. Current status does not allow this operation")
)
