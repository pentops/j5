// Code generated by protoc-gen-go-sugar. DO NOT EDIT.

package bar_testpb

import (
	driver "database/sql/driver"
	fmt "fmt"
)

// BarEnum
const (
	BarEnum_UNSPECIFIED BarEnum = 0
	BarEnum_FOO         BarEnum = 1
	BarEnum_BAR         BarEnum = 2
)

var (
	BarEnum_name_short = map[int32]string{
		0: "UNSPECIFIED",
		1: "FOO",
		2: "BAR",
	}
	BarEnum_value_short = map[string]int32{
		"UNSPECIFIED": 0,
		"FOO":         1,
		"BAR":         2,
	}
	BarEnum_value_either = map[string]int32{
		"UNSPECIFIED":          0,
		"BAR_ENUM_UNSPECIFIED": 0,
		"FOO":                  1,
		"BAR_ENUM_FOO":         1,
		"BAR":                  2,
		"BAR_ENUM_BAR":         2,
	}
)

// ShortString returns the un-prefixed string representation of the enum value
func (x BarEnum) ShortString() string {
	return BarEnum_name_short[int32(x)]
}
func (x BarEnum) Value() (driver.Value, error) {
	return []uint8(x.ShortString()), nil
}
func (x *BarEnum) Scan(value interface{}) error {
	var strVal string
	switch vt := value.(type) {
	case []uint8:
		strVal = string(vt)
	case string:
		strVal = vt
	default:
		return fmt.Errorf("invalid type %T", value)
	}
	val := BarEnum_value_either[strVal]
	*x = BarEnum(val)
	return nil
}
