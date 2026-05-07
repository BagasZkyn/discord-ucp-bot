package utils

import "regexp"

var ucpRegex = regexp.MustCompile(`^[a-zA-Z0-9]{3,20}$`)

// IsValidUCP memvalidasi nama UCP: hanya alfanumerik, 3-20 karakter
func IsValidUCP(name string) bool {
	return ucpRegex.MatchString(name)
}
