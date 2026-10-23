package entry

import (
	"github.com/lib/pq"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/entry"
)

func errorMapper(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "P0401":
			return entry.ErrInvalidEntryUpdate
		case "P0402":
			return entry.ErrInvalidEntryUpdate
		case "P0403":
			return entry.ErrEntryCountExceedsLimit
		case "P0404":
			return entry.ErrSchoolMismatch
		case "P0405":
			return entry.ErrCategoryMismatch
		case "P0406":
			return entry.ErrStudentAlreadyInEntry
		case "P0407":
			return entry.ErrMemberCountBelowMinimum
		case "P0408":
			return entry.ErrMemberCountExceedsLimit
		case "P0409":
			return entry.ErrRegistrationNotOpen
		//finalized-event locks (013_lock_finalized_events.sql) - an entry
		//delete can trip any of them through its cascade
		case "P0410", "P0411", "P0505", "P0603":
			return entry.ErrEventFinalized
		case "23505":
			if pqErr.Constraint == "uniq_event_chest_number" {
				return entry.ErrChestNumberConflict
			}
			return entry.ErrStudentAlreadyInEntry
		default:
			return entry.ErrInternal
		}
	}

	return err
}
