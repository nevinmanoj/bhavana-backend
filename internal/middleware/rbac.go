package middleware

import (
	"net/http"

	"github.com/nevinmanoj/bhavana-backend/internal/rbac"
)

func RequirePermission(perm rbac.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userRole := ctx.Value(ContextUserRole).(rbac.UserRole)
			if !rbac.HasPermission(userRole, perm) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
