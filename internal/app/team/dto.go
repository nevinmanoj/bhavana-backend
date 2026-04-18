package team

import (
	"time"

	"github.com/nevinmanoj/bhavana-backend/internal/domain/team"
)

// requests
type CreateTeamRequest struct {
	EventID  int64               `json:"event_id" validate:"required"`
	SchoolID int64               `json:"school_id" validate:"required"`
	Members  []TeamMemberRequest `json:"members"`
}

type TeamMemberRequest struct {
	StudentID int64 `json:"student_id"`
}

type UpdateTeamRequest struct {
	ID int64 `json:"id"`
	CreateTeamRequest
}

// responses
type TeamMembersResponse struct {
	Name      string `json:"name"`
	StudentID int64  `json:"student_id"`
}

type TeamFullResponse struct {
	ID          int64                 `json:"id"`
	EventID     int64                 `json:"event_id"`
	SchoolID    int64                 `json:"school_id"`
	ChestNumber int                   `json:"chest_number"`
	CreatedAt   time.Time             `json:"created_at"`
	Members     []TeamMembersResponse `json:"members"`
}
type TeamResponseJudge struct {
	ID          int64 `json:"id"`
	EventID     int64 `json:"event_id"`
	ChestNumber int   `json:"chest_number"`
}

func ToTeamFullResponse(team *team.TeamFull) TeamFullResponse {
	members := make([]TeamMembersResponse, len(team.Members))
	for i, member := range team.Members {
		members[i] = TeamMembersResponse{
			Name:      member.Name,
			StudentID: member.StudentID,
		}
	}
	return TeamFullResponse{
		ID:          team.ID,
		EventID:     team.EventID,
		SchoolID:    team.SchoolID,
		ChestNumber: team.ChestNumber,
		CreatedAt:   team.CreatedAt,
		Members:     members,
	}
}
func ToTeamResponseJudge(team *team.TeamFull) TeamResponseJudge {
	return TeamResponseJudge{
		ID:          team.ID,
		EventID:     team.EventID,
		ChestNumber: team.ChestNumber,
	}
}
