package entry

import (
	"net/url"
	"strconv"

	errmap "github.com/nevinmanoj/bhavana-backend/internal/app/errmap"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/entry"
	"github.com/nevinmanoj/bhavana-backend/internal/util"
)

func parseEntryFilter(q url.Values) (entry.EntryFilter, *errmap.BadRequestError) {
	var f entry.EntryFilter

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

	// Pagination is opt-in: a caller that never sends limit/offset gets the
	// unbounded query it always got (f.Limit stays 0). Only present here so
	// this list can be paginated when a client actually asks for a page.
	if v := q.Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil {
			return f, &errmap.BadRequestError{
				Param:  "limit",
				Reason: err.Error(),
			}
		}
		if limit <= 0 || limit > 100 {
			limit = 100
		}
		f.Limit = int64(limit)
	}

	if v := q.Get("offset"); v != "" {
		offset, err := strconv.Atoi(v)
		if err != nil {
			return f, &errmap.BadRequestError{
				Param:  "offset",
				Reason: err.Error(),
			}
		}
		if offset < 0 {
			offset = 0
		}
		f.Offset = int64(offset)
	}

	return f, nil
}
