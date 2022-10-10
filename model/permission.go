package model

type Permission string

const (
	MODERATE_USERS    Permission = "moderate_users"
	MODERATE_POSTS    Permission = "moderate_posts"
	AUTHOR_BLOGS      Permission = "author_blogs"
	AUTHOR_EXERCISES  Permission = "author_exercises"
	AUTHOR_FOOD       Permission = "author_food"
	AUTHOR_LOCATION   Permission = "author_location"
	VIEW_AUDIT        Permission = "view_audit"
	VIEW_ROLES        Permission = "view_roles"
	VIEW_PERMISSIONS  Permission = "view_permissions"
	BYPASS_PRIVACY    Permission = "bypass_privacy"
	GRANT_PERMISSIONS Permission = "grant_permissions"
	GRANT_ROLES       Permission = "grant_roles"
)

// GetAllPermissions returns all permissions defined as a slice
func GetAllPermissions() []Permission {
	return []Permission{
		MODERATE_USERS,
		MODERATE_POSTS,
		AUTHOR_BLOGS,
		AUTHOR_EXERCISES,
		AUTHOR_FOOD,
		AUTHOR_LOCATION,
		VIEW_AUDIT,
		VIEW_ROLES,
		VIEW_PERMISSIONS,
		BYPASS_PRIVACY,
		GRANT_PERMISSIONS,
		GRANT_ROLES,
	}
}
