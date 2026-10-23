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
		//finalized-event locks (013_lock_finalized_events.sql), reached
		//through the cascade from a school or student delete
		case "P0410", "P0411", "P0505", "P0603":
			return school.ErrLockedByFinalizedEvent
		default:
			return school.ErrInternal
		}
	}

	return err
}
