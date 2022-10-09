package model

type Permission string

const (
	MODERATE_USERS   Permission = "moderate_users"
	MODERATE_POSTS   Permission = "moderate_posts"
	AUTHOR_BLOGS     Permission = "author_blogs"
	AUTHOR_EXERCISES Permission = "author_exercises"
	AUTHOR_FOOD      Permission = "author_food"
	AUTHOR_LOCATION  Permission = "author_location"
	VIEW_AUDIT       Permission = "view_audit"
	BYPASS_PRIVACY   Permission = "bypass_privacy"
)
