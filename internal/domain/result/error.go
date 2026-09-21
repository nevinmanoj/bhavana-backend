package result

import "errors"

var (
	ErrInternal     = errors.New("Internal error")
	ErrUnauthorized = errors.New("Unauthorized")
)
