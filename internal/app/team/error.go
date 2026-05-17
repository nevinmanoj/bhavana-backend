package team

import (
	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/team"
)

func GetTeamDomainErrorResponse(err error) ErrorResponse {
	switch err {
	//user errrors
	case team.ErrUnauthorized:
		return ErrorResponse{
			StatusCode: 403,
			Message:    team.ErrUnauthorized.Error(),
		}
	case team.ErrTeamNotFound:
		return ErrorResponse{
			StatusCode: 404,
			Message:    team.ErrTeamNotFound.Error(),
		}
	case team.ErrTeamMemberNotFound:
		return ErrorResponse{
			StatusCode: 404,
			Message:    team.ErrTeamMemberNotFound.Error(),
		}
	case team.ErrInvalidTeamUpdate:
		return ErrorResponse{
			StatusCode: 400,
			Message:    team.ErrInvalidTeamUpdate.Error(),
		}
	case team.ErrTeamCountExceedsLimit:
		return ErrorResponse{
			StatusCode: 400,
			Message:    team.ErrTeamCountExceedsLimit.Error(),
		}
	case team.ErrTeamSizeExceedsLimit:
		return ErrorResponse{
			StatusCode: 400,
			Message:    team.ErrTeamSizeExceedsLimit.Error(),
		}
	case team.ErrTeamSizeBelowMinimum:
		return ErrorResponse{
			StatusCode: 400,
			Message:    team.ErrTeamSizeBelowMinimum.Error(),
		}
	case team.ErrSchoolMismatch:
		return ErrorResponse{
			StatusCode: 400,
			Message:    team.ErrSchoolMismatch.Error(),
		}
	case team.ErrCategoryMismatch:
		return ErrorResponse{
			StatusCode: 400,
			Message:    team.ErrCategoryMismatch.Error(),
		}
	case team.ErrStudentAlreadyInTeam:
		return ErrorResponse{
			StatusCode: 400,
			Message:    team.ErrStudentAlreadyInTeam.Error(),
		}

	default:
		return ErrorResponse{
			StatusCode: 500,
			Message:    "Internal server error " + err.Error(),
		}
	}
}
