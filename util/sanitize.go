package util

import "regexp"

// IsAlphanumeric returns a bool if the provided string
// matches alphanumeric regexp
func IsAlphanumeric(s string) bool {
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
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
