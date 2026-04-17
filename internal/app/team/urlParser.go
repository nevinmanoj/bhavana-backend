package team

import (
	"net/url"

	errmap "github.com/nevinmanoj/bhavana-backend/internal/app/errmap"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/team"
	"github.com/nevinmanoj/bhavana-backend/internal/util"
)

func parseTeamFilter(q url.Values) (team.TeamFilter, *errmap.BadRequestError) {
	var f team.TeamFilter

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
	if v := q.Get("school_id"); v != "" {
		schoolID, err := util.ParseStrToInt64(v)
		if err != nil {
			return f, &errmap.BadRequestError{
				Param:  "school_id",
				Reason: err.Error(),
			}
		}
		f.SchoolID = schoolID
	}
	if v := q.Get("event_id"); v != "" {
		eventID, err := util.ParseStrToInt64(v)
		if err != nil {
			return f, &errmap.BadRequestError{
				Param:  "event_id",
				Reason: err.Error(),
			}
		}
		f.EventID = eventID
	}

	// // Pagination defaults
	// f.Limit = 100
	// f.Offset = 0

	// if v := q.Get("limit"); v != "" {
	// 	limit, err := strconv.Atoi(v)
	// 	if err != nil {
	// 		return f, &errMap.BadRequestError{
	// 			Param:  "limit",
	// 			Reason: err.Error(),
	// 		}
	// 	} else if limit > 0 && limit < 100 {
	// 		f.Limit = limit
	// 	}
	// }

	// if v := q.Get("offset"); v != "" {
	// 	offset, err := strconv.Atoi(v)
	// 	if err != nil {
	// 		return f, &errMap.BadRequestError{
	// 			Param:  "offset",
	// 			Reason: err.Error(),
	// 		}
	// 	} else if offset > 0 {
	// 		f.Offset = offset
	// 	}
	// }

	return f, nil
}
