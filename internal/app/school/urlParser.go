package school

import (
	"net/url"

	errmap "github.com/nevinmanoj/bhavana-backend/internal/app/errmap"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
	"github.com/nevinmanoj/bhavana-backend/internal/util"
)

func parseStudentFilter(q url.Values) (school.StudentFilter, *errmap.BadRequestError) {
	var f school.StudentFilter

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
		schoolId, err := util.ParseStrToInt64(v)
		if err != nil {
			return f, &errmap.BadRequestError{
				Param:  "school_id",
				Reason: err.Error(),
			}
		}
		f.SchoolID = schoolId
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
