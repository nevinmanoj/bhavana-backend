package rbac

const (
	//users
	PermViewUser Permission = "user:view"

	// events
	PermCreateEvent Permission = "event:create"
	PermUpdateEvent Permission = "event:update"
	PermDeleteEvent Permission = "event:delete"
	PermViewEvent   Permission = "event:view"

	//schools
	PermCreateSchool Permission = "school:create"
	PermUpdateSchool Permission = "school:update"
	PermDeleteSchool Permission = "school:delete"
	PermViewSchool   Permission = "school:view"

	//students
	PermCreateStudent Permission = "student:create"
	PermUpdateStudent Permission = "student:update"
	PermDeleteStudent Permission = "student:delete"
	PermViewStudent   Permission = "student:view"
)
