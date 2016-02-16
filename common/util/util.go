package util

import "strings"

// StrContains returns true if needle is contained in haystack.
func StrContains(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

// IsDevel returns true if the pkg is a devel package (ends on -{bzr,git,hg,svn}).
func IsDevel(pkg string) bool {
	if strings.HasSuffix(pkg, "-git") {
		return true
	}

	if strings.HasSuffix(pkg, "-svn") {
		return true
	}

	if strings.HasSuffix(pkg, "-hg") {
		return true
	}

	if strings.HasSuffix(pkg, "-bzr") {
		return true
	}

	return false
}
