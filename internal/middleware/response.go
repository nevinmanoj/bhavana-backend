package middleware

import (
	"encoding/json"
	"net/http"

	. "github.com/nevinmanoj/bhavana-backend/api"
)

// writeError answers with the same ErrorResponse envelope the handlers use,
// but with the real HTTP status as well - the UI relies on the HTTP 401 to
// detect an expired or rejected token and route the user back to /login.
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		StatusCode: status,
		Message:    message,
	})
}
