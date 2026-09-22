package entry

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/access"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/event"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
	"github.com/nevinmanoj/bhavana-backend/internal/middleware"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
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
	// CreateEntries registers a batch of entries for one event and school in a
	// single transaction. Entries are filled in with their assigned IDs.
	CreateEntries(ctx context.Context, eventID, schoolID int64, entriesToCreate []*EntryFull) error
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

// entryRegistrationAllowed reports whether a role may register entries while the
// event sits in status. School admins get the registration window only; admins
// can still add late entries right through to the end of scoring, which is why
// assign_chest_number keeps its MAX+1 fallback for entries created after the
// draw has already run.
func entryRegistrationAllowed(role rbac.UserRole, status core.EventStatus) bool {
	if role == rbac.UserRoleAdmin {
		switch status {
		case core.EventStatusRegistrationOpen,
			core.EventStatusRegistrationClosed,
			core.EventStatusPreparing,
			core.EventStatusOpen,
			core.EventStatusClosed:
			return true
		}
		return false
	}
	return status == core.EventStatusRegistrationOpen
}

// validateBatch applies the event's own rules to a batch before any of it is
// written. Keeping it pure and separate means the whole batch is rejected up
// front rather than half-inserted and rolled back.
func validateBatch(ev *event.Event, entriesToCreate []*EntryFull) error {
	if len(entriesToCreate) == 0 {
		return ErrEmptyBatch
	}
	seen := make(map[int64]bool)
	for _, e := range entriesToCreate {
		if ev.MaxMembers < int64(len(e.Members)) {
			return ErrMemberCountExceedsLimit
		}
		if ev.MinMembers > int64(len(e.Members)) {
			return ErrMemberCountBelowMinimum
		}
		// a student can only appear once across the whole event, so catch
		// duplicates inside the batch before the DB trigger aborts the tx
		for _, m := range e.Members {
			if seen[m.StudentID] {
				return ErrStudentAlreadyInEntry
			}
			seen[m.StudentID] = true
		}
	}
	return nil
}

func (s *entryService) CreateEntry(ctx context.Context, entryToCreate *EntryFull) error {
	return s.CreateEntries(ctx, entryToCreate.EventID, entryToCreate.SchoolID, []*EntryFull{entryToCreate})
}

func (s *entryService) CreateEntries(ctx context.Context, eventID, schoolID int64, entriesToCreate []*EntryFull) error {
	if len(entriesToCreate) == 0 {
		return ErrEmptyBatch
	}
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	//one access check covers the batch: every entry shares the same school
	access, err := s.accessService.CanCreateEntry(ctx, schoolID)
	if err != nil {
		return ErrUnauthorized
	}
	if !access {
		return ErrUnauthorized
	}

	//get event associated to check constraints
	ev, err := s.eventsRepo.GetEventByID(ctx, tx, eventID)
	if err != nil {
		return ErrInternal
	}

	//registration window, which differs by role
	role, _ := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	if !entryRegistrationAllowed(role, ev.Status) {
		return ErrRegistrationNotOpen
	}

	for _, e := range entriesToCreate {
		if e.EventID != eventID || e.SchoolID != schoolID {
			return ErrInvalidEntryUpdate
		}
	}
	if err := validateBatch(ev, entriesToCreate); err != nil {
		return err
	}

	//fetch entries for this event from this school to see if MaxEntriesPerSchool
	//is exceeded - the whole batch has to fit, not just one more entry
	entriesFilter := EntryFilter{
		EventID:  &eventID,
		SchoolID: &schoolID,
	}
	entriesInDB, err := s.repo.GetAllEntries(ctx, tx, entriesFilter)
	if err != nil {
		return ErrInternal
	}
	if int64(len(entriesInDB)+len(entriesToCreate)) > ev.MaxEntriesPerSchool {
		return ErrEntryCountExceedsLimit
	}

	for _, e := range entriesToCreate {
		//create Entry
		if err := s.repo.CreateEntry(ctx, tx, &e.Entry); err != nil {
			return err
		}
		//create Entry Members
		if err := s.syncEntryMembers(ctx, tx, e); err != nil {
			return err
		}
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

	//registration window, which differs by role - admins may still fix up an
	//entry right through scoring; school admins only while registration_open
	role, _ := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	if role != rbac.UserRoleAdmin && !entryRegistrationAllowed(role, event.Status) {
		return ErrRegistrationNotOpen
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

	//registration window, which differs by role - admins may remove a late or
	//mistaken entry any time; school admins only while registration_open
	role, _ := ctx.Value(middleware.ContextUserRole).(rbac.UserRole)
	if role != rbac.UserRoleAdmin {
		entryToDelete, err := s.repo.GetEntryByID(ctx, s.db, entryID)
		if err != nil {
			return err
		}
		ev, err := s.eventsRepo.GetEventByID(ctx, s.db, entryToDelete.EventID)
		if err != nil {
			return ErrInternal
		}
		if !entryRegistrationAllowed(role, ev.Status) {
			return ErrRegistrationNotOpen
		}
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
