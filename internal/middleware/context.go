package middleware

type contextKey string

const (
	ContextUserKey  contextKey = "userID"
	ContextUserRole contextKey = "userRole"
)
