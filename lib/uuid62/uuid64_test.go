package uuid62

import (
	"fmt"
	"testing"
)

func TestEdges(t *testing.T) {

	for _, tc := range []struct {
		uuid UUID
		str  string
	}{{
		uuid: UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		str:  "0000000000000000000000",
	}, {
		uuid: UUID{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}, {
		uuid: UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		str:  "0000000000000000000001",
	}, {
		uuid: UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
	}} {
		t.Run(fmt.Sprintf("%x", tc.uuid), func(t *testing.T) {
			asStr := tc.uuid.String()
			t.Logf("asStr: %q", asStr)
			if tc.str != "" {
				if asStr != tc.str {
					t.Errorf("expected %q, got %q", tc.str, asStr)
				}
			}
			if len(asStr) != 22 {
				t.Errorf("expected 22, got %d", len(asStr))
			}
			if !Pattern.MatchString(asStr) {
				t.Errorf("expected match, got no match")
			}
			asID, err := Parse(asStr)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if asID != tc.uuid {
				t.Errorf("expected %v, got %v", tc.uuid, asID)
			}
		})
	}

}
