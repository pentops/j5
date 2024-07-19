package testlib

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func AssertEqualProto(t *testing.T, want, got proto.Message) {
	t.Helper()
	wantJSON := protojson.Format(want)
	gotJSON := protojson.Format(got)
	if string(wantJSON) == string(gotJSON) {
		t.Log("STRINGS MATCH")
	}

	matched := true

	lineA := 0
	lineB := 0

	wantLines := strings.Split(string(wantJSON), "\n")
	gotLines := strings.Split(string(gotJSON), "\n")
	for {
		if lineA >= len(wantLines) || lineB >= len(gotLines) {
			break
		}
		wantLine := string(wantLines[lineA])
		gotLine := string(gotLines[lineB])
		if wantLine != gotLine {
			matched = false
			t.Logf("    W: %s", wantLine)
			t.Logf("    G: %s", gotLine)
			t.Log(strings.Repeat(`/\`, 10))

			break
		} else {
			t.Logf("   OK: %s", wantLine)
		}
		lineA++
		lineB++
	}
	if lineA < len(wantLines) {
		matched = false
		t.Logf("Remaining Want: \n%s", strings.Join(wantLines[lineA:], "  \n"))
	}
	if lineB < len(gotLines) {
		matched = false
		t.Logf("Remaining Got: \n%s", strings.Join(gotLines[lineB:], "  \n"))
	}

	if !matched {
		t.Errorf("unexpected JSON")
	}

}
