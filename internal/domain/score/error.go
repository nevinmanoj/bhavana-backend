package score

import (
	"errors"
	"strings"
)

var (
	ErrScoreNotFound      = errors.New("score not found")
	ErrInternal           = errors.New("Internal error")
	ErrUnauthorized       = errors.New("Unauthorized")
	ErrScoreAlreadyExists = errors.New("score already exists")
)
var (
	ErrEventNotOpen     = errors.New("event is not open for scoring")
	ErrScoreOutOfRange  = errors.New("score is out of allowed range")
	ErrNotAJudge        = errors.New("user is not a judge for this event")
	ErrCriteriaMismatch = errors.New("team does not belong to the same event as criteria")
)

func mapScoreError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "status is not open"):
		return ErrEventNotOpen
	case strings.Contains(msg, "out of range"):
		return ErrScoreOutOfRange
	case strings.Contains(msg, "is not a judge for this event"):
		return ErrNotAJudge
	case strings.Contains(msg, "does not belong to the same event as criteria"):
		return ErrCriteriaMismatch
	}

	return err
}
