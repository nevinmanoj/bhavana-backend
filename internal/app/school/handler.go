package school

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	. "github.com/nevinmanoj/bhavana-backend/api"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/school"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/user"
	"github.com/nevinmanoj/bhavana-backend/internal/util"
)

type SchoolHandler struct {
	service   school.SchoolService
	validator *validator.Validate
}

func NewSchoolHandler(s school.SchoolService, v *validator.Validate) *SchoolHandler {
	return &SchoolHandler{service: s, validator: v}
}

// school handlers
func (h *SchoolHandler) GetSchools(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	var resp any
	schools, err := h.service.GetAllSchools(ctx)
	if err != nil {
		resp = GetSchoolDomainErrorResponse(err)
	} else {
		schoolResponses := make([]SchoolResponse, len(schools))
		for i, s := range schools {
			schoolResponses[i] = ToSchoolResponse(&s)
		}
		resp = GetAllResponsePage[SchoolResponse]{
			StatusCode: 200,
			Message:    "Schools fetched successfully",
			Data:       schoolResponses,
		}
	}
	json.NewEncoder(w).Encode(resp)
}
func (h *SchoolHandler) GetSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	var resp any
	schoolIdStr := chi.URLParam(r, "schoolId")
	schoolId, err := strconv.ParseInt(schoolIdStr, 10, 64)
	if err != nil {
		resp = ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid school ID in URL parameter",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	school, err := h.service.GetSchoolByID(ctx, schoolId)
	if err != nil {
		resp = GetSchoolDomainErrorResponse(err)
	} else {
		resp = GetResponsePage[SchoolResponse]{
			StatusCode: 200,
			Message:    "School fetched successfully",
			Data:       ToSchoolResponse(school),
		}
	}
	json.NewEncoder(w).Encode(resp)
}
func (h *SchoolHandler) CreateSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateSchoolRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body",
		})
		return
	}

	req.normalize()
	if err := h.validator.Struct(req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	schoolToCreate := req.toSchool()
	err := h.service.CreateSchool(ctx, &schoolToCreate)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	schoolResponse := ToSchoolResponse(&schoolToCreate)
	json.NewEncoder(w).Encode(PostResponsePage[SchoolResponse]{
		Message:    "School created successfully",
		Data:       schoolResponse,
		StatusCode: http.StatusCreated,
	})
}
func (h *SchoolHandler) UpdateSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	schoolIdStr := chi.URLParam(r, "schoolId")
	var req UpdateSchoolRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body",
		})
		return
	}

	req.normalize()
	if err := h.validator.Struct(req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}
	schoolId, err := util.ParseStrToInt64(schoolIdStr)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid school ID in URL parameter",
		})
		return
	}
	if req.ID != *schoolId {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "School ID in request body does not match URL parameter",
		})
		return
	}

	schoolToUpdate := req.toSchool()
	schoolToUpdate.ID = req.ID
	schoolToUpdate.SchoolAdmin = req.SchoolAdmin

	err = h.service.UpdateSchool(ctx, &schoolToUpdate)
	if err != nil {
		json.NewEncoder(w).Encode(GetSchoolDomainErrorResponse(err))
		return
	}
	// re-read so the response carries status / created_at, which update does not set
	updated, err := h.service.GetSchoolByID(ctx, schoolToUpdate.ID)
	if err != nil {
		json.NewEncoder(w).Encode(GetSchoolDomainErrorResponse(err))
		return
	}
	schoolResponse := ToSchoolResponse(updated)
	json.NewEncoder(w).Encode(PostResponsePage[SchoolResponse]{
		Message:    "School updated successfully",
		Data:       schoolResponse,
		StatusCode: http.StatusOK,
	})
}
func (h *SchoolHandler) DeleteSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	schoolIdStr := chi.URLParam(r, "schoolId")
	w.Header().Set("Content-Type", "application/json")
	var resp any
	schoolId, err := strconv.ParseInt(schoolIdStr, 10, 64)
	if err != nil {
		resp = ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid school ID in URL parameter",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	err = h.service.DeleteSchool(ctx, schoolId)
	if err != nil {
		resp = GetSchoolDomainErrorResponse(err)
	} else {
		resp = DeleteResponsePage{
			StatusCode: http.StatusNoContent,
			Message:    "School deleted successfully",
		}
	}

	json.NewEncoder(w).Encode(resp)
}

