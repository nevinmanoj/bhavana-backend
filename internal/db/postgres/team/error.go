package team

import (
	"github.com/lib/pq"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/team"
)

func errorMapper(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "P0401":
			return team.ErrInvalidTeamUpdate
		case "P0402":
			return team.ErrInvalidTeamUpdate
		case "P0403":
			return team.ErrTeamCountExceedsLimit
		case "P0404":
			return team.ErrSchoolMismatch
		case "P0405":
			return team.ErrCategoryMismatch
		case "P0406":
			return team.ErrStudentAlreadyInTeam
		case "P0407":
			return team.ErrTeamSizeBelowMinimum
		case "P0408":
			return team.ErrTeamSizeExceedsLimit
		default:
			return team.ErrInternal
		}
	}

	return err
}
