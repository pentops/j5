package extenders

import (
	"fmt"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
)

type UnknownExtender struct{}

func (v UnknownExtender) do(key, typename string) error {
	return fmt.Errorf("unexpected key %s on %s", key, typename)
}

func (v UnknownExtender) FieldLevel(f *schema_j5pb.ObjectProperty, key string, value Value) (bool, error) {
	return false, nil
}

func (v UnknownExtender) String(f *schema_j5pb.StringField, key string, value Value) error {
	return v.do(key, "string")
}

func (v UnknownExtender) Object(f *schema_j5pb.ObjectField, key string, value Value) error {
	return v.do(key, "object")
}

func (v UnknownExtender) Enum(f *schema_j5pb.EnumField, key string, value Value) error {
	return v.do(key, "enum")
}

func (v UnknownExtender) Any(f *schema_j5pb.AnyField, key string, value Value) error {
	return v.do(key, "any")
}

func (v UnknownExtender) Oneof(f *schema_j5pb.OneofField, key string, value Value) error {
	return v.do(key, "oneof")
}

func (v UnknownExtender) Array(f *schema_j5pb.ArrayField, key string, value Value) error {
	return v.do(key, "array")
}

func (v UnknownExtender) Map(f *schema_j5pb.MapField, key string, value Value) error {
	return v.do(key, "map")
}

func (v UnknownExtender) Integer(f *schema_j5pb.IntegerField, key string, value Value) error {
	return v.do(key, "integer")
}

func (v UnknownExtender) Float(f *schema_j5pb.FloatField, key string, value Value) error {
	return v.do(key, "float")
}

func (v UnknownExtender) Boolean(f *schema_j5pb.BooleanField, key string, value Value) error {
	return v.do(key, "boolean")
}

func (v UnknownExtender) Bytes(f *schema_j5pb.BytesField, key string, value Value) error {
	return v.do(key, "bytes")
}

func (v UnknownExtender) Decimal(f *schema_j5pb.DecimalField, key string, value Value) error {
	return v.do(key, "decimal")
}

func (v UnknownExtender) Date(f *schema_j5pb.DateField, key string, value Value) error {
	return v.do(key, "date")
}

func (v UnknownExtender) Timestamp(f *schema_j5pb.TimestampField, key string, value Value) error {
	return v.do(key, "timestamp")
}

func (v UnknownExtender) Key(f *schema_j5pb.KeyField, key string, value Value) error {
	return v.do(key, "key")
}
