package util

import "regexp"

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func IsSlug(s string) bool {
	return slugRegex.Match([]byte(s))
}
