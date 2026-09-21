package entry

import "errors"

var (
	ErrUnauthorized           = errors.New("Unauthorized")
	ErrEntryNotFound          = errors.New("Entry not found")
	ErrEntryMemberNotFound    = errors.New("Entry member not found")
	ErrInternal               = errors.New("Internal server error")
	ErrInvalidEntryUpdate     = errors.New("Cannot change event_id or school_id for an entry")
	ErrEntryCountExceedsLimit = errors.New("Entry count for school exceeds the limit for the event")
	ErrMemberCountExceedsLimit = errors.New("Member count exceeds the limit for the event")
	ErrMemberCountBelowMinimum = errors.New("Member count is below the minimum for the event")
	ErrSchoolMismatch         = errors.New("School of student does not match school of entry")
	ErrCategoryMismatch       = errors.New("Category of student does not match category of event")
	ErrStudentAlreadyInEntry  = errors.New("Student is already in an entry for this event")
)
