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
)
