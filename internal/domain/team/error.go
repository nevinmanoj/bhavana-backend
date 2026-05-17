package team

import "errors"

var (
	ErrUnauthorized          = errors.New("Unauthorized")
	ErrTeamNotFound          = errors.New("Team not found")
	ErrTeamMemberNotFound    = errors.New("Team member not found")
	ErrInternal              = errors.New("Internal server error")
	ErrInvalidTeamUpdate     = errors.New("Cannot change event_id or school_id for a team")
	ErrTeamCountExceedsLimit = errors.New("Team count for school exceeds the limit for the event")
	ErrTeamSizeExceedsLimit  = errors.New("Team size exceeds the limit for the event")
	ErrTeamSizeBelowMinimum  = errors.New("Team size is below the minimum for the event")
	ErrSchoolMismatch        = errors.New("School of student does not match school of team")
	ErrCategoryMismatch      = errors.New("Category of student does not match category of event")
	ErrStudentAlreadyInTeam  = errors.New("Student is already in a team for this event")
)
