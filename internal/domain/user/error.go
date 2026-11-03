package user

import (
	"errors"
)

var (
	ErrNotFound                = errors.New("User not found")
	ErrInternal                = errors.New("Internal error")
	ErrUnauthorized            = errors.New("Unauthorized")
	ErrInvalidCredentials      = errors.New("Invalid credentials")
	ErrAlreadyExists           = errors.New("User already exists")
	ErrRegistrationUnderReview = errors.New("Your school's registration is under review. We'll email your password once it's accepted.")
	ErrInvalidResetToken       = errors.New("This link has expired or was already used. Ask for a new one from the sign-in page.")
	ErrWrongCurrentPassword    = errors.New("Your current password is incorrect.")
)
