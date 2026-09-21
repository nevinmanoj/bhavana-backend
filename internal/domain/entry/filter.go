package entry

import "github.com/nevinmanoj/bhavana-backend/internal/core"

type EntryFilter struct {
	EventID  *int64
	SchoolID *int64
	Category *core.Category
}
