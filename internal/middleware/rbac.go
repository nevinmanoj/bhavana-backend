package middleware

import (
	"context"
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

func InjectScope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := ctx.Value(ContextUserID).(int64)
		userRole := ctx.Value(ContextUserRole).(rbac.UserRole)
		scope := rbac.ResolveScope(userID, userRole)
		ctx = context.WithValue(ctx, ContextScope, scope)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
