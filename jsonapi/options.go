package jsonapi

import (
	"fmt"
	"regexp"
	"strings"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Options struct {
	// When using the buf convention of naming enum values with the type as a
	// prefix, enum Foo { FOO_UNSPECIFIED = 0; FOO_BAR = 1; } we can strip the
	// prefix when encoding to JSON.
	ShortEnums *ShortEnumsOption

	// Promotes the oneof constraint into a visible key in the JSON output.
	// In standard protojson, oneof is just a constraint and doesn't appear in
	// the JSON structure. The oneof name will be a
	// field in the JSON output with exactly one key set.
	// When messages are marked as sugar.v1.message.oneof_wrapper = true, this
	// is not the case (as the wrapper already creates this structure)
	// e.g. message Foo { oneof bar { int32 baz = 1; } }
	// becomes { "bar": { "baz": 1 } }
	// instead of { "baz": 1 }
	WrapOneof bool
}

type ShortEnumsOption struct {
	// This specifies the suffix used to discover the prefix to strip from enum
	// values when encoding to JSON. For example, if the suffix is "_UNSPECIFIED"
	// then the prefix will be "FOO_" for the enum value "FOO_UNSPECIFIED".
	// This prefix is then dropped from other values, so FOO_BAR becomes BAR.
	// This is only used if ShortEnums is true.
	// Defaults to _UNSPECIFIED
	UnspecifiedSuffix string

	// When decoding, either long or short will work unless this is set
	StrictUnmarshal bool
}

type EnumValueDescription struct {
	Name        string `json:"name"`
	Number      int32  `json:"number"`
	Description string `json:"description"`
}

func (se *ShortEnumsOption) EnumValues(src protoreflect.EnumValueDescriptors, constraint *validate.EnumRules) ([]EnumValueDescription, error) {

	specMap := map[int32]struct{}{}
	var notIn bool
	var isIn bool

	if constraint != nil {
		if constraint.NotIn != nil {
			for _, notIn := range constraint.NotIn {
				specMap[notIn] = struct{}{}
			}
			notIn = true

		} else if constraint.In != nil {
			for _, in := range constraint.In {
				specMap[in] = struct{}{}
			}
			isIn = true
		}
	}

	if notIn && isIn {
		return nil, fmt.Errorf("enum cannot have both in and not_in constraints")
	}

	values := make([]EnumValueDescription, 0, src.Len())
	for ii := 0; ii < src.Len(); ii++ {
		option := src.Get(ii)
		number := int32(option.Number())

		if notIn {
			_, exclude := specMap[number]
			if exclude {
				continue
			}
		} else if isIn {
			_, include := specMap[number]
			if !include {
				continue
			}
		}

		values = append(values, EnumValueDescription{
			Name:        string(option.Name()),
			Number:      number,
			Description: commentDescription(option, ""),
		})
	}

	if se == nil {
		return values, nil
	}

	suffix := se.unspecifiedSuffix()
	unspecifiedVal := string(src.Get(0).Name())
	if !strings.HasSuffix(unspecifiedVal, suffix) {
		return nil, fmt.Errorf("enum does not have an unspecified value ending in %q", suffix)
	}
	trimPrefix := strings.TrimSuffix(unspecifiedVal, suffix)

	for ii := range values {
		values[ii].Name = strings.TrimPrefix(values[ii].Name, trimPrefix)
	}
	return values, nil
}

func (se *ShortEnumsOption) unspecifiedSuffix() string {
	if se.UnspecifiedSuffix != "" {
		return se.UnspecifiedSuffix
	}
	return "UNSPECIFIED"
}

func (se *ShortEnumsOption) Decode(enum protoreflect.EnumDescriptor, stringVal string) (protoreflect.EnumNumber, error) {

	vals := enum.Values()

	if se == nil {
		enumVal := vals.ByName(protoreflect.Name(stringVal))
		if enumVal == nil {
			return 0, fmt.Errorf("unknown enum value %s for enum %s", stringVal, enum.FullName())
		}
		return enumVal.Number(), nil
	}

	unspecified := vals.ByNumber(0)
	if unspecified != nil {
		unspecifiedSuffix := se.unspecifiedSuffix()
		unspecifiedName := string(unspecified.Name())
		if strings.HasSuffix(unspecifiedName, unspecifiedSuffix) {
			prefix := strings.TrimSuffix(unspecifiedName, unspecifiedSuffix)
			if se.StrictUnmarshal || !strings.HasPrefix(stringVal, prefix) {
				stringVal = prefix + stringVal
			}
		}
	}

	enumVal := vals.ByName(protoreflect.Name(stringVal))
	if enumVal == nil {
		return 0, fmt.Errorf("unknown enum value %s for enum %s", stringVal, enum.FullName())
	}
	return enumVal.Number(), nil

}

func (se *ShortEnumsOption) Encode(enum protoreflect.EnumDescriptor, enumVal protoreflect.EnumNumber) (string, error) {
	vals := enum.Values()
	fullStringValue := string(vals.ByNumber(enumVal).Name())
	if se == nil {
		return fullStringValue, nil
	}

	// TODO: Cache all of this
	unspecified := vals.ByNumber(0)
	if unspecified == nil {
		return "", fmt.Errorf("enum %s has no unspecified value", enum.FullName())
	}

	unspecifiedSuffix := se.unspecifiedSuffix()
	unspecifiedName := string(unspecified.Name())

	if !strings.HasSuffix(unspecifiedName, unspecifiedSuffix) {
		return "", fmt.Errorf("enum %s has unspecified value %s without suffix %s", enum.FullName(), unspecifiedName, unspecifiedSuffix)
	}

	prefix := strings.TrimSuffix(unspecifiedName, unspecifiedSuffix)
	// End Cache TODO

	return strings.TrimPrefix(fullStringValue, prefix), nil
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
func protoNameToJSON(s string) string {
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
