package rbac

const (
	//users
	PermViewUser Permission = "user:view"

	// events
	PermCreateEvent Permission = "event:create"
	PermUpdateEvent Permission = "event:update"
	PermDeleteEvent Permission = "event:delete"
	PermViewEvent   Permission = "event:view"
)
