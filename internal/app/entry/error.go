package entry

import (
	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/entry"
)

func GetEntryDomainErrorResponse(err error) ErrorResponse {
	switch err {
	//user errrors
	case entry.ErrUnauthorized:
		return ErrorResponse{
			StatusCode: 403,
			Message:    entry.ErrUnauthorized.Error(),
		}
	case entry.ErrEntryNotFound:
		return ErrorResponse{
			StatusCode: 404,
			Message:    entry.ErrEntryNotFound.Error(),
		}
	case entry.ErrEntryMemberNotFound:
		return ErrorResponse{
			StatusCode: 404,
			Message:    entry.ErrEntryMemberNotFound.Error(),
		}
	case entry.ErrInvalidEntryUpdate:
		return ErrorResponse{
			StatusCode: 400,
			Message:    entry.ErrInvalidEntryUpdate.Error(),
		}
	case entry.ErrEntryCountExceedsLimit:
		return ErrorResponse{
			StatusCode: 400,
			Message:    entry.ErrEntryCountExceedsLimit.Error(),
		}
	case entry.ErrMemberCountExceedsLimit:
		return ErrorResponse{
			StatusCode: 400,
			Message:    entry.ErrMemberCountExceedsLimit.Error(),
		}
	case entry.ErrMemberCountBelowMinimum:
		return ErrorResponse{
			StatusCode: 400,
			Message:    entry.ErrMemberCountBelowMinimum.Error(),
		}
	case entry.ErrSchoolMismatch:
		return ErrorResponse{
			StatusCode: 400,
			Message:    entry.ErrSchoolMismatch.Error(),
		}
	case entry.ErrCategoryMismatch:
		return ErrorResponse{
			StatusCode: 400,
			Message:    entry.ErrCategoryMismatch.Error(),
		}
	case entry.ErrStudentAlreadyInEntry:
		return ErrorResponse{
			StatusCode: 400,
			Message:    entry.ErrStudentAlreadyInEntry.Error(),
		}
	case entry.ErrRegistrationNotOpen:
		return ErrorResponse{
			StatusCode: 409,
			Message:    entry.ErrRegistrationNotOpen.Error(),
		}
	case entry.ErrChestNumberConflict:
		return ErrorResponse{
			StatusCode: 409,
			Message:    entry.ErrChestNumberConflict.Error(),
		}
	case entry.ErrEmptyBatch:
		return ErrorResponse{
			StatusCode: 400,
			Message:    entry.ErrEmptyBatch.Error(),
		}

	default:
		return ErrorResponse{
			StatusCode: 500,
			Message:    "Internal server error " + err.Error(),
		}
	}
}
