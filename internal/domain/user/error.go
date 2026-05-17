package user

import (
	"errors"
)

var (
	ErrNotFound           = errors.New("User not found")
	ErrInternal           = errors.New("Internal error")
	ErrUnauthorized       = errors.New("Unauthorized")
	ErrInvalidCredentials = errors.New("Invalid credentials")
	ErrAlreadyExists      = errors.New("User already exists")
)
