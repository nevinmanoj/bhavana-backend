package school

import core "github.com/nevinmanoj/bhavana-backend/internal/core"

type StudentFilter struct {
	SchoolID *int64
	Category *core.Category
}
