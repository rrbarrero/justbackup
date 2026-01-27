package utils

import (
	"regexp"
	"strings"
)

// Slugify converts a string into a URL-friendly slug.
// It converts to lowercase, replaces spaces with hyphens, and removes non-alphanumeric characters.
func Slugify(s string) string {
	// Lowercase
	s = strings.ToLower(s)
	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	// Remove non-alphanumeric characters (except hyphens)
	reg, err := regexp.Compile("[^a-z0-9-]+")
	if err != nil {
		return s // Should not happen
	}
	s = reg.ReplaceAllString(s, "")
	return s
}
