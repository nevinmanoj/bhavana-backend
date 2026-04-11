package event

import core "github.com/nevinmanoj/bhavana-backend/internal/core"

type EventFilter struct {
	Status   *core.EventStatus
	Category *core.Category
}
