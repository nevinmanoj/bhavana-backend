package school

import (
	"time"

	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
	"github.com/nevinmanoj/bhavana-backend/internal/util"
)

// SchoolFields is shared by create, update and register: every field is
// required. Headcounts are pointers because `required` rejects a plain 0.
type SchoolFields struct {
	Name            string `json:"name" validate:"required"`
	Address         string `json:"address"  validate:"required"`
	ContactName     string `json:"contact_name"  validate:"required"`
	ContactEmail    string `json:"contact_email" validate:"required,email"`
	ContactPhone    string `json:"contact_phone" validate:"required,mobile"`
	PrincipalName   string `json:"principal_name" validate:"required"`
	PrincipalPhone  string `json:"principal_phone" validate:"required,mobile"`
	ParticipantsMC  *int   `json:"participants_mc" validate:"required,gte=0"`
	ParticipantsPC  *int   `json:"participants_pc" validate:"required,gte=0"`
	ParticipantsHI  *int   `json:"participants_hi" validate:"required,gte=0"`
	NonParticipants *int   `json:"non_participants" validate:"required,gte=0"`
}

// normalize trims text and canonicalizes email / phones. Handlers call it
// before validating so the validators see (and the DB stores) the clean value.
func (f *SchoolFields) normalize() {
	f.Name = trim(f.Name)
	f.Address = trim(f.Address)
	f.ContactName = trim(f.ContactName)
	f.PrincipalName = trim(f.PrincipalName)
	f.ContactEmail = util.NormalizeEmail(f.ContactEmail)
	f.ContactPhone = util.NormalizeMobile(f.ContactPhone)
	f.PrincipalPhone = util.NormalizeMobile(f.PrincipalPhone)
}

// toSchool assumes the request has already passed validation (pointers non-nil).
func (f *SchoolFields) toSchool() school.School {
	return school.School{
		Name:            f.Name,
		Address:         f.Address,
		ContactName:     f.ContactName,
		ContactEmail:    f.ContactEmail,
		ContactPhone:    f.ContactPhone,
		PrincipalName:   f.PrincipalName,
		PrincipalPhone:  f.PrincipalPhone,
		ParticipantsMC:  *f.ParticipantsMC,
		ParticipantsPC:  *f.ParticipantsPC,
		ParticipantsHI:  *f.ParticipantsHI,
		NonParticipants: *f.NonParticipants,
	}
}

type CreateSchoolRequest struct {
	SchoolFields
}

type UpdateSchoolRequest struct {
	ID          int64  `json:"id" validate:"required"`
	SchoolAdmin *int64 `json:"school_admin"`
	SchoolFields
}

// RegisterSchoolRequest has no password: acceptance emails a generated one.
type RegisterSchoolRequest struct {
	SchoolFields
}

type UpdateSchoolStatusRequest struct {
	Status core.SchoolStatus `json:"status" validate:"required,school_status"`
}

type UpdateSchoolStatusResponse struct {
	SchoolResponse
	LoginEmail school.LoginEmail `json:"login_email"`
}

type SendLoginDetailsRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type SettingsRequest struct {
	SchoolRegistrationOpen  *bool `json:"school_registration_open" validate:"required"`
	StudentRegistrationOpen *bool `json:"student_registration_open" validate:"required"`
}

type SettingsResponse struct {
	SchoolRegistrationOpen  bool `json:"school_registration_open"`
	StudentRegistrationOpen bool `json:"student_registration_open"`
}

func ToSettingsResponse(s *school.Settings) SettingsResponse {
	return SettingsResponse{
		SchoolRegistrationOpen:  s.SchoolRegistrationOpen,
		StudentRegistrationOpen: s.StudentRegistrationOpen,
	}
}

type CreateStudentRequest struct {
	Name     string        `json:"name" validate:"required"`
	Age      int           `json:"age" validate:"required,gte=0"`
	Category core.Category `json:"category" validate:"required,category"`
}
type UpdateStudentRequest struct {
	ID   int64  `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type SchoolResponse struct {
	ID              int64             `json:"id"`
	SchoolAdmin     *int64            `json:"school_admin,omitempty"`
	Name            string            `json:"name"`
	Address         string            `json:"address"`
	ContactName     string            `json:"contact_name"`
	ContactEmail    string            `json:"contact_email"`
	ContactPhone    string            `json:"contact_phone"`
	Status          core.SchoolStatus `json:"status"`
	PrincipalName   string            `json:"principal_name"`
	PrincipalPhone  string            `json:"principal_phone"`
	ParticipantsMC  int               `json:"participants_mc"`
	ParticipantsPC  int               `json:"participants_pc"`
	ParticipantsHI  int               `json:"participants_hi"`
	NonParticipants int               `json:"non_participants"`
	CreatedAt       time.Time         `json:"created_at"`
}

type StudentResponse struct {
	ID        int64         `json:"id"`
	SchoolID  int64         `json:"school_id"`
	Name      string        `json:"name"`
	Age       int           `json:"age"`
	Category  core.Category `json:"category"`
	CreatedAt time.Time     `json:"created_at"`
}

func ToStudentResponse(st *school.Student) StudentResponse {
	return StudentResponse{
		ID:        st.ID,
		SchoolID:  st.SchoolID,
		Name:      st.Name,
		Age:       st.Age,
		Category:  st.Category,
		CreatedAt: st.CreatedAt,
	}
}
func ToSchoolResponse(s *school.School) SchoolResponse {
	return SchoolResponse{
		ID:              s.ID,
		Name:            s.Name,
		Address:         s.Address,
		ContactName:     s.ContactName,
		ContactEmail:    s.ContactEmail,
		ContactPhone:    s.ContactPhone,
		Status:          s.Status,
		PrincipalName:   s.PrincipalName,
		PrincipalPhone:  s.PrincipalPhone,
		ParticipantsMC:  s.ParticipantsMC,
		ParticipantsPC:  s.ParticipantsPC,
		ParticipantsHI:  s.ParticipantsHI,
		NonParticipants: s.NonParticipants,
		CreatedAt:       s.CreatedAt,
		SchoolAdmin:     s.SchoolAdmin,
	}
}
