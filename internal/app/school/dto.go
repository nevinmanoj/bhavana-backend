package school

import (
	"time"

	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
)

type CreateSchoolRequest struct {
	Name         string `json:"name" validate:"required"`
	Address      string `json:"address"  validate:"required"`
	ContactName  string `json:"contact_name"  validate:"required"`
	ContactEmail string `json:"contact_email" validate:"required,email"`
	ContactPhone string `json:"contact_phone" validate:"required"`
}

type UpdateSchoolRequest struct {
	ID int64 `json:"id" validate:"required"`
	CreateSchoolRequest
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
	ID           int64     `json:"id"`
	SchoolAdmin  *int64    `json:"school_admin,omitempty"`
	Name         string    `json:"name"`
	Address      string    `json:"address"`
	ContactName  string    `json:"contact_name"`
	ContactEmail string    `json:"contact_email"`
	ContactPhone string    `json:"contact_phone"`
	CreatedAt    time.Time `json:"created_at"`
}
type StudentResponse struct {
	ID        int64         `json:"id"`
	SchoolID  int64         `json:"school_id"`
	Name      string        `json:"name"`
	Age       int           `json:"age"`
	Category  core.Category `json:"category"`
	CreatedAt time.Time     `json:"created_at"`
}

func ToSchoolResponse(s *school.School) SchoolResponse {
	return SchoolResponse{
		ID:           s.ID,
		Name:         s.Name,
		Address:      s.Address,
		ContactName:  s.ContactName,
		ContactEmail: s.ContactEmail,
		ContactPhone: s.ContactPhone,
		CreatedAt:    s.CreatedAt,
		SchoolAdmin:  s.SchoolAdmin,
	}
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
