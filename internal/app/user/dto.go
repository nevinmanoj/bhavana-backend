package user

import (
	user "github.com/nevinmanoj/bhavana-backend/internal/domain/user"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

type LoginUserRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type CreateUserRequest struct {
	Email    string        `json:"email" validate:"required,email"`
	Password string        `json:"password" validate:"required,min=6"`
	Name     string        `json:"name" validate:"required"`
	Role     rbac.UserRole `json:"role" validate:"required,user_role"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

type LoginUserResponse struct {
	UserResponse
	Token string `json:"token"`
}

type UserResponse struct {
	ID    int64         `json:"id"`
	Email string        `json:"email"`
	Name  string        `json:"name"`
	Role  rbac.UserRole `json:"role"`
	// false for a registered school admin still waiting for acceptance;
	// the hash itself is never serialized
	HasPassword bool `json:"has_password"`
}

func ToUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		Name:        u.Name,
		Role:        u.Role,
		HasPassword: u.PasswordHash != "",
	}
}
func ToLoginUserResponse(u *user.User, token string) LoginUserResponse {
	return LoginUserResponse{
		UserResponse: ToUserResponse(u),
		Token:        token,
	}
}
