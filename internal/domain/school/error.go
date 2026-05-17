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
)
