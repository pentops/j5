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

	linesA := strings.Split(string(wantJSON), "\n")
	linesB := strings.Split(string(gotJSON), "\n")
	for {
		if lineA >= len(linesA) || lineB >= len(linesB) {
			break
		}
		wantLine := string(linesA[lineA])
		gotLine := string(linesB[lineB])
		if wantLine == gotLine {
			t.Logf("   OK: %s", wantLine)
			lineA++
			lineB++
			continue
		}

		matched = false
		t.Logf("    W: %s", wantLine)
		t.Logf("    G: %s", gotLine)
		t.Log(strings.Repeat(`/\`, 10))

		if lineA+1 < len(linesA) && lineB+1 < len(linesB) {
			if linesA[lineA+1] == linesB[lineB+1] {
				lineA++
				lineB++
				continue
			}
		}
		break
	}

	if lineA < len(linesA) {
		matched = false
		t.Logf("Remaining Want: \n%s", strings.Join(linesA[lineA:], "  \n"))
	}
	if lineB < len(linesB) {
		matched = false
		t.Logf("Remaining Got: \n%s", strings.Join(linesB[lineB:], "  \n"))
	}

	if !matched {
		t.Errorf("unexpected JSON")
	}

}
