package errmap

import (
	. "github.com/nevinmanoj/bhavana-backend/api"
	user "github.com/nevinmanoj/bhavana-backend/internal/domain/user"
)

func GetDomainErrorResponse(err error) ErrorResponse {
	switch err {
	//user errrors
	case user.ErrUnauthorized:
		return ErrorResponse{
			StatusCode: 403,
			Message:    "Unauthorized to access user",
		}
	case user.ErrNotFound:
		return ErrorResponse{
			StatusCode: 404,
			Message:    "User not found",
		}
	case user.ErrAlreadyExists:
		return ErrorResponse{
			StatusCode: 400,
			Message:    "User already exists",
		}

	default:
		return ErrorResponse{
			StatusCode: 500,
			Message:    "Internal server error" + err.Error(),
		}
	}
}
