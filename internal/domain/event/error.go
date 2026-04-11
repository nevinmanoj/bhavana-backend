package event

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("Event not found")
	ErrInternal      = errors.New("Internal error")
	ErrUnauthorized  = errors.New("Unauthorized")
	ErrAlreadyExists = errors.New("Event already exists")
)
