package core

import (
	"fmt"
	"strings"
)

// UserRole Enum for users.role
type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleJudge UserRole = "judge"
)

func ParseUserRole(v string) (UserRole, error) {
	switch strings.ToLower(v) {
	case "admin":
		return UserRoleAdmin, nil
	case "judge":
		return UserRoleJudge, nil
	default:
		return "", fmt.Errorf("Invalid role, must be ['admin','judge'] ")
	}
}
func (r UserRole) IsValid() bool {
	switch r {
	case UserRoleAdmin, UserRoleJudge:
		return true
	}
	return false
}

// EventStatus enum for events.status
type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusOpen      EventStatus = "open"
	EventStatusClosed    EventStatus = "closed"
	EventStatusFinalized EventStatus = "finalized"
)

func ParseEventStatus(v string) (EventStatus, error) {
	switch strings.ToLower(v) {
	case "draft":
		return EventStatusDraft, nil
	case "open":
		return EventStatusOpen, nil
	case "closed":
		return EventStatusClosed, nil
	case "finalized":
		return EventStatusFinalized, nil
	default:
		return "", fmt.Errorf("Invalid event status, must be ['draft','open','closed','finalized'] ")
	}
}
func (s EventStatus) IsValid() bool {
	switch s {
	case EventStatusDraft, EventStatusOpen, EventStatusClosed, EventStatusFinalized:
		return true
	}
	return false
}

// Category enum for events.category and students.category
type Category string

const (
	CategoryHC Category = "HC"
	CategoryMC Category = "MC"
	CategoryPC Category = "PC"
)

func ParseCategory(v string) (Category, error) {
	switch strings.ToUpper(v) {
	case "HC":
		return CategoryHC, nil
	case "MC":
		return CategoryMC, nil
	case "PC":
		return CategoryPC, nil
	default:
		return "", fmt.Errorf("Invalid category, must be ['HC','MC','PC'] ")
	}
}
func (c Category) IsValid() bool {
	switch c {
	case CategoryHC, CategoryMC, CategoryPC:
		return true
	}
	return false
}
