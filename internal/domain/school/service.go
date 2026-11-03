package school

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/auth"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/access"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/user"
	"github.com/nevinmanoj/bhavana-backend/internal/mail"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
	"github.com/nevinmanoj/bhavana-backend/internal/util"
)

// LoginEmail reports whether accepting a school emailed its admin a password.
type LoginEmail string

const (
	LoginEmailSent      LoginEmail = "sent"
	LoginEmailNotNeeded LoginEmail = "not_needed"
)

const backgroundMailTimeout = 30 * time.Second

type SchoolService interface {
	GetSchoolByID(ctx context.Context, id int64) (*School, error)
	GetAllSchools(ctx context.Context) ([]School, error)
	CreateSchool(ctx context.Context, school *School) error
	// RegisterSchool is the public sign-up: it creates the school admin user
	// (with no password yet) and the school (in_review) in one transaction, if
	// registration is open.
	RegisterSchool(ctx context.Context, school *School, admin *user.User) error
	// UpdateSchoolStatus emails the school admin a generated password when an
	// in-review school is accepted and the admin has none yet.
	UpdateSchoolStatus(ctx context.Context, id int64, status core.SchoolStatus) (LoginEmail, error)
	// SendLoginDetails emails the accepted school's admin a new password,
	// first changing their login email if email differs.
	SendLoginDetails(ctx context.Context, schoolID int64, email string) error
	GetRegistrationOpen(ctx context.Context) (bool, error)
	SetRegistrationOpen(ctx context.Context, open bool) error
	UpdateSchool(ctx context.Context, school *School) error
	DeleteSchool(ctx context.Context, id int64) error

	CreateStudent(ctx context.Context, student *Student) error
	UpdateStudent(ctx context.Context, student *Student) error
	DeleteStudent(ctx context.Context, id int64) error
	// GetAllStudents returns the matching page (per filter.Limit/Offset, or
	// every row if unset) plus the total count for that same filter.
	GetAllStudents(ctx context.Context, filter StudentFilter) ([]Student, int64, error)
}

type schoolService struct {
	db            *sqlx.DB
	accessService access.AccessService
	repo          SchoolWriteRepository
	userRepo      user.UserWriteRepository
	mailer        mail.Mailer
	appURL        string
}

func NewSchoolService(db *sqlx.DB, accessService access.AccessService, repo SchoolWriteRepository, userRepo user.UserWriteRepository, mailer mail.Mailer, appURL string) SchoolService {
	return &schoolService{db: db, accessService: accessService, repo: repo, userRepo: userRepo, mailer: mailer, appURL: appURL}
}

// schools
func (s *schoolService) GetAllSchools(ctx context.Context) ([]School, error) {
	return s.repo.GetAllSchools(ctx, s.db)
}
func (s *schoolService) GetSchoolByID(ctx context.Context, id int64) (*School, error) {
	return s.repo.GetSchoolByID(ctx, s.db, id)
}
func (s *schoolService) CreateSchool(ctx context.Context, school *School) error {
	// admin-created schools skip review
	school.Status = core.SchoolStatusAccepted
	return s.repo.CreateSchool(ctx, s.db, school)
}
func (s *schoolService) UpdateSchool(ctx context.Context, school *School) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return ErrInternal
	}
	defer tx.Rollback()

	locked, err := s.repo.LockSchool(ctx, tx, school.ID)
	if err != nil {
		return err
	}
	if contactEmailLocked(locked, school) {
		return ErrContactEmailLocked
	}
	if err := s.repo.UpdateSchool(ctx, tx, school); err != nil {
		return err
	}
	if syncsAdminEmail(locked, school) {
		if err := s.userRepo.UpdateEmail(ctx, tx, *locked.SchoolAdmin, school.ContactEmail); err != nil {
			if err == user.ErrAlreadyExists {
				return ErrEmailTaken
			}
			return ErrInternal
		}
	}
	if err := tx.Commit(); err != nil {
		return ErrInternal
	}
	return nil
}

