package result

import (
	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/result"
)

func GetResultDomainErrorResponse(err error) ErrorResponse {
	switch err {
	case result.ErrUnauthorized:
		return ErrorResponse{
			StatusCode: 403,
			Message:    result.ErrUnauthorized.Error(),
		}
	default:
		return ErrorResponse{
			StatusCode: 500,
			Message:    "Internal server error " + err.Error(),
		}
	}
}
