package j5reflect

import (
	"fmt"
	"time"

	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5schema"
	"github.com/pentops/j5/j5types/date_j5t"
	"github.com/pentops/j5/j5types/decimal_j5t"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Property interface {
	Type() FieldType
	IsSet() bool
	JSONName() string
	Field() (Value, error)
}

type Value interface {
	Type() FieldType
	IsSet() bool
	protoValue
}

type RangeCallback func(Property) error

type FieldSet interface {
	RangeProperties(RangeCallback) error
	RangeSetProperties(RangeCallback) error
	GetOne() (Property, error)
}

type FieldType string

const (
	FieldTypeUnknown = FieldType("?")
	FieldTypeObject  = FieldType("object")
	FieldTypeOneof   = FieldType("oneof")
	FieldTypeEnum    = FieldType("enum")
	FieldTypeArray   = FieldType("array")
	FieldTypeMap     = FieldType("map")
	FieldTypeScalar  = FieldType("scalar")
)

type ObjectField interface {
	Value
	Object() (*Object, error)
}

type objectField struct {
	protoValue
	schema  *j5schema.ObjectField
	_object *Object
}

func (obj *objectField) Object() (*Object, error) {
	if obj._object == nil {
		built, err := newObject(obj.schema.Schema(), obj.getValue().Message())
		if err != nil {
			return nil, err
		}
		obj._object = built
	}
	return obj._object, nil
}

type OneofField interface {
	Value
	Oneof() (*Oneof, error)
}

type oneofField struct {
	protoValue
	schema *j5schema.OneofField
	_oneof *Oneof
}

func (field *oneofField) Oneof() (*Oneof, error) {
	if field._oneof == nil {
		obj, err := newOneof(field.schema.Schema(), field.getValue().Message())
		if err != nil {
			return nil, err
		}
		field._oneof = obj
	}

	return field._oneof, nil
}

type EnumField interface {
	Value
	GetValue() (*j5schema.EnumOption, error)
}

type enumField struct {
	protoValue
	schema *j5schema.EnumField
}

func (ef *enumField) GetValue() (*j5schema.EnumOption, error) {
	value := int32(ef.getValue().Enum())
	for _, val := range ef.schema.Schema().Options {
		if val.Number == value {
			return val, nil
		}
	}
	return nil, fmt.Errorf("enum value %d not found", value)
}

type ArrayField interface {
	Value
	Range(func(Value) error) error
}

type arrayField struct {
	protoValue
	schema *j5schema.ArrayField
}

func (field *arrayField) Range(cb func(Value) error) error {
	list := field.getValue().List()
	for i := 0; i < list.Len(); i++ {
		val := list.Get(i)
		wrapped := &listFieldValue{
			value:     val,
			list:      list,
			idx:       i,
			fieldType: schemaType(field.schema.Schema),
		}
		property, err := newValue(field.schema.Schema, wrapped)
		if err != nil {
			return err
		}

		err = cb(property)
		if err != nil {
			return err
		}
	}
	return nil
}

type MapField interface {
	Value
	Range(func(string, Value) error) error
}

type mapField struct {
	protoValue
	schema *j5schema.MapField
}

func (field *mapField) Range(cb func(string, Value) error) error {
	if !field.IsSet() {
		return nil
	}
	mapVal := field.getValue().Map()
	var outerErr error

	mapVal.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		keyStr := key.Value().String()

		wrapped := &mapFieldValue{
			value:     val,
			mapVal:    mapVal,
			key:       key,
			fieldType: schemaType(field.schema.Schema),
		}
		property, err := newValue(field.schema.Schema, wrapped)
		if err != nil {
			outerErr = err
			return false
		}
		outerErr = cb(keyStr, property)
		return outerErr == nil
	})
	return outerErr
}

type ScalarField interface {
	Value
	Schema() *j5schema.ScalarSchema
	GetInterface() interface{}
}

type scalarField struct {
	protoValue
	schema *j5schema.ScalarSchema
}

func (sf *scalarField) Schema() *j5schema.ScalarSchema {
	return sf.schema
}

type AnyField interface{}

type KeyField string

