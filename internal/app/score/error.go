package score

import (
	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/score"
)

func GetScoreDomainErrorResponse(err error) ErrorResponse {
	switch err {
	//user errrors
	case score.ErrUnauthorized:
		return ErrorResponse{
			StatusCode: 403,
			Message:    score.ErrUnauthorized.Error(),
		}
	case score.ErrScoreNotFound:
		return ErrorResponse{
			StatusCode: 404,
			Message:    score.ErrScoreNotFound.Error(),
		}
	case score.ErrAlreadyExists:
		return ErrorResponse{
			StatusCode: 400,
			Message:    score.ErrAlreadyExists.Error(),
		}
	case score.ErrEventNotOpen:
		return ErrorResponse{
			StatusCode: 400,
			Message:    score.ErrEventNotOpen.Error(),
		}
	case score.ErrScoreOutOfRange:
		return ErrorResponse{
			StatusCode: 400,
			Message:    score.ErrScoreOutOfRange.Error(),
		}
	case score.ErrNotAJudge:
		return ErrorResponse{
			StatusCode: 400,
			Message:    score.ErrNotAJudge.Error(),
		}
	case score.ErrCriteriaMismatch:
		return ErrorResponse{
			StatusCode: 400,
			Message:    score.ErrCriteriaMismatch.Error(),
		}
	case score.ErrAlreadyExists:
		return ErrorResponse{
			StatusCode: 400,
			Message:    score.ErrAlreadyExists.Error(),
		}

	default:
		return ErrorResponse{
			StatusCode: 500,
			Message:    "Internal server error " + err.Error(),
		}
	}
}
