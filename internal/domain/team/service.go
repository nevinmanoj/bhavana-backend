package team

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/access"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/event"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
)

type teamService struct {
	db            *sqlx.DB
	accessService access.AccessService
	repo          TeamWriteRepository
	eventsRepo    event.EventReadRepository
	schoolRepo    school.SchoolReadRepository
}

type TeamService interface {
	GetTeams(ctx context.Context, filter TeamFilter) ([]TeamFull, error)
	GetTeamsByID(ctx context.Context, teamID int64) (*TeamFull, error)
	CreateTeam(ctx context.Context, teamToCreate *TeamFull) error
	UpdateTeam(ctx context.Context, teamToUpdate *TeamFull) error
	DeleteTeam(ctx context.Context, teamID int64) error
}

func NewTeamService(
	db *sqlx.DB,
	accessService access.AccessService,
	repo TeamWriteRepository,
	eventsRepo event.EventReadRepository,
	schoolRepo school.SchoolReadRepository,
) TeamService {
	return &teamService{
		db:            db,
		accessService: accessService,
		repo:          repo,
		eventsRepo:    eventsRepo,
		schoolRepo:    schoolRepo}
}

func (s *teamService) GetTeams(ctx context.Context, filter TeamFilter) ([]TeamFull, error) {
	teamsFull := []TeamFull{}
	teams, err := s.repo.GetAllTeams(ctx, s.db, filter)
	if err != nil || teams == nil || len(teams) == 0 {
		return []TeamFull{}, err
	}
	for _, team := range teams {
		teamMembers, err := s.repo.GetTeamMembers(ctx, s.db, team.ID)
		if err != nil {
			return []TeamFull{}, err
		}
		teamsFull = append(teamsFull, TeamFull{
			Team:    team.Team,
			Members: teamMembers,
		})
	}
	return teamsFull, nil
}

func (s *teamService) GetTeamsByID(ctx context.Context, teamID int64) (*TeamFull, error) {
	team, err := s.repo.GetTeamByID(ctx, s.db, teamID)
	if err != nil || team == nil {
		return nil, err
	}

	teamMembers, err := s.repo.GetTeamMembers(ctx, s.db, team.ID)
	if err != nil {
		return nil, err
	}

	return &TeamFull{
		Team:    team.Team,
		Members: teamMembers,
	}, nil
}

func (s *teamService) CreateTeam(ctx context.Context, teamToCreate *TeamFull) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()
	access, err := s.accessService.CanCreateTeam(ctx, teamToCreate.SchoolID)
	if err != nil {
		return ErrUnauthorized
	}
	if !access {
		return ErrUnauthorized
	}
	//get event associated to check constraints
	event, err := s.eventsRepo.GetEventByID(ctx, tx, teamToCreate.EventID)
	if err != nil {
		return ErrInternal
	}
	//check if team members constraints are satisfied
	if event.MaxTeamSize < int64(len(teamToCreate.Members)) {
		return ErrTeamSizeExceedsLimit
	}
	if event.MinTeamSize > int64(len(teamToCreate.Members)) {
		return ErrTeamSizeBelowMinimum
	}

	//fetch teams for this event from this school to see if MaxTeamsPerSchool is exceeded
	teamsFilter := TeamFilter{
		EventID:  &teamToCreate.EventID,
		SchoolID: &teamToCreate.SchoolID,
	}
	teamsInDB, err := s.repo.GetAllTeams(ctx, tx, teamsFilter)
	if event.MaxTeamsPerSchool <= int64(len(teamsInDB)) {
		return ErrTeamCountExceedsLimit
	}
	//create Team
	err = s.repo.CreateTeam(ctx, tx, &teamToCreate.Team)
	if err != nil {
		return ErrInternal
	}
	//create Team Members
	err = s.syncTeamMembers(ctx, tx, teamToCreate)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *teamService) UpdateTeam(ctx context.Context, teamToUpdate *TeamFull) error {
	access, err := s.accessService.CanModifyTeam(ctx, teamToUpdate.ID)
	if err != nil {
		return ErrUnauthorized
	}
	if !access {
		return ErrUnauthorized
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return ErrInternal
	}
	defer tx.Rollback()

	//get event associated to check constraints
	event, err := s.eventsRepo.GetEventByID(ctx, tx, teamToUpdate.EventID)
	if err != nil {
		return ErrInternal
	}
	//check if team members constraints are satisfied
	if event.MaxTeamSize < int64(len(teamToUpdate.Members)) {
		return ErrTeamSizeExceedsLimit
	}
	if event.MinTeamSize > int64(len(teamToUpdate.Members)) {
		return ErrTeamSizeBelowMinimum
	}

	//fetch teams for this event from this school to see if MaxTeamsPerSchool is exceeded
	teamsFilter := TeamFilter{
		EventID:  &teamToUpdate.EventID,
		SchoolID: &teamToUpdate.SchoolID,
	}
	teamsInDB, err := s.repo.GetAllTeams(ctx, tx, teamsFilter)
	//here we are only checking < because this event is also counted in teams
	if event.MaxTeamsPerSchool < int64(len(teamsInDB)) {
		return ErrTeamCountExceedsLimit
	}
	//cant update Team as cant chnge school, event or chest number

	//update Team Members
	err = s.syncTeamMembers(ctx, tx, teamToUpdate)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *teamService) DeleteTeam(ctx context.Context, teamID int64) error {
	access, err := s.accessService.CanCreateTeam(ctx, teamID)
	if err != nil {
		return ErrUnauthorized
	}
	if !access {
		return ErrUnauthorized
	}
	return s.repo.DeleteTeam(ctx, s.db, teamID)
}

// helper functions
func (s *teamService) syncTeamMembers(ctx context.Context, tx *sqlx.Tx, team *TeamFull) error {
	teamID := team.ID
	existing := []TeamMember{}
	var err error
	if teamID != 0 {
		existing, err = s.repo.GetTeamMembers(ctx, tx, team.ID)
		if err != nil {
			return ErrInternal
		}
	}
	requested := team.Members
	existingMap := make(map[int64]bool)
	for _, j := range existing {
		existingMap[j.StudentID] = true
	}

	requestedMap := make(map[int64]bool)
	for _, j := range requested {
		requestedMap[j.StudentID] = true
	}

	// delete removed members
	for _, j := range existing {
		if !requestedMap[j.StudentID] {
			if err := s.repo.DeleteTeamMember(ctx, tx, teamID, j.StudentID); err != nil {
				return err
			}
		}
	}
	// insert new members
	for _, j := range requested {
		if !existingMap[j.StudentID] {
			// Check if the student exists
			student, err := s.schoolRepo.GetStudentByID(ctx, tx, j.StudentID)
			if err != nil {
				return ErrInternal
			}
			event, err := s.eventsRepo.GetEventByID(ctx, tx, team.EventID)
			if err != nil {
				return ErrInternal
			}
			if event.Category != student.Category {
				return ErrCategoryMismatch
			}
			if team.Team.SchoolID != student.SchoolID {
				return ErrSchoolMismatch
			}

			if err := s.repo.CreateTeamMember(ctx, tx, &TeamMember{
				TeamID:    team.ID,
				StudentID: j.StudentID,
			}); err != nil {
				return err
			}
		}
	}
	createdTeamMembers, err := s.repo.GetTeamMembers(ctx, tx, team.ID)
	if err != nil {
		return ErrInternal
	}
	team.Members = createdTeamMembers
	return nil
}
