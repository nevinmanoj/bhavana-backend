package school

import core "github.com/nevinmanoj/bhavana-backend/internal/core"

type StudentFilter struct {
	SchoolID *int64
	Category *core.Category
	// Limit/Offset are 0 by default, meaning "unbounded" - the query builder
	// only appends LIMIT/OFFSET when Limit > 0. GetStudentsBySchoolID never
	// sets these (only the global /students list does), so a single
	// school's roster is unaffected by pagination.
	Limit  int64
	Offset int64
}