func (sf *scalarField) GetInterface() interface{} {
	val := sf.getValue()
	switch sf.schema.Proto.Type.(type) {
	case *schema_j5pb.Field_Any:
		return val.Interface()
	case *schema_j5pb.Field_Boolean:
		return val.Bool()
	case *schema_j5pb.Field_String_:
		return val.String()
	case *schema_j5pb.Field_Key:
		return val.String()
	case *schema_j5pb.Field_Integer:
		switch sf.schema.Proto.GetInteger().Format {
		case schema_j5pb.IntegerField_FORMAT_INT32:
			return int32(val.Int())
		case schema_j5pb.IntegerField_FORMAT_INT64:
			return val.Int()
		case schema_j5pb.IntegerField_FORMAT_UINT32:
			return uint32(val.Uint())
		case schema_j5pb.IntegerField_FORMAT_UINT64:
			return val.Uint()
		default:
			return val.Interface()
		}

	case *schema_j5pb.Field_Float:
		switch sf.schema.Proto.GetFloat().Format {
		case schema_j5pb.FloatField_FORMAT_FLOAT32:
			return float32(val.Float())
		case schema_j5pb.FloatField_FORMAT_FLOAT64:
			return val.Float()
		default:
			return val.Interface()
		}
	case *schema_j5pb.Field_Bytes:
		return val.Bytes()

	case *schema_j5pb.Field_Date:
		msg := val.Message()
		val := &date_j5t.Date{}
		pv := val.ProtoReflect()
		copyReflect(msg, pv)
		return pv

	case *schema_j5pb.Field_Decimal:
		msg := val.Message()
		val := &decimal_j5t.Decimal{}
		pv := val.ProtoReflect()
		copyReflect(msg, pv)
		return pv

	case *schema_j5pb.Field_Timestamp:
		msg := val.Message()
		seconds := msg.Get(msg.Descriptor().Fields().ByName("seconds")).Int()
		nanos := msg.Get(msg.Descriptor().Fields().ByName("nanos")).Int()
		t := time.Unix(seconds, nanos).In(time.UTC)
		return t

	default:
		return val.Interface()
	}

}

func copyReflect(a, b protoreflect.Message) {
	a.Range(func(fd protoreflect.FieldDescriptor, val protoreflect.Value) bool {
		b.Set(fd, val)
		return true
	})
}

type property struct {
	schema *j5schema.ObjectProperty
	value  protoValue

	_type   *FieldType
	_object *objectField
	_oneof  *oneofField
	_enum   *enumField
	_array  *arrayField
	_map    *mapField
	_scalar *scalarField
}

func (p *property) IsSet() bool {
	return p.value.IsSet()
}

func (p *property) JSONName() string {
	return p.schema.JSONName
}

func (prop *property) ObjectField() (ObjectField, error) {
	if prop._object == nil {
		schema, ok := prop.schema.Schema.(*j5schema.ObjectField)
		if !ok {
			return nil, fmt.Errorf("schema is a %T, not an object", prop.Type())
		}

		prop._object = &objectField{
			protoValue: prop.value,
			schema:     schema,
		}
	}
	return prop._object, nil
}

func (prop *property) OneofField() (OneofField, error) {
	if prop._oneof == nil {
		schema, ok := prop.schema.Schema.(*j5schema.OneofField)
		if !ok {
			return nil, fmt.Errorf("schema is a %T, not a oneof", prop.Type())
		}

		prop._oneof = &oneofField{
			protoValue: prop.value,
			schema:     schema,
		}
	}
	return prop._oneof, nil
}

func (prop *property) EnumField() (EnumField, error) {
	if prop._enum == nil {
		schema, ok := prop.schema.Schema.(*j5schema.EnumField)
		if !ok {
			return nil, fmt.Errorf("schema is a %T, not an enum", prop.Type())
		}
		prop._enum = &enumField{
			protoValue: prop.value,
			schema:     schema,
		}
	}
	return prop._enum, nil
}

func (prop *property) ArrayField() (ArrayField, error) {
	if prop._array == nil {
		schema, ok := prop.schema.Schema.(*j5schema.ArrayField)
		if !ok {
			return nil, fmt.Errorf("schema is a %T, not an array", prop.Type())
		}

		prop._array = &arrayField{
			protoValue: prop.value,
			schema:     schema,
		}
	}
	return prop._array, nil
}

func (prop *property) MapField() (MapField, error) {
	if prop._map == nil {
		schema, ok := prop.schema.Schema.(*j5schema.MapField)
		if !ok {
			return nil, fmt.Errorf("schema is a %T, not a map", prop.Type())
		}

		prop._map = &mapField{
			protoValue: prop.value,
			schema:     schema,
		}
	}
	return prop._map, nil
}

func (prop *property) ScalarField() (ScalarField, error) {
	if prop._scalar == nil {
		schema, ok := prop.schema.Schema.(*j5schema.ScalarSchema)
		if !ok {
			return nil, fmt.Errorf("schema is a %T, not a scalar", prop.Type())
		}

		prop._scalar = &scalarField{
			protoValue: prop.value,
			schema:     schema,
		}
	}
	return prop._scalar, nil
}

