package school

import (
	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
)

func GetSchoolDomainErrorResponse(err error) ErrorResponse {
	switch err {
	case school.ErrUnauthorized:
		return ErrorResponse{
			StatusCode: 403,
			Message:    school.ErrUnauthorized.Error(),
		}
	case school.ErrSchoolNotFound:
		return ErrorResponse{
			StatusCode: 404,
			Message:    school.ErrSchoolNotFound.Error(),
		}
	case school.ErrStudentNotFound:
		return ErrorResponse{
			StatusCode: 404,
			Message:    school.ErrStudentNotFound.Error(),
		}
	case school.ErrSchoolAdminAlreadyHasSchool:
		return ErrorResponse{
			StatusCode: 400,
			Message:    school.ErrSchoolAdminAlreadyHasSchool.Error(),
		}
	case school.ErrEmailTaken:
		return ErrorResponse{
			StatusCode: 400,
			Message:    school.ErrEmailTaken.Error(),
		}
	case school.ErrRegistrationClosed:
		return ErrorResponse{
			StatusCode: 403,
			Message:    school.ErrRegistrationClosed.Error(),
		}
	case school.ErrStudentRegistrationClosed:
		return ErrorResponse{
			StatusCode: 403,
			Message:    school.ErrStudentRegistrationClosed.Error(),
		}
	case school.ErrMailFailed:
		return ErrorResponse{
			StatusCode: 502,
			Message:    school.ErrMailFailed.Error(),
		}
	case school.ErrMailRateLimited:
		return ErrorResponse{
			StatusCode: 429,
			Message:    school.ErrMailRateLimited.Error(),
		}
	case school.ErrNoSchoolAdmin:
		return ErrorResponse{
			StatusCode: 400,
			Message:    school.ErrNoSchoolAdmin.Error(),
		}
	case school.ErrSchoolInReview:
		return ErrorResponse{
			StatusCode: 400,
			Message:    school.ErrSchoolInReview.Error(),
		}
	case school.ErrContactEmailLocked:
		return ErrorResponse{
			StatusCode: 400,
			Message:    school.ErrContactEmailLocked.Error(),
		}
	case school.ErrLockedByFinalizedEvent:
		return ErrorResponse{
			StatusCode: 409,
			Message:    school.ErrLockedByFinalizedEvent.Error(),
		}

	default:
		return ErrorResponse{
			StatusCode: 500,
			Message:    "Internal server error " + err.Error(),
		}
	}
}
