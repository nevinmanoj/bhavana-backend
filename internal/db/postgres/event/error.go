package event

import (
	"github.com/lib/pq"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/event"
)

func errorMapper(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		//events
		case "23505":
			if pqErr.Constraint == "uq_event_standing_position" {
				return event.ErrDuplicateStandingPosition
			}
			return event.ErrinvalidTeamSize
		case "P0201":
			return event.ErrInvalidStatusChange
		case "P0202":
			return event.ErrEventFinalized
		case "P0203":
			return event.ErrEventNotDraft
		//event_judge related errors
		case "P0204":
			return event.ErrInvalidJudgeRemoval
		case "P0205":
			return event.ErrInvalidJudgeAssignment
		case "P0206":
			return event.ErrInvalidJudgeAssignment
		case "P0207":
			return event.ErrInvalidJudgeRemoval
		//event_criteria related errors
		case "P0208":
			return event.ErrInvalidCriteriaDeletion
		case "P0209":
			return event.ErrInvalidCriteriaAddition
		case "P0210":
			return event.ErrInvalidCriteriaEdit
		case "P0211":
			return event.ErrInvalidCriteriaMove
		//event_standings related errors
		case "P0212", "P0213", "P0214", "P0215":
			return event.ErrInvalidStandingModification
		//events - finalized delete guard
		case "P0216":
			return event.ErrEventFinalized
		default:
			return event.ErrInternal
		}
	}

	return err
}