// contactEmailLocked: once a school with a school admin is accepted, its email
// only changes through SendLoginDetails, which moves the sign-in email with it
// and sends a password to the new address. A school without an admin has no
// sign-in to keep in step, so its contact email stays freely editable.
func contactEmailLocked(locked *LockedSchool, updated *School) bool {
	return locked.Status == core.SchoolStatusAccepted &&
		locked.SchoolAdmin != nil &&
		updated.ContactEmail != locked.ContactEmail
}

// syncsAdminEmail: while a registration is in review, its contact email is
// also the admin's login, so a reviewer correcting a typo before accepting
// sends the login details to the corrected address. Not when the edit also
// reassigns the school admin, and not after acceptance, when the two may
// legitimately differ.
func syncsAdminEmail(locked *LockedSchool, updated *School) bool {
	return locked.Status == core.SchoolStatusInReview &&
		locked.SchoolAdmin != nil &&
		updated.SchoolAdmin != nil && *updated.SchoolAdmin == *locked.SchoolAdmin &&
		updated.ContactEmail != locked.ContactEmail
}
func (s *schoolService) DeleteSchool(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return ErrInternal
	}
	defer tx.Rollback()

	locked, err := s.repo.LockSchool(ctx, tx, id)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteSchool(ctx, tx, id); err != nil {
		return err
	}
	// rejecting an in-review registration must not leave its login behind
	if locked.Status == core.SchoolStatusInReview && locked.SchoolAdmin != nil {
		if err := s.userRepo.DeleteUser(ctx, tx, *locked.SchoolAdmin); err != nil {
			return ErrInternal
		}
	}
	if err := tx.Commit(); err != nil {
		return ErrInternal
	}
	return nil
}

func (s *schoolService) RegisterSchool(ctx context.Context, school *School, admin *user.User) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return ErrInternal
	}
	defer tx.Rollback()

	open, err := s.repo.IsRegistrationOpen(ctx, tx)
	if err != nil {
		return ErrInternal
	}
	if !open {
		return ErrRegistrationClosed
	}

	admin.Role = rbac.UserRoleSchoolAdmin
	// no password yet: accepting the school emails a generated one
	if err := s.userRepo.CreateUser(ctx, tx, "", admin); err != nil {
		if err == user.ErrAlreadyExists {
			return ErrEmailTaken
		}
		return err
	}

	school.SchoolAdmin = &admin.ID
	school.Status = core.SchoolStatusInReview
	if err := s.repo.CreateSchool(ctx, tx, school); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return ErrInternal
	}
	// best effort: the registration is saved whether or not this arrives
	s.sendInBackground(mail.RegistrationReceivedEmail(admin.Email, admin.Name, school.Name))
	return nil
}

func (s *schoolService) UpdateSchoolStatus(ctx context.Context, id int64, status core.SchoolStatus) (LoginEmail, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", ErrInternal
	}
	defer tx.Rollback()

	locked, err := s.repo.LockSchool(ctx, tx, id)
	if err != nil {
		return "", err
	}
	if err := s.repo.UpdateSchoolStatus(ctx, tx, id, status); err != nil {
		return "", err
	}

	result := LoginEmailNotNeeded
	// Only the first acceptance sends a password: an admin who already has
	// one (accepted before, then moved back to review) keeps it.
	if locked.Status == core.SchoolStatusInReview && status == core.SchoolStatusAccepted && locked.SchoolAdmin != nil {
		admin, err := s.userRepo.GetUserByID(ctx, tx, *locked.SchoolAdmin)
		if err != nil {
			return "", ErrInternal
		}
		if admin.PasswordHash == "" {
			// a failed send rolls the acceptance back too
			if err := s.issueLogin(ctx, tx, admin, locked.Name, mail.AcceptedLoginEmail); err != nil {
				return "", err
			}
			result = LoginEmailSent
		}
	}
	if err := tx.Commit(); err != nil {
		return "", ErrInternal
	}
	return result, nil
}

