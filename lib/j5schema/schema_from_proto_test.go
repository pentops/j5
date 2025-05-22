package j5schema

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestCommentBuilder(t *testing.T) {

	for _, tc := range []struct {
		name     string
		leading  string
		trailing string
		expected string
	}{{
		name:     "leading",
		leading:  "comment",
		expected: "comment",
	}, {
		name:     "fallback",
		expected: "fallback",
	}, {
		name:     "both",
		leading:  "leading",
		trailing: "trailing",
		expected: "leading\ntrailing",
	}, {
		name:     "multiline",
		leading:  "line1\n  line2",
		trailing: "line3\n  line4",
		expected: "line1\nline2\nline3\nline4",
	}, {
		name:     "multiline commented",
		leading:  "#line1\nline2",
		expected: "line2",
	}, {
		name:     "commented fallback",
		leading:  "#line1",
		expected: "fallback",
	}} {
		t.Run(tc.name, func(t *testing.T) {
			sl := protoreflect.SourceLocation{
				LeadingComments:  tc.leading,
				TrailingComments: tc.trailing,
			}

			got := buildComment(sl, "fallback")
			if got != tc.expected {
				t.Errorf("expected comment: '%s', got '%s'", tc.expected, got)
			}

		})
	}
}
