package school

import (
	"github.com/lib/pq"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
)

func errorMapper(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "23505":
			//constraint violation
			switch pqErr.Constraint {
			case "schools_school_admin_key":
				return school.ErrSchoolAdminAlreadyHasSchool
			}
		default:
			return school.ErrInternal
		}
	}

	return err
}
