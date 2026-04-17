package team

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type TeamWriteRepository interface {
	TeamReadRepository
	CreateTeam(ctx context.Context, db sqlx.ExtContext, teamToCreate *Team) error
	DeleteTeam(ctx context.Context, db sqlx.ExtContext, teamID int64) error

	CreateTeamMember(ctx context.Context, db sqlx.ExtContext, teamMemberToCreate *TeamMember) error
	DeleteTeamMember(ctx context.Context, db sqlx.ExtContext, teamID, studentId int64) error
}
type TeamReadRepository interface {
	GetAllTeams(ctx context.Context, db sqlx.ExtContext, filter TeamFilter) ([]TeamFull, error)
	GetTeamByID(ctx context.Context, db sqlx.ExtContext, teamId int64) (*TeamFull, error)

	GetTeamMembers(ctx context.Context, db sqlx.ExtContext, teamId int64) ([]TeamMember, error)
}
