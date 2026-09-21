package entry

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/access"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/event"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
)

type entryService struct {
	db            *sqlx.DB
	accessService access.AccessService
	repo          EntryWriteRepository
	eventsRepo    event.EventReadRepository
	schoolRepo    school.SchoolReadRepository
}

type EntryService interface {
	GetEntries(ctx context.Context, filter EntryFilter) ([]EntryFull, error)
	GetEntryByID(ctx context.Context, entryID int64) (*EntryFull, error)
	CreateEntry(ctx context.Context, entryToCreate *EntryFull) error
	UpdateEntry(ctx context.Context, entryToUpdate *EntryFull) error
	DeleteEntry(ctx context.Context, entryID int64) error
}

func NewEntryService(
	db *sqlx.DB,
	accessService access.AccessService,
	repo EntryWriteRepository,
	eventsRepo event.EventReadRepository,
	schoolRepo school.SchoolReadRepository,
) EntryService {
	return &entryService{
		db:            db,
		accessService: accessService,
		repo:          repo,
		eventsRepo:    eventsRepo,
		schoolRepo:    schoolRepo}
}

func (s *entryService) GetEntries(ctx context.Context, filter EntryFilter) ([]EntryFull, error) {
	entriesFull := []EntryFull{}
	entries, err := s.repo.GetAllEntries(ctx, s.db, filter)
	if err != nil || entries == nil || len(entries) == 0 {
		return []EntryFull{}, err
	}
	for _, entry := range entries {
		entryMembers, err := s.repo.GetEntryMembers(ctx, s.db, entry.ID)
		if err != nil {
			return []EntryFull{}, err
		}
		entriesFull = append(entriesFull, EntryFull{
			Entry:   entry.Entry,
			Members: entryMembers,
		})
	}
	return entriesFull, nil
}

func (s *entryService) GetEntryByID(ctx context.Context, entryID int64) (*EntryFull, error) {
	entry, err := s.repo.GetEntryByID(ctx, s.db, entryID)
	if err != nil || entry == nil {
		return nil, err
	}

	entryMembers, err := s.repo.GetEntryMembers(ctx, s.db, entry.ID)
	if err != nil {
		return nil, err
	}

	return &EntryFull{
		Entry:   entry.Entry,
		Members: entryMembers,
	}, nil
}

func (s *entryService) CreateEntry(ctx context.Context, entryToCreate *EntryFull) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()
	access, err := s.accessService.CanCreateEntry(ctx, entryToCreate.SchoolID)
	if err != nil {
		return ErrUnauthorized
	}
	if !access {
		return ErrUnauthorized
	}
	//get event associated to check constraints
	event, err := s.eventsRepo.GetEventByID(ctx, tx, entryToCreate.EventID)
	if err != nil {
		return ErrInternal
	}
	//check if entry members constraints are satisfied
	if event.MaxMembers < int64(len(entryToCreate.Members)) {
		return ErrMemberCountExceedsLimit
	}
	if event.MinMembers > int64(len(entryToCreate.Members)) {
		return ErrMemberCountBelowMinimum
	}

	//fetch entries for this event from this school to see if MaxEntriesPerSchool is exceeded
	entriesFilter := EntryFilter{
		EventID:  &entryToCreate.EventID,
		SchoolID: &entryToCreate.SchoolID,
	}
	entriesInDB, err := s.repo.GetAllEntries(ctx, tx, entriesFilter)
	if event.MaxEntriesPerSchool <= int64(len(entriesInDB)) {
		return ErrEntryCountExceedsLimit
	}
	//create Entry
	err = s.repo.CreateEntry(ctx, tx, &entryToCreate.Entry)
	if err != nil {
		return ErrInternal
	}
	//create Entry Members
	err = s.syncEntryMembers(ctx, tx, entryToCreate)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *entryService) UpdateEntry(ctx context.Context, entryToUpdate *EntryFull) error {
	access, err := s.accessService.CanModifyEntry(ctx, entryToUpdate.ID)
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
	event, err := s.eventsRepo.GetEventByID(ctx, tx, entryToUpdate.EventID)
	if err != nil {
		return ErrInternal
	}
	//check if entry members constraints are satisfied
	if event.MaxMembers < int64(len(entryToUpdate.Members)) {
		return ErrMemberCountExceedsLimit
	}
	if event.MinMembers > int64(len(entryToUpdate.Members)) {
		return ErrMemberCountBelowMinimum
	}

	//fetch entries for this event from this school to see if MaxEntriesPerSchool is exceeded
	entriesFilter := EntryFilter{
		EventID:  &entryToUpdate.EventID,
		SchoolID: &entryToUpdate.SchoolID,
	}
	entriesInDB, err := s.repo.GetAllEntries(ctx, tx, entriesFilter)
	//here we are only checking < because this entry is also counted in entries
	if event.MaxEntriesPerSchool < int64(len(entriesInDB)) {
		return ErrEntryCountExceedsLimit
	}
	//cant update Entry as cant chnge school, event or chest number

	//update Entry Members
	err = s.syncEntryMembers(ctx, tx, entryToUpdate)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *entryService) DeleteEntry(ctx context.Context, entryID int64) error {
	access, err := s.accessService.CanModifyEntry(ctx, entryID)
	if err != nil {
		return ErrUnauthorized
	}
	if !access {
		return ErrUnauthorized
	}
	return s.repo.DeleteEntry(ctx, s.db, entryID)
}

// helper functions
func (s *entryService) syncEntryMembers(ctx context.Context, tx *sqlx.Tx, entry *EntryFull) error {
	entryID := entry.ID
	existing := []EntryMember{}
	var err error
	if entryID != 0 {
		existing, err = s.repo.GetEntryMembers(ctx, tx, entry.ID)
		if err != nil {
			return ErrInternal
		}
	}
	requested := entry.Members
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
			if err := s.repo.DeleteEntryMember(ctx, tx, entryID, j.StudentID); err != nil {
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
			event, err := s.eventsRepo.GetEventByID(ctx, tx, entry.EventID)
			if err != nil {
				return ErrInternal
			}
			if event.Category != student.Category {
				return ErrCategoryMismatch
			}
			if entry.Entry.SchoolID != student.SchoolID {
				return ErrSchoolMismatch
			}

			if err := s.repo.CreateEntryMember(ctx, tx, &EntryMember{
				EntryID:   entry.ID,
				StudentID: j.StudentID,
			}); err != nil {
				return err
			}
		}
	}
	createdEntryMembers, err := s.repo.GetEntryMembers(ctx, tx, entry.ID)
	if err != nil {
		return ErrInternal
	}
	entry.Members = createdEntryMembers
	return nil
}
