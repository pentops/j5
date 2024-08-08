// Code generated by protoc-gen-go-sugar. DO NOT EDIT.

package realm_j5pb

import (
	driver "database/sql/driver"
	fmt "fmt"
)

// RealmEventType is a oneof wrapper
type RealmEventTypeKey string

const (
	RealmEvent_Created RealmEventTypeKey = "created"
	RealmEvent_Updated RealmEventTypeKey = "updated"
)

func (x *RealmEventType) TypeKey() (RealmEventTypeKey, bool) {
	switch x.Type.(type) {
	case *RealmEventType_Created_:
		return RealmEvent_Created, true
	case *RealmEventType_Updated_:
		return RealmEvent_Updated, true
	default:
		return "", false
	}
}

type IsRealmEventTypeWrappedType interface {
	TypeKey() RealmEventTypeKey
}

func (x *RealmEventType) Set(val IsRealmEventTypeWrappedType) {
	switch v := val.(type) {
	case *RealmEventType_Created:
		x.Type = &RealmEventType_Created_{Created: v}
	case *RealmEventType_Updated:
		x.Type = &RealmEventType_Updated_{Updated: v}
	}
}
func (x *RealmEventType) Get() IsRealmEventTypeWrappedType {
	switch v := x.Type.(type) {
	case *RealmEventType_Created_:
		return v.Created
	case *RealmEventType_Updated_:
		return v.Updated
	default:
		return nil
	}
}
func (x *RealmEventType_Created) TypeKey() RealmEventTypeKey {
	return RealmEvent_Created
}
func (x *RealmEventType_Updated) TypeKey() RealmEventTypeKey {
	return RealmEvent_Updated
}

type IsRealmEventType_Type = isRealmEventType_Type

// RealmStatus
const (
	RealmStatus_UNSPECIFIED RealmStatus = 0
	RealmStatus_ACTIVE      RealmStatus = 1
)

var (
	RealmStatus_name_short = map[int32]string{
		0: "UNSPECIFIED",
		1: "ACTIVE",
	}
	RealmStatus_value_short = map[string]int32{
		"UNSPECIFIED": 0,
		"ACTIVE":      1,
	}
	RealmStatus_value_either = map[string]int32{
		"UNSPECIFIED":              0,
		"REALM_STATUS_UNSPECIFIED": 0,
		"ACTIVE":                   1,
		"REALM_STATUS_ACTIVE":      1,
	}
)

// ShortString returns the un-prefixed string representation of the enum value
func (x RealmStatus) ShortString() string {
	return RealmStatus_name_short[int32(x)]
}
func (x RealmStatus) Value() (driver.Value, error) {
	return []uint8(x.ShortString()), nil
}
func (x *RealmStatus) Scan(value interface{}) error {
	var strVal string
	switch vt := value.(type) {
	case []uint8:
		strVal = string(vt)
	case string:
		strVal = vt
	default:
		return fmt.Errorf("invalid type %T", value)
	}
	val := RealmStatus_value_either[strVal]
	*x = RealmStatus(val)
	return nil
}