package codec

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Codec struct {
}

func NewCodec() *Codec {
	return &Codec{}
}

func (c *Codec) ToProto(jsonData []byte, msg protoreflect.Message) error {
	return Decode(jsonData, msg)
}

func (c *Codec) FromProto(msg protoreflect.Message) ([]byte, error) {
	return Encode(msg)
}
func enumPrefix(enum protoreflect.EnumDescriptor) string {
	// TODO: Cache all of this
	unspecified := enum.Values().ByNumber(0)
	prefix := ""
	if unspecified == nil {
		return ""
	}

	unspecifiedSuffix := "_UNSPECIFIED"
	unspecifiedName := string(unspecified.Name())

	if strings.HasSuffix(unspecifiedName, unspecifiedSuffix) {
		prefix = fmt.Sprintf("%s_", strings.TrimSuffix(unspecifiedName, unspecifiedSuffix))
	} else {
		parts := strings.Split(unspecifiedName, "_")
		if len(parts) < 2 {
			return ""
		}
		suffix := parts[len(parts)-1]
		prefix = strings.TrimSuffix(unspecifiedName, suffix)
	}

	return prefix
}

var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func jsonNameToProto(str string) string {
	snake := matchAllCap.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}

// Is c an ASCII lower-case letter?
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

// protoNameToJSON returns the CamelCased name.
// Copied from protoc-gen-go/generator/generator.go which is now deprecated
// but changed the first letter to lower case.
func protoNameToJSON(s protoreflect.Name) string {
	if s == "" {
		return ""
	}
	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		// Need a capital letter; drop the '_'.
		t = append(t, 'X')
		i++
	}

	firstWord := true

	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		// Assume we have a letter now - if not, it's a bogus identifier.
		// The next word is a sequence of characters that must start upper case.
		if isASCIILower(c) && !firstWord {
			c ^= ' ' // Make it a capital letter.
		}
		firstWord = false
		t = append(t, c) // Guaranteed not lower case.
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}
