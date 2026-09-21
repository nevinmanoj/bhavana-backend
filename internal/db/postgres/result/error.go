package result

import (
	"github.com/lib/pq"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/result"
)

func errorMapper(err error) error {
	if _, ok := err.(*pq.Error); ok {
		// P0601 (results require a finalized event) and P0602 (results are
		// immutable) should never actually fire here: GenerateForEvent only
		// runs inside the finalize transaction after the status write, and
		// results are only ever inserted, never updated. Fall through to a
		// generic internal error rather than exposing raw SQLSTATEs.
		return result.ErrInternal
	}

	return err
}
