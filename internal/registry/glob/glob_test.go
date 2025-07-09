package glob

import (
	"fmt"
	"testing"
)

func TestWildcardMatcher(t *testing.T) {

	for _, spec := range []struct {
		input    string
		pattern  string
		expected bool
	}{
		{"foo", "foo", true},
		{"foo", "bar", false},
		{"foo", "f*", true},
		{"foo", "b*", false},
		{"foo", "f*o", true},
		{"foo", "b*r", false},
		{"foo/bar", "foo/bar", true},
		{"foo/bar", "foo/*", true},
		{"foo/bar", "foo/baz", false},
		{"foo/bar/baz", "foo/*/*", true},
		{"foo/*/bar", "foo/*/bar", true},
		{"foo/*/bar", "foo/*/*", true},
		{"acct.dkr.ecr.us-east-1.amazonaws.com/packageo:version", "*.dkr.ecr.*.amazonaws.com/*", true},
	} {
		t.Run(fmt.Sprintf("%s-%s", spec.input, spec.pattern), func(t *testing.T) {

			gotMatch := GlobMatch(spec.pattern, spec.input)
			if spec.expected && !gotMatch {
				t.Errorf("expected %s to match %s", spec.input, spec.pattern)
			} else if !spec.expected && gotMatch {
				t.Errorf("expected %s to not match %s", spec.input, spec.pattern)
			}

		})
	}

}
