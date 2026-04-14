package school

import (
	"errors"
)

var (
	ErrSchoolNotFound       = errors.New("school not found")
	ErrStudentNotFound      = errors.New("student not found")
	ErrInternal             = errors.New("Internal error")
	ErrUnauthorized         = errors.New("Unauthorized")
	ErrSchoolAlreadyExists  = errors.New("school already exists")
	ErrStudentAlreadyExists = errors.New("student already exists")
)