// student handlers
func (h *SchoolHandler) GetStudents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	var resp any
	filter, errresp := parseStudentFilter(r.URL.Query())
	if errresp != nil {
		resp = errresp
		json.NewEncoder(w).Encode(resp)
		return
	}
	students, total, err := h.service.GetAllStudents(ctx, filter)
	if err != nil {
		resp = GetSchoolDomainErrorResponse(err)
	} else {
		studentResponses := make([]StudentResponse, len(students))
		for i, s := range students {
			studentResponses[i] = ToStudentResponse(&s)
		}
		resp = GetAllResponsePage[StudentResponse]{
			StatusCode:   200,
			Message:      "Students fetched successfully",
			Data:         studentResponses,
			TotalRecords: total,
			Limit:        filter.Limit,
			Offset:       filter.Offset,
		}
	}
	json.NewEncoder(w).Encode(resp)
}
func (h *SchoolHandler) GetStudentsBySchoolID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	var resp any
	schoolIdStr := chi.URLParam(r, "schoolId")
	schoolId, err := strconv.ParseInt(schoolIdStr, 10, 64)
	if err != nil {
		resp = ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid school ID in URL parameter",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	filter, errresp := parseStudentFilter(r.URL.Query())
	filter.SchoolID = &schoolId
	if errresp != nil {
		resp = errresp
		json.NewEncoder(w).Encode(resp)
		return
	}
	students, total, err := h.service.GetAllStudents(ctx, filter)
	if err != nil {
		resp = GetSchoolDomainErrorResponse(err)
	} else {
		studentResponses := make([]StudentResponse, len(students))
		for i, s := range students {
			studentResponses[i] = ToStudentResponse(&s)
		}
		resp = GetAllResponsePage[StudentResponse]{
			StatusCode:   200,
			Message:      "Students fetched successfully",
			Data:         studentResponses,
			TotalRecords: total,
			Limit:        filter.Limit,
			Offset:       filter.Offset,
		}
	}
	json.NewEncoder(w).Encode(resp)
}
func (h *SchoolHandler) CreateStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateStudentRequest
	var resp any
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body",
		})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}
	schoolIdStr := chi.URLParam(r, "schoolId")
	schoolId, err := strconv.ParseInt(schoolIdStr, 10, 64)
	if err != nil {
		resp = ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid school ID in URL parameter",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	studentToCreate := school.Student{
		Name:     req.Name,
		Age:      req.Age,
		SchoolID: schoolId,
		Category: req.Category,
	}
	err = h.service.CreateStudent(ctx, &studentToCreate)
	if err != nil {
		json.NewEncoder(w).Encode(GetSchoolDomainErrorResponse(err))
		return
	}
	studentResponse := ToStudentResponse(&studentToCreate)
	json.NewEncoder(w).Encode(PostResponsePage[StudentResponse]{
		Message:    "Student created successfully",
		Data:       studentResponse,
		StatusCode: http.StatusCreated,
	})
}
func (h *SchoolHandler) UpdateStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	studentIdStr := chi.URLParam(r, "studentId")
	studentId, err := util.ParseStrToInt64(studentIdStr)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid student ID in URL parameter",
		})
		return
	}
	var req UpdateStudentRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body",
		})
		return
	}

	if err := h.validator.Struct(req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}
	schoolIdStr := chi.URLParam(r, "schoolId")
	schoolId, err := util.ParseStrToInt64(schoolIdStr)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid school ID in URL parameter",
		})
		return
	}
	if req.ID != *studentId {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Student ID in request body does not match URL parameter",
		})
		return
	}

	studentToUpdate := school.Student{
		ID:       req.ID,
		Name:     req.Name,
		SchoolID: *schoolId,
	}

	err = h.service.UpdateStudent(ctx, &studentToUpdate)
	if err != nil {
		json.NewEncoder(w).Encode(GetSchoolDomainErrorResponse(err))
		return
	}
	studentResponse := ToStudentResponse(&studentToUpdate)
	json.NewEncoder(w).Encode(PostResponsePage[StudentResponse]{
		Message:    "Student updated successfully",
		Data:       studentResponse,
		StatusCode: http.StatusOK,
	})
}
func (h *SchoolHandler) DeleteStudent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	studentIdStr := chi.URLParam(r, "studentId")
	w.Header().Set("Content-Type", "application/json")
	var resp any
	studentId, err := strconv.ParseInt(studentIdStr, 10, 64)
	if err != nil {
		resp = ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid student ID in URL parameter",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}
	err = h.service.DeleteStudent(ctx, studentId)
	if err != nil {
		resp = GetSchoolDomainErrorResponse(err)
	} else {
		resp = DeleteResponsePage{
			StatusCode: http.StatusNoContent,
			Message:    "Student deleted successfully",
		}
	}

	json.NewEncoder(w).Encode(resp)
}

