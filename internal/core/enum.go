package core

import (
	"fmt"
	"strings"
)

// EventStatus enum for events.status
type EventStatus string

// The event lifecycle:
//
//	draft -> registration_open <-> registration_closed -> preparing
//	      -> open <-> closed -> finalized
//
// Entries are registered during the registration_* states and chest numbers are
// drawn in one batch on registration_closed -> preparing.
const (
	EventStatusDraft              EventStatus = "draft"
	EventStatusRegistrationOpen   EventStatus = "registration_open"
	EventStatusRegistrationClosed EventStatus = "registration_closed"
	EventStatusPreparing          EventStatus = "preparing"
	EventStatusOpen               EventStatus = "open"
	EventStatusClosed             EventStatus = "closed"
	EventStatusFinalized          EventStatus = "finalized"
)

// AllEventStatuses is the lifecycle in order.
var AllEventStatuses = []EventStatus{
	EventStatusDraft,
	EventStatusRegistrationOpen,
	EventStatusRegistrationClosed,
	EventStatusPreparing,
	EventStatusOpen,
	EventStatusClosed,
	EventStatusFinalized,
}

func ParseEventStatus(v string) (EventStatus, error) {
	candidate := EventStatus(strings.ToLower(v))
	if candidate.IsValid() {
		return candidate, nil
	}
	names := make([]string, len(AllEventStatuses))
	for i, s := range AllEventStatuses {
		names[i] = string(s)
	}
	return "", fmt.Errorf("Invalid event status, must be one of [%s] ", strings.Join(names, ","))
}
func (s EventStatus) IsValid() bool {
	for _, known := range AllEventStatuses {
		if s == known {
			return true
		}
	}
	return false
}

// Category enum for events.category and students.category.
//
// CategoryCommon is event-only: a common event accepts students of every
// category, but no student belongs to it. ParseCategory and IsValid therefore
// reject it; events use ParseEventCategory and IsValidForEvent instead.
type Category string

const (
	CategoryHI     Category = "HI"
	CategoryMC     Category = "MC"
	CategoryPC     Category = "PC"
	CategoryCommon Category = "COMMON"
)

func ParseCategory(v string) (Category, error) {
	switch strings.ToUpper(v) {
	case "HI":
		return CategoryHI, nil
	case "MC":
		return CategoryMC, nil
	case "PC":
		return CategoryPC, nil
	default:
		return "", fmt.Errorf("Invalid category, must be ['HI','MC','PC'] ")
	}
}
func (c Category) IsValid() bool {
	switch c {
	case CategoryHI, CategoryMC, CategoryPC:
		return true
	}
	return false
}

// ParseEventCategory is ParseCategory plus the event-only CategoryCommon.
func ParseEventCategory(v string) (Category, error) {
	if strings.ToUpper(v) == string(CategoryCommon) {
		return CategoryCommon, nil
	}
	c, err := ParseCategory(v)
	if err != nil {
		return "", fmt.Errorf("Invalid category, must be ['HI','MC','PC','COMMON'] ")
	}
	return c, nil
}

// IsValidForEvent reports whether c can be assigned to an event.
func (c Category) IsValidForEvent() bool {
	return c == CategoryCommon || c.IsValid()
}

// AcceptsStudent reports whether an event of category c can register a student
// of the given category.
func (c Category) AcceptsStudent(student Category) bool {
	return c == CategoryCommon || c == student
}

// SchoolStatus enum for schools.status
//
//	in_review <-> accepted
//
// Schools registered through the public endpoint start in_review and their
// school admin cannot sign in until an admin accepts them.
type SchoolStatus string

const (
	SchoolStatusInReview SchoolStatus = "in_review"
	SchoolStatusAccepted SchoolStatus = "accepted"
)

func ParseSchoolStatus(v string) (SchoolStatus, error) {
	candidate := SchoolStatus(strings.ToLower(v))
	if !candidate.IsValid() {
		return "", fmt.Errorf("Invalid school status, must be ['in_review','accepted'] ")
	}
	return candidate, nil
}
func (s SchoolStatus) IsValid() bool {
	switch s {
	case SchoolStatusInReview, SchoolStatusAccepted:
		return true
	}
	return false
}
