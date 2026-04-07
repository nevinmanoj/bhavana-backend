package errmap

import (
	"fmt"

	. "github.com/nevinmanoj/bhavana-backend/api"
)

type BadRequestError struct {
	Param  string
	Reason string
}

func (e BadRequestError) Error() string {
	return fmt.Sprintf("invalid %s: %s", e.Param, e.Reason)
}

func GetHttpErrorResponse(err any) ErrorResponse {
	switch e := err.(type) {

	case *BadRequestError:
		return ErrorResponse{
			StatusCode: 400,
			Message:    e.Error(),
		}

	default:
		return ErrorResponse{
			StatusCode: 500,
			Message:    "internal server error",
		}
	}
}
