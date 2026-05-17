package score

import (
	"github.com/lib/pq"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/score"
)

func errorMapper(err error) error {
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "23505":
			return score.ErrAlreadyExists
		case "P0501":
			return score.ErrEventNotOpen
		case "P0502":
			return score.ErrScoreOutOfRange
		case "P0503":
			return score.ErrNotAJudge
		case "P0504":
			return score.ErrCriteriaMismatch
		default:
			return score.ErrInternal
		}
	}

	return err
}