func (p *property) Type() FieldType {
	if p._type == nil {
		tt := schemaType(p.schema.Schema)
		p._type = &tt
	}
	return *p._type
}

func newValue(base j5schema.FieldSchema, value protoValue) (Value, error) {
	switch base := base.(type) {
	case *j5schema.ObjectField:
		return &objectField{schema: base, protoValue: value}, nil
	case *j5schema.OneofField:
		return &oneofField{schema: base, protoValue: value}, nil
	case *j5schema.EnumField:
		return &enumField{schema: base, protoValue: value}, nil
	case *j5schema.ArrayField:
		return &arrayField{schema: base, protoValue: value}, nil
	case *j5schema.MapField:
		return &mapField{schema: base, protoValue: value}, nil
	case *j5schema.ScalarSchema:
		return &scalarField{schema: base, protoValue: value}, nil
	default:
		return nil, fmt.Errorf("unsupported value type %T", base)
	}
}

func schemaType(base j5schema.FieldSchema) FieldType {
	switch base.(type) {
	case *j5schema.ObjectField:
		return FieldTypeObject
	case *j5schema.OneofField:
		return FieldTypeOneof
	case *j5schema.EnumField:
		return FieldTypeEnum
	case *j5schema.ArrayField:
		return FieldTypeArray
	case *j5schema.MapField:
		return FieldTypeMap
	case *j5schema.ScalarSchema:
		return FieldTypeScalar
	default:
		return FieldTypeUnknown
	}
}

func (p *property) Field() (Value, error) {
	switch p.Type() {
	case FieldTypeObject:
		return p.ObjectField()
	case FieldTypeOneof:
		return p.OneofField()
	case FieldTypeEnum:
		return p.EnumField()
	case FieldTypeArray:
		return p.ArrayField()
	case FieldTypeMap:
		return p.MapField()
	case FieldTypeScalar:
		return p.ScalarField()
	default:
		return nil, fmt.Errorf("field %s is not a message but has nested types", p.JSONName())
	}
}

type fieldset struct {
	asMap   map[string]*property
	asSlice []*property
}

func newFieldset(props []*property) (*fieldset, error) {
	fs := &fieldset{
		asMap: map[string]*property{},
	}

	for _, prop := range props {
		fs.asMap[prop.schema.JSONName] = prop
		fs.asSlice = append(fs.asSlice, prop)
	}

	return fs, nil
}

func (fs *fieldset) RangeProperties(callback RangeCallback) error {
	var err error
	for _, prop := range fs.asSlice {
		err = callback(prop)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *fieldset) RangeSetProperties(callback RangeCallback) error {
	var err error
	for _, prop := range fs.asSlice {
		if prop.IsSet() {
			err = callback(prop)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fs *fieldset) GetOne() (Property, error) {
	var property Property

	for _, prop := range fs.asSlice {
		if prop.value.IsSet() {
			if property != nil {
				return nil, fmt.Errorf("multiple values set for oneof")
			}
			property = prop
		}
	}

	return property, nil
}

func collectProperties(properties []*j5schema.ObjectProperty, msg protoreflect.Message) ([]*property, error) {
	out := make([]*property, 0)

properties:
	for _, schema := range properties {
		if len(schema.ProtoField) == 0 {
			var childProperties []*j5schema.ObjectProperty

			switch wrapper := schema.Schema.(type) {
			case *j5schema.ObjectField:
				childProperties = wrapper.Schema().Properties
			case *j5schema.OneofField:
				childProperties = wrapper.Schema().Properties
			default:
				return nil, fmt.Errorf("unsupported schema type %T for nested json", wrapper)
			}
			children, err := collectProperties(childProperties, msg)
			if err != nil {
				return nil, err
			}
			out = append(out, children...)
			continue
		}

		var walkFieldNumber protoreflect.FieldNumber
		var walkField protoreflect.FieldDescriptor
		walkPath := schema.ProtoField[:]
		walkMessage := msg
		for len(walkPath) > 1 {
			walkFieldNumber, walkPath = walkPath[0], walkPath[1:]
			walkField = walkMessage.Descriptor().Fields().ByNumber(walkFieldNumber)

			if !walkMessage.Has(walkField) {
				continue properties
			}
			if walkField.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("field %s is not a message but has nested types", walkField.FullName())
			}
			walkMessage = walkMessage.Get(walkField).Message()
		}
		walkFieldNumber = walkPath[0]
		walkField = walkMessage.Descriptor().Fields().ByNumber(walkFieldNumber)

		out = append(out, &property{
			schema: schema,
			value:  newFieldWrapper(schemaType(schema.Schema), walkMessage, walkField),
		})
	}

	return out, nil
}
