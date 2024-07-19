// Code generated by protoc-gen-go-sugar. DO NOT EDIT.

package schema_j5pb

import (
	driver "database/sql/driver"
	fmt "fmt"
)

type IsRootSchema_Type = isRootSchema_Type
type IsField_Type = isField_Type
type IsObjectField_Schema = isObjectField_Schema
type IsOneofField_Schema = isOneofField_Schema
type IsEnumField_Schema = isEnumField_Schema

// EntityPart
const (
	EntityPart_UNSPECIFIED EntityPart = 0
	EntityPart_KEYS        EntityPart = 1
	EntityPart_STATE       EntityPart = 2
	EntityPart_EVENT       EntityPart = 3
)

var (
	EntityPart_name_short = map[int32]string{
		0: "UNSPECIFIED",
		1: "KEYS",
		2: "STATE",
		3: "EVENT",
	}
	EntityPart_value_short = map[string]int32{
		"UNSPECIFIED": 0,
		"KEYS":        1,
		"STATE":       2,
		"EVENT":       3,
	}
	EntityPart_value_either = map[string]int32{
		"UNSPECIFIED":             0,
		"ENTITY_PART_UNSPECIFIED": 0,
		"KEYS":                    1,
		"ENTITY_PART_KEYS":        1,
		"STATE":                   2,
		"ENTITY_PART_STATE":       2,
		"EVENT":                   3,
		"ENTITY_PART_EVENT":       3,
	}
)

// ShortString returns the un-prefixed string representation of the enum value
func (x EntityPart) ShortString() string {
	return EntityPart_name_short[int32(x)]
}
func (x EntityPart) Value() (driver.Value, error) {
	return []uint8(x.ShortString()), nil
}
func (x *EntityPart) Scan(value interface{}) error {
	var strVal string
	switch vt := value.(type) {
	case []uint8:
		strVal = string(vt)
	case string:
		strVal = vt
	default:
		return fmt.Errorf("invalid type %T", value)
	}
	val := EntityPart_value_either[strVal]
	*x = EntityPart(val)
	return nil
}

// KeyFormat
const (
	KeyFormat_UNSPECIFIED KeyFormat = 0
	KeyFormat_UUID        KeyFormat = 1
)

var (
	KeyFormat_name_short = map[int32]string{
		0: "UNSPECIFIED",
		1: "UUID",
	}
	KeyFormat_value_short = map[string]int32{
		"UNSPECIFIED": 0,
		"UUID":        1,
	}
	KeyFormat_value_either = map[string]int32{
		"UNSPECIFIED":            0,
		"KEY_FORMAT_UNSPECIFIED": 0,
		"UUID":                   1,
		"KEY_FORMAT_UUID":        1,
	}
)

// ShortString returns the un-prefixed string representation of the enum value
func (x KeyFormat) ShortString() string {
	return KeyFormat_name_short[int32(x)]
}
func (x KeyFormat) Value() (driver.Value, error) {
	return []uint8(x.ShortString()), nil
}
func (x *KeyFormat) Scan(value interface{}) error {
	var strVal string
	switch vt := value.(type) {
	case []uint8:
		strVal = string(vt)
	case string:
		strVal = vt
	default:
		return fmt.Errorf("invalid type %T", value)
	}
	val := KeyFormat_value_either[strVal]
	*x = KeyFormat(val)
	return nil
}
