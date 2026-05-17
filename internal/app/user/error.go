package user

import (
	. "github.com/nevinmanoj/bhavana-backend/api"
	user "github.com/nevinmanoj/bhavana-backend/internal/domain/user"
)

func GetUserDomainErrorResponse(err error) ErrorResponse {
	switch err {
	//user errrors
	case user.ErrUnauthorized:
		return ErrorResponse{
			StatusCode: 403,
			Message:    user.ErrUnauthorized.Error(),
		}
	case user.ErrNotFound:
		return ErrorResponse{
			StatusCode: 404,
			Message:    user.ErrNotFound.Error(),
		}
	case user.ErrAlreadyExists:
		return ErrorResponse{
			StatusCode: 400,
			Message:    user.ErrAlreadyExists.Error(),
		}
	case user.ErrInvalidCredentials:
		return ErrorResponse{
			StatusCode: 401,
			Message:    user.ErrInvalidCredentials.Error(),
		}

	default:
		return ErrorResponse{
			StatusCode: 500,
			Message:    "Internal server error " + err.Error(),
		}
	}
}
