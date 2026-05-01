package user

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	. "github.com/nevinmanoj/bhavana-backend/api"
	errmap "github.com/nevinmanoj/bhavana-backend/internal/app/errmap"
	user "github.com/nevinmanoj/bhavana-backend/internal/domain/user"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

type UserHandler struct {
	service   user.UserService
	validator *validator.Validate
}

func NewUserHandler(s user.UserService, v *validator.Validate) *UserHandler {
	return &UserHandler{service: s, validator: v}
}
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	var resp any
	q := r.URL.Query()

	filter, errresp := parseUserFilter(q)
	if errresp != nil {
		json.NewEncoder(w).Encode(errresp)
		return
	}
	users, err := h.service.GetAllUsers(ctx, filter)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
	} else {

		userResponses := make([]UserResponse, len(users))
		for i, u := range users {
			userResponses[i] = ToUserResponse(&u)
		}
		resp = GetAllResponsePage[UserResponse]{
			StatusCode: 200,
			Message:    "Users fetched successfully",
			Data:       userResponses,
		}
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userIdStr := chi.URLParam(r, "userId")
	w.Header().Set("Content-Type", "application/json")
	var resp any
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
		json.NewEncoder(w).Encode(resp)
		return
	}
	result, err := h.service.GetUserByID(ctx, userId)
	if err != nil {
		resp = errmap.GetDomainErrorResponse(err)
	} else {
		userResponse := ToUserResponse(result)
		resp = GetResponsePage[UserResponse]{
			StatusCode: 200,
			Message:    "User fetched successfully",
			Data:       userResponse,
		}
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req CreateUserRequest
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
	var email string = req.Email
	var password string = req.Password
	var name string = req.Name
	var role rbac.UserRole = req.Role
	var token = r.Header.Get("Authorization")
	userToCreate := &user.User{
		Email: email,
		Name:  name,
		Role:  role,
	}
	err := h.service.CreateUser(ctx, userToCreate, password, token)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    err.Error(),
		})
		return
	}
	userResponse := ToUserResponse(userToCreate)
	json.NewEncoder(w).Encode(PostResponsePage[UserResponse]{
		Message:    "User created successfully",
		Data:       userResponse,
		StatusCode: http.StatusCreated,
	})
}
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req LoginUserRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	w.Header().Set("Content-Type", "application/json")
	if err := dec.Decode(&req); err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid JSON body1",
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

	var email string = req.Email
	var password string = req.Password

	token, user, err := h.service.LoginUser(ctx, email, password)
	if err != nil {
		resp := errmap.GetDomainErrorResponse(err)
		json.NewEncoder(w).Encode(resp)
		return
	}
	logingResponse := ToLoginUserResponse(user, token)

	json.NewEncoder(w).Encode(PostResponsePage[LoginUserResponse]{
		Message:    "User logged in successfully",
		Data:       logingResponse,
		StatusCode: http.StatusCreated,
	})
}
