package entry

import (
	"time"

	"github.com/nevinmanoj/bhavana-backend/internal/domain/entry"
)

// requests
type CreateEntryRequest struct {
	EventID  int64                `json:"event_id" validate:"required"`
	SchoolID int64                `json:"school_id" validate:"required"`
	Members  []EntryMemberRequest `json:"members"`
}

type EntryMemberRequest struct {
	StudentID int64 `json:"student_id"`
}

type UpdateEntryRequest struct {
	ID int64 `json:"id"`
	CreateEntryRequest
}

// responses
type EntryMemberResponse struct {
	Name      string `json:"name"`
	StudentID int64  `json:"student_id"`
}

type EntryFullResponse struct {
	ID          int64                  `json:"id"`
	EventID     int64                  `json:"event_id"`
	SchoolID    int64                  `json:"school_id"`
	ChestNumber int                    `json:"chest_number"`
	CreatedAt   time.Time              `json:"created_at"`
	Members     []EntryMemberResponse  `json:"members"`
}
type EntryResponseJudge struct {
	ID          int64 `json:"id"`
	EventID     int64 `json:"event_id"`
	ChestNumber int   `json:"chest_number"`
}

func ToEntryFullResponse(entry *entry.EntryFull) EntryFullResponse {
	members := make([]EntryMemberResponse, len(entry.Members))
	for i, member := range entry.Members {
		members[i] = EntryMemberResponse{
			Name:      member.Name,
			StudentID: member.StudentID,
		}
	}
	return EntryFullResponse{
		ID:          entry.ID,
		EventID:     entry.EventID,
		SchoolID:    entry.SchoolID,
		ChestNumber: entry.ChestNumber,
		CreatedAt:   entry.CreatedAt,
		Members:     members,
	}
}
func ToEntryResponseJudge(entry *entry.EntryFull) EntryResponseJudge {
	return EntryResponseJudge{
		ID:          entry.ID,
		EventID:     entry.EventID,
		ChestNumber: entry.ChestNumber,
	}
}
