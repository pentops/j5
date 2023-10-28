package codec

type Options struct {
	// When using the buf convention of naming enum values with the type as a
	// prefix, enum Foo { FOO_UNSPECIFIED = 0; FOO_BAR = 1; } we can strip the
	// prefix when encoding to JSON.
	// When decoding, either option will work.
	ShortEnums bool

	// This specifies the suffix used to discover the prefix to strip from enum
	// values when encoding to JSON. For example, if the suffix is "_UNSPECIFIED"
	// then the prefix will be "FOO_" for the enum value "FOO_UNSPECIFIED".
	// This prefix is then dropped from other values, so FOO_BAR becomes BAR.
	// This is only used if ShortEnums is true.
	// Defaults to _UNSPECIFIED when ShortEnums is set
	UnspecifiedEnumSuffix string
}
