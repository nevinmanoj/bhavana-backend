package result

import (
	"net/url"

	errmap "github.com/nevinmanoj/bhavana-backend/internal/app/errmap"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/result"
)

func parseLeaderboardFilter(q url.Values) (result.LeaderboardFilter, *errmap.BadRequestError) {
	var f result.LeaderboardFilter

	if v := q.Get("category"); v != "" {
		category, err := core.ParseCategory(v)
		if err != nil {
			return f, &errmap.BadRequestError{
				Param:  "category",
				Reason: err.Error(),
			}
		}
		f.Category = &category
	}

	return f, nil
}
