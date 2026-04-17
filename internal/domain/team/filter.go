package team

import "github.com/nevinmanoj/bhavana-backend/internal/core"

type TeamFilter struct {
	EventID  *int64
	SchoolID *int64
	Category *core.Category
}
