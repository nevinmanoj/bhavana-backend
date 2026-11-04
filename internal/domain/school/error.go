package school

import (
	"errors"
)

var (
	ErrSchoolNotFound              = errors.New("school not found")
	ErrStudentNotFound             = errors.New("student not found")
	ErrInternal                    = errors.New("Internal error")
	ErrUnauthorized                = errors.New("Unauthorized")
	ErrSchoolAdminAlreadyHasSchool = errors.New("school admin already has a school")
	ErrLockedByFinalizedEvent      = errors.New("cannot delete: it has entries, scores or results in a finalized event")
	ErrEmailTaken                  = errors.New("An account with this email already exists.")
	ErrRegistrationClosed          = errors.New("School registration is currently closed.")
	ErrStudentRegistrationClosed   = errors.New("Student registration isn't open yet.")
	ErrMailFailed                  = errors.New("Couldn't send the email, so nothing was changed. Please try again.")
	ErrMailRateLimited             = errors.New("Too many emails were sent recently. Nothing was changed. Try again in a few minutes.")
	ErrNoSchoolAdmin               = errors.New("This school has no school admin account to send login details to.")
	ErrSchoolInReview              = errors.New("Accept the registration first; that sends the login details.")
	ErrContactEmailLocked          = errors.New("This school's email is also its sign-in email. Use Send login details to change it.")
)
