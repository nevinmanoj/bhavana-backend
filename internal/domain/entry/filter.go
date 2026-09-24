package entry

import "github.com/nevinmanoj/bhavana-backend/internal/core"

type EntryFilter struct {
	EventID  *int64
	SchoolID *int64
	Category *core.Category
	// Limit/Offset are 0 by default, meaning "unbounded" - the query builder
	// only appends LIMIT/OFFSET when Limit > 0. Internal callers that need
	// every matching row (e.g. counting existing entries against
	// max_entries_per_school) build an EntryFilter without ever setting
	// these, so they are unaffected by pagination.
	Limit  int64
	Offset int64
}
