package team

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/team"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

type eventRepository struct {
}

func NewTeamWriteRepository() team.TeamWriteRepository {
	return &eventRepository{}
}
func NewTeamReadRepository() team.TeamReadRepository {
	return &eventRepository{}
}

// teams
func (e *eventRepository) GetAllTeams(ctx context.Context, db sqlx.ExtContext, filter team.TeamFilter) ([]team.TeamFull, error) {
	scope := ctx.Value(middleware.ContextScope).(rbac.Scope)
	role := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	teams := []team.TeamFull{}
	baseQuery := `SELECT   
	t.id AS "team.id",
	t.chest_number AS "team.chest_number",
    t.school_id  AS "team.school_id",
    t.event_id   AS "team.event_id",
    t.created_at AS "team.created_at",
	s.name AS school_name,
	s.address AS school_address,
	e.title AS event_title,
	e.category AS category 
	FROM teams t
	JOIN schools s ON s.id = t.school_id
	JOIN events e ON e.id = t.event_id`
	args := []any{}
	conditions := []string{}
	if scope.UserID != nil && role == rbac.UserRoleJudge {
		baseQuery += " JOIN event_judges ej ON ej.event_id = e.id "
		conditions = append(conditions, "ej.user_id = ?")
		args = append(args, *scope.UserID)
	}
	finalQuery, finalargs, err := buildTeamQuery(baseQuery, conditions, args, filter)
	if err != nil {
		return nil, err
	}
	err = sqlx.SelectContext(
		ctx, db,
		&teams,
		finalQuery, finalargs...,
	)
	if err != nil {
		return nil, err
	}
	return teams, nil
}
func (e *eventRepository) GetTeamByID(ctx context.Context, db sqlx.ExtContext, teamId int64) (*team.TeamFull, error) {
	teams := []team.TeamFull{}
	baseQuery := `SELECT   
	t.id AS "team.id",
    t.school_id  AS "team.school_id",
    t.event_id   AS "team.event_id",
    t.created_at AS "team.created_at",
	s.name AS school_name,
	s.address AS school_address,
	e.title AS event_title,
	e.category AS category 
	FROM teams t
	JOIN schools s ON s.id = t.school_id
	JOIN events e ON e.id = t.event_id
	WHERE t.id = $1`

	err := sqlx.SelectContext(
		ctx, db,
		&teams,
		baseQuery, teamId,
	)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, fmt.Errorf("Team with id %d not found", teamId)
	}
	return &teams[0], nil
}
func (e *eventRepository) CreateTeam(ctx context.Context, db sqlx.ExtContext, teamToCreate *team.Team) error {
	query := `
		INSERT INTO teams (
			event_id,
			school_id
		)
		VALUES (
			:event_id,
			:school_id
		)
		RETURNING id,chest_number, created_at
	`

	rows, err := sqlx.NamedQueryContext(ctx, db, query, teamToCreate)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&teamToCreate.ID, &teamToCreate.ChestNumber, &teamToCreate.CreatedAt)
		return nil
	}

	return sql.ErrNoRows
}
func (e *eventRepository) DeleteTeam(ctx context.Context, db sqlx.ExtContext, teamID int64) error {
	query := `DELETE FROM teams WHERE id = $1`
	_, err := db.ExecContext(ctx, query, teamID)
	return err
}

// team memebers
func (e *eventRepository) GetTeamMembers(ctx context.Context, db sqlx.ExtContext, teamId int64) ([]team.TeamMember, error) {
	teamMembers := []team.TeamMember{}
	baseQuery := `SELECT   
	s.name,tm.team_id,tm.student_id,tm.created_at
	FROM team_members tm
	JOIN students s ON s.id = tm.student_id
	WHERE tm.team_id = $1`
	err := sqlx.SelectContext(
		ctx, db,
		&teamMembers,
		baseQuery, teamId,
	)
	if err != nil {
		return nil, err
	}
	return teamMembers, nil
}
func (e *eventRepository) CreateTeamMember(ctx context.Context, db sqlx.ExtContext, teamMemberToCreate *team.TeamMember) error {
	query := `
		INSERT INTO team_members (
			team_id,
			student_id
		)
		VALUES (
			:team_id,
			:student_id
		)
		RETURNING created_at
	`

	rows, err := sqlx.NamedQueryContext(ctx, db, query, teamMemberToCreate)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&teamMemberToCreate.CreatedAt)
		return nil
	}

	return sql.ErrNoRows
}
func (e *eventRepository) DeleteTeamMember(ctx context.Context, db sqlx.ExtContext, teamID int64, studentID int64) error {
	query := `DELETE FROM team_members 
	WHERE team_id = $1 
	AND student_id = $2`
	_, err := db.ExecContext(ctx, query, teamID, studentID)
	return err
}
