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
