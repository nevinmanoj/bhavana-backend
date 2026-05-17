package event

import (
	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/event"
)

func GetEventDomainErrorResponse(err error) ErrorResponse {
	switch err {
	//user errrors
	case event.ErrUnauthorized:
		return ErrorResponse{
			StatusCode: 403,
			Message:    event.ErrUnauthorized.Error(),
		}
	case event.ErrNotFound:
		return ErrorResponse{
			StatusCode: 404,
			Message:    event.ErrNotFound.Error(),
		}
	case event.ErrAlreadyExists:
		return ErrorResponse{
			StatusCode: 400,
			Message:    event.ErrAlreadyExists.Error(),
		}
	case event.ErrEventFinalized:
		return ErrorResponse{
			StatusCode: 400,
			Message:    event.ErrEventFinalized.Error(),
		}
	case event.ErrInvalidStatusChange:
		return ErrorResponse{
			StatusCode: 400,
			Message:    event.ErrInvalidStatusChange.Error(),
		}
	case event.ErrEventNotDraft:
		return ErrorResponse{
			StatusCode: 400,
			Message:    event.ErrEventNotDraft.Error(),
		}
	case event.ErrinvalidTeamSize:
		return ErrorResponse{
			StatusCode: 400,
			Message:    event.ErrinvalidTeamSize.Error(),
		}
	case event.ErrInvalidJudge:
		return ErrorResponse{
			StatusCode: 400,
			Message:    event.ErrInvalidJudge.Error(),
		}
	case event.ErrInvalidJudgeAssignment:
		return ErrorResponse{
			StatusCode: 400,
			Message:    event.ErrInvalidJudgeAssignment.Error(),
		}
	case event.ErrInvalidJudgeRemoval:
		return ErrorResponse{
			StatusCode: 400,
			Message:    event.ErrInvalidJudgeRemoval.Error(),
		}
	case event.ErrInvalidCriteriaDeletion:
		return ErrorResponse{
			StatusCode: 400,
			Message:    event.ErrInvalidCriteriaDeletion.Error(),
		}
	case event.ErrInvalidCriteriaAddition:
		return ErrorResponse{
			StatusCode: 400,
			Message:    event.ErrInvalidCriteriaAddition.Error(),
		}
	case event.ErrInvalidCriteriaEdit:
		return ErrorResponse{
			StatusCode: 400,
			Message:    event.ErrInvalidCriteriaEdit.Error(),
		}
	case event.ErrInvalidCriteriaMove:
		return ErrorResponse{
			StatusCode: 400,
			Message:    event.ErrInvalidCriteriaMove.Error(),
		}

	default:
		return ErrorResponse{
			StatusCode: 500,
			Message:    "Internal server error " + err.Error(),
		}
	}
}
