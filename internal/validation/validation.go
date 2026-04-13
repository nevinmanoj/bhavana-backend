package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

func NewValidator() *validator.Validate {
	v := validator.New()

	// register event status validation
	v.RegisterValidation("event_status", func(fl validator.FieldLevel) bool {
		status, ok := fl.Field().Interface().(core.EventStatus)
		if !ok {
			return false
		}
		return status.IsValid()
	})

	//register user role validation
	v.RegisterValidation("user_role", func(fl validator.FieldLevel) bool {
		role, ok := fl.Field().Interface().(rbac.UserRole)
		if !ok {
			return false
		}
		return role.IsValid()
	})

	//register category validation
	v.RegisterValidation("category", func(fl validator.FieldLevel) bool {
		category, ok := fl.Field().Interface().(core.Category)
		if !ok {
			return false
		}
		return category.IsValid()
	})

	return v
}
