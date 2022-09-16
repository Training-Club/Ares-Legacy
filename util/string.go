package util

import "strings"

// GenerateSlug accepts a string and returns a slug-style
// string
//
// Example: Example Blog Post
// 			to: example-blog-post
func GenerateSlug(s string) string {
	replaced := strings.ReplaceAll(s, " ", "-")
	replaced = strings.ReplaceAll(replaced, "'", "")
	return strings.ToLower(replaced)
}
