package util

import "regexp"

// IsAlphanumeric returns a bool if the provided string
// matches alphanumeric regexp
func IsAlphanumeric(s string) bool {
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	return re.MatchString(s)
}

// IsAlphanumericWithWhitespace returns a bool if the provided
// string matches alphanumeric regexp with spaces allowed
func IsAlphanumericWithWhitespace(s string) bool {
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]+$`)
	return re.MatchString(s)
}

// IsAlphanumericWithDashes returns a bool if the provided
// string matches alphanumeric regexp with dashes allowed
func IsAlphanumericWithDashes(s string) bool {
	re := regexp.MustCompile(`[/^[\w-]+$/]`)
	return re.MatchString(s)
}

// IsValidPassword returns a bool if the provided string
// matches valid password regexp
func IsValidPassword(s string) bool {
	re := regexp.MustCompile(`/[ A-Za-z\d_@./#&+-]*/`)
	return re.MatchString(s)
}

// IsValidEmail returns a bool if the provided string
// matches valid email regexp
func IsValidEmail(s string) bool {
	re := regexp.MustCompile(`!/(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}])|(([a-zA-Z\-\d]+\.)+[a-zA-Z]{2,}))/`)
	return re.MatchString(s)
}

// IsValidWord returns a bool if the provided string
// matches a valid a-z A-Z format
func IsValidWord(s string) bool {
	re := regexp.MustCompile(`[^a-zA-Z]+`)
	return re.MatchString(s)
}
