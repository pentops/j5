package protobuild

import (
	"errors"
	"strings"
)

func hasAPrefix(s string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

var ErrNotFound = errors.New("file not found")
