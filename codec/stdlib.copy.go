package codec

import (
	"errors"
	"math/bits"
	"strconv"
	"unicode/utf8"
)

// copied from google.golang.org/protobuf/encoding/protojson/encode.go
var errInvalidUTF8 = errors.New("invalid UTF-8")

// copied from google.golang.org/protobuf/encoding/protojson/encode.go
func appendString(out []byte, in string) ([]byte, error) {
	out = append(out, '"')
	i := indexNeedEscapeInString(in)
	in, out = in[i:], append(out, in[:i]...)
	for len(in) > 0 {
		switch r, n := utf8.DecodeRuneInString(in); {
		case r == utf8.RuneError && n == 1:
			return out, errInvalidUTF8
		case r < ' ' || r == '"' || r == '\\':
			out = append(out, '\\')
			switch r {
			case '"', '\\':
				out = append(out, byte(r))
			case '\b':
				out = append(out, 'b')
			case '\f':
				out = append(out, 'f')
			case '\n':
				out = append(out, 'n')
			case '\r':
				out = append(out, 'r')
			case '\t':
				out = append(out, 't')
			default:
				out = append(out, 'u')
				out = append(out, "0000"[1+(bits.Len32(uint32(r))-1)/4:]...)
				out = strconv.AppendUint(out, uint64(r), 16)
			}
			in = in[n:]
		default:
			i := indexNeedEscapeInString(in[n:])
			in, out = in[n+i:], append(out, in[:n+i]...)
		}
	}
	out = append(out, '"')
	return out, nil
}

// copied from google.golang.org/protobuf/encoding/protojson/encode.go
func indexNeedEscapeInString(s string) int {
	for i, r := range s {
		if r < ' ' || r == '\\' || r == '"' || r == utf8.RuneError {
			return i
		}
	}
	return len(s)
}
