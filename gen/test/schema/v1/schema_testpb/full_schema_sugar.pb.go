// Code generated by protoc-gen-go-sugar. DO NOT EDIT.

package schema_testpb

import (
	driver "database/sql/driver"
	fmt "fmt"
)

type IsFullSchema_AnonOneof = isFullSchema_AnonOneof
type IsFullSchema_ExposedOneof = isFullSchema_ExposedOneof
type IsWrappedOneof_Type = isWrappedOneof_Type
type IsImplicitOneof_Type = isImplicitOneof_Type
type IsNestedExposed_Type = isNestedExposed_Type

// Enum
const (
	Enum_UNSPECIFIED Enum = 0
	Enum_VALUE1      Enum = 1
	Enum_VALUE2      Enum = 2
)

var (
	Enum_name_short = map[int32]string{
		0: "UNSPECIFIED",
		1: "VALUE1",
		2: "VALUE2",
	}
	Enum_value_short = map[string]int32{
		"UNSPECIFIED": 0,
		"VALUE1":      1,
		"VALUE2":      2,
	}
	Enum_value_either = map[string]int32{
		"UNSPECIFIED":      0,
		"ENUM_UNSPECIFIED": 0,
		"VALUE1":           1,
		"ENUM_VALUE1":      1,
		"VALUE2":           2,
		"ENUM_VALUE2":      2,
	}
)

// ShortString returns the un-prefixed string representation of the enum value
func (x Enum) ShortString() string {
	return Enum_name_short[int32(x)]
}
func (x Enum) Value() (driver.Value, error) {
	return []uint8(x.ShortString()), nil
}
func (x *Enum) Scan(value interface{}) error {
	var strVal string
	switch vt := value.(type) {
	case []uint8:
		strVal = string(vt)
	case string:
		strVal = vt
	default:
		return fmt.Errorf("invalid type %T", value)
	}
	val := Enum_value_either[strVal]
	*x = Enum(val)
	return nil
}
