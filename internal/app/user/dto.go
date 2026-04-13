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

type LoginUserResponse struct {
	UserResponse
	Token string `json:"token"`
}

type UserResponse struct {
	ID    int64         `json:"id"`
	Email string        `json:"email"`
	Name  string        `json:"name"`
	Role  rbac.UserRole `json:"role"`
}

func ToUserResponse(u *user.User) UserResponse {
	return UserResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
		Role:  u.Role,
	}
}
func ToLoginUserResponse(u *user.User, token string) LoginUserResponse {
	return LoginUserResponse{
		UserResponse: UserResponse{
			Email: u.Email,
			Name:  u.Name,
			Role:  u.Role,
		},
		Token: token,
	}
}
