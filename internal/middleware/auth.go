package middleware

import (
	"context"

	"net/http"

	auth "github.com/nevinmanoj/bhavana-backend/internal/auth"
)

func Authorization(jwtSecret []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var token = r.Header.Get("Authorization")
			claims, err := auth.ParseToken(token, jwtSecret)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ContextUserKey, claims.UserID)
			ctx = context.WithValue(ctx, ContextUserRole, claims.Role)
			// var userid int64 = 38
			// ctx := context.WithValue(r.Context(), ContextUserKey, userid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
