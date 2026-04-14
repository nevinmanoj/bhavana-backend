package middleware

type contextKey string

const (
	ContextUserID   contextKey = "userID"
	ContextUserRole contextKey = "userRole"
	ContextScope    contextKey = "scope"
)
