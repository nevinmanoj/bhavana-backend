package score

import (
	"errors"
)

var (
	ErrScoreNotFound = errors.New("score not found")
	ErrInternal      = errors.New("Internal error")
	ErrUnauthorized  = errors.New("Unauthorized")
)
var (
	ErrEventNotOpen     = errors.New("evt is not openen for scoring")
	ErrScoreOutOfRange  = errors.New("score is out of allowed range")
	ErrNotAJudge        = errors.New("user is not a judge for this event")
	ErrCriteriaMismatch = errors.New("team does not belong to the same event as criteria")
	ErrAlreadyExists    = errors.New("score already exists for this team, judge and criteria")
)