func trim(s string) string { return strings.TrimSpace(s) }

// RegisterSchool is the public sign-up: it creates the school admin login and
// the school (in review) together. No token is returned, since the account
// can't sign in until an admin accepts the school.
func (h *SchoolHandler) RegisterSchool(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req RegisterSchoolRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body",
		})
		return
	}
	req.normalize()
	if err := h.validator.Struct(req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}

	schoolToCreate := req.toSchool()
	// the contact person is the school admin
	admin := &user.User{Name: req.ContactName, Email: req.ContactEmail}
	if err := h.service.RegisterSchool(ctx, &schoolToCreate, admin, req.Password); err != nil {
		json.NewEncoder(w).Encode(GetSchoolDomainErrorResponse(err))
		return
	}
	json.NewEncoder(w).Encode(PostResponsePage[SchoolResponse]{
		Message:    "School registered successfully. An admin will review your details.",
		Data:       ToSchoolResponse(&schoolToCreate),
		StatusCode: http.StatusCreated,
	})
}

func (h *SchoolHandler) UpdateSchoolStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	schoolId, err := util.ParseStrToInt64(chi.URLParam(r, "schoolId"))
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid school ID in URL parameter",
		})
		return
	}
	var req UpdateSchoolStatusRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body",
		})
		return
	}
	if err := h.validator.Struct(req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}
	if err := h.service.UpdateSchoolStatus(ctx, *schoolId, req.Status); err != nil {
		json.NewEncoder(w).Encode(GetSchoolDomainErrorResponse(err))
		return
	}
	updated, err := h.service.GetSchoolByID(ctx, *schoolId)
	if err != nil {
		json.NewEncoder(w).Encode(GetSchoolDomainErrorResponse(err))
		return
	}
	json.NewEncoder(w).Encode(PostResponsePage[SchoolResponse]{
		Message:    "School status updated successfully",
		Data:       ToSchoolResponse(updated),
		StatusCode: http.StatusOK,
	})
}

func (h *SchoolHandler) GetRegistrationOpen(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	open, err := h.service.GetRegistrationOpen(r.Context())
	if err != nil {
		json.NewEncoder(w).Encode(GetSchoolDomainErrorResponse(err))
		return
	}
	json.NewEncoder(w).Encode(GetResponsePage[RegistrationOpenResponse]{
		StatusCode: 200,
		Message:    "Registration status fetched successfully",
		Data:       RegistrationOpenResponse{Open: open},
	})
}

func (h *SchoolHandler) SetRegistrationOpen(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req RegistrationOpenRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body",
		})
		return
	}
	if err := h.validator.Struct(req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		})
		return
	}
	if err := h.service.SetRegistrationOpen(r.Context(), *req.Open); err != nil {
		json.NewEncoder(w).Encode(GetSchoolDomainErrorResponse(err))
		return
	}
	json.NewEncoder(w).Encode(GetResponsePage[RegistrationOpenResponse]{
		StatusCode: 200,
		Message:    "Registration status updated successfully",
		Data:       RegistrationOpenResponse{Open: *req.Open},
	})
}
