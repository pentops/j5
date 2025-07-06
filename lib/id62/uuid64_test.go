package id62

import (
	"testing"
)

func TestEdges(t *testing.T) {
	tcs := []struct {
		name       string
		uuid       UUID
		expId62Str string
		expB64Str  string
	}{
		{
			name:       "Zero",
			uuid:       UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expId62Str: "0000000000000000000000",
			expB64Str:  "AAAAAAAAAAAAAAAAAAAAAA==",
		},
		{
			name:       "Leading One",
			uuid:       UUID{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expId62Str: "01SZwviYzes2mjOamuMJWw",
			expB64Str:  "AQAAAAAAAAAAAAAAAAAAAA==",
		},
		{
			name:       "One",
			uuid:       UUID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			expId62Str: "0000000000000000000001",
			expB64Str:  "AAAAAAAAAAAAAAAAAAAAAQ==",
		},
		{
			name:       "Max",
			uuid:       UUID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			expId62Str: "7N42dgm5tFLK9N8MT7fHC7",
			expB64Str:  "/////////////////////w==",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			str := tc.uuid.String()
			t.Logf("uuid: %s", tc.uuid.UUIDString())
			t.Logf("id62: %s", str)
			t.Logf("base64: %s", tc.uuid.Base64String())

			// Regex match
			if !Pattern.MatchString(str) {
				t.Errorf("expected match, got no match")
			}

			// Deterministic output
			if str != tc.expId62Str {
				t.Errorf("expected %q, got %q", tc.expId62Str, str)
			}

			if tc.uuid.Base64String() != tc.expB64Str {
				t.Errorf("expB64 %q, got %q", tc.expB64Str, tc.uuid.Base64String())
			}

			// Can parse from ID62 string
			asID, err := Parse(str)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Parsed matches original
			if asID != tc.uuid {
				t.Errorf("expected %v, got %v", tc.uuid, asID)
			}

			// Can parse from UUID string
			asID, err = ParseUUID(tc.uuid.UUIDString())
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Parsed matches original
			if asID != tc.uuid {
				t.Errorf("expected %v, got %v", tc.uuid, asID)
			}
		})
	}
}