func (s *schoolService) SendLoginDetails(ctx context.Context, schoolID int64, email string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return ErrInternal
	}
	defer tx.Rollback()

	locked, err := s.repo.LockSchool(ctx, tx, schoolID)
	if err != nil {
		return err
	}
	if locked.SchoolAdmin == nil {
		return ErrNoSchoolAdmin
	}
	// Accept sends the first login details. Sending them now would give the
	// admin a password, and acceptance would then skip its email.
	if locked.Status == core.SchoolStatusInReview {
		return ErrSchoolInReview
	}
	admin, err := s.userRepo.GetUserByID(ctx, tx, *locked.SchoolAdmin)
	if err != nil {
		return ErrInternal
	}

	email = util.NormalizeEmail(email)
	if email != admin.Email {
		if err := s.userRepo.UpdateEmail(ctx, tx, admin.ID, email); err != nil {
			if err == user.ErrAlreadyExists {
				return ErrEmailTaken
			}
			return ErrInternal
		}
		if err := s.repo.UpdateContactEmail(ctx, tx, schoolID, email); err != nil {
			return err
		}
		admin.Email = email
	}

	if err := s.issueLogin(ctx, tx, admin, locked.Name, mail.NewPasswordEmail); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return ErrInternal
	}
	return nil
}

type loginEmailBuilder func(appURL, email, name, schoolName, password string) mail.Message

// issueLogin replaces the admin's password with a generated one and emails
// it, inside the caller's transaction. Sending before the commit means the
// database never keeps a password nobody received.
func (s *schoolService) issueLogin(ctx context.Context, tx sqlx.ExtContext, admin *user.User, schoolName string, build loginEmailBuilder) error {
	password, err := auth.GeneratePassword()
	if err != nil {
		return ErrInternal
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return ErrInternal
	}
	if err := s.userRepo.SetPasswordHash(ctx, tx, admin.ID, hash); err != nil {
		return ErrInternal
	}
	if err := s.userRepo.InvalidateResetTokens(ctx, tx, admin.ID); err != nil {
		return ErrInternal
	}
	if err := s.mailer.Send(ctx, build(s.appURL, admin.Email, admin.Name, schoolName, password)); err != nil {
		if errors.Is(err, mail.ErrRateLimited) {
			return ErrMailRateLimited
		}
		log.Printf("send login details to user %d: %v", admin.ID, err)
		return ErrMailFailed
	}
	return nil
}

// sendInBackground sends without holding up the request; failures are only logged.
func (s *schoolService) sendInBackground(msg mail.Message) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), backgroundMailTimeout)
		defer cancel()
		if err := s.mailer.Send(ctx, msg); err != nil {
			log.Printf("send %q: %v", msg.Subject, err)
		}
	}()
}
func (s *schoolService) GetRegistrationOpen(ctx context.Context) (bool, error) {
	return s.repo.IsRegistrationOpen(ctx, s.db)
}
func (s *schoolService) SetRegistrationOpen(ctx context.Context, open bool) error {
	return s.repo.SetRegistrationOpen(ctx, s.db, open)
}

// students
func (s *schoolService) GetAllStudents(ctx context.Context, filter StudentFilter) ([]Student, int64, error) {
	students, err := s.repo.GetAllStudents(ctx, s.db, filter)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountStudents(ctx, s.db, filter)
	if err != nil {
		return nil, 0, err
	}
	return students, total, nil
}
func (s *schoolService) UpdateStudent(ctx context.Context, student *Student) error {
	access, err := s.accessService.CanModifyStudent(ctx, student.ID)
	if err != nil {
		return err
	}
	if !access {
		return ErrUnauthorized
	}
	return s.repo.UpdateStudent(ctx, s.db, student)
}
func (s *schoolService) CreateStudent(ctx context.Context, student *Student) error {
	access, err := s.accessService.CanCreateStudent(ctx, student.SchoolID)
	if err != nil {
		return err
	}
	if !access {
		return ErrUnauthorized
	}
	exists, err := s.repo.SchoolExists(ctx, s.db, student.SchoolID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrSchoolNotFound
	}
	return s.repo.CreateStudent(ctx, s.db, student)
}
func (s *schoolService) DeleteStudent(ctx context.Context, id int64) error {
	access, err := s.accessService.CanModifyStudent(ctx, id)
	if err != nil {
		return err
	}
	if !access {
		return ErrUnauthorized
	}
	return s.repo.DeleteStudent(ctx, s.db, id)
}
