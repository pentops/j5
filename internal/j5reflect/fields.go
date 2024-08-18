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

type Field interface {
	Type() FieldType
	IsSet() bool
	asProperty(fieldBase) Property
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

type fieldFactory interface {
	buildField(value protoValueContext) Field
}

type objectFieldFactory struct {
	schema *j5schema.ObjectField
}

func (f *objectFieldFactory) buildField(value protoValueContext) Field {
	return newObjectField(f.schema, value)
}

type oneofFieldFactory struct {
	schema *j5schema.OneofField
}

func (f *oneofFieldFactory) buildField(value protoValueContext) Field {
	return newOneofField(f.schema, value)
}

type enumFieldFactory struct {
	schema *j5schema.EnumField
}

func (f *enumFieldFactory) buildField(value protoValueContext) Field {
	return newEnumField(f.schema, value)
}

type scalarFieldFactory struct {
	schema *j5schema.ScalarSchema
}

func (f *scalarFieldFactory) buildField(value protoValueContext) Field {
	return newScalarField(f.schema, value)
}

func newFieldFactory(schema j5schema.FieldSchema, field protoreflect.FieldDescriptor) (fieldFactory, error) {
	switch st := schema.(type) {
	case *j5schema.ObjectField:
		if field.Kind() != protoreflect.MessageKind {
			return nil, fmt.Errorf("ObjectField is kind %s", field.Kind())
		}
		return &objectFieldFactory{schema: st}, nil

	case *j5schema.OneofField:
		if field.Kind() != protoreflect.MessageKind {
			return nil, fmt.Errorf("OneofField is kind %s", field.Kind())
		}
		return &oneofFieldFactory{schema: st}, nil

	case *j5schema.EnumField:
		if field.Kind() != protoreflect.EnumKind {
			return nil, fmt.Errorf("EnumField is kind %s", field.Kind())
		}
		return &enumFieldFactory{schema: st}, nil

	case *j5schema.ScalarSchema:
		if st.WellKnownTypeName != "" {
			if field.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("ScalarField is kind %s, want message for %T", field.Kind(), st.Proto.Type)
			}
			if string(field.Message().FullName()) != string(st.WellKnownTypeName) {
				return nil, fmt.Errorf("ScalarField message is %s, want %s for %T", field.Message().FullName(), st.WellKnownTypeName, st.Proto.Type)
			}
		} else if field.Kind() != st.Kind {
			return nil, fmt.Errorf("ScalarField value is kind %s, want %s for %T", field.Kind(), st.Kind, st.Proto.Type)
		}
		return &scalarFieldFactory{schema: st}, nil

	default:
		return nil, fmt.Errorf("unsupported schema type %T", schema)
	}
}

type ObjectField interface {
	Field
	Object() (*Object, error)
}

type objectField struct {
	protoValueContext
	schema  *j5schema.ObjectField
	_object *Object
}

var _ ObjectField = (*objectField)(nil)

func newObjectField(schema *j5schema.ObjectField, value protoValueContext) *objectField {
	of := &objectField{
		protoValueContext: value,
		schema:            schema,
	}
	return of
}

func (obj *objectField) asProperty(base fieldBase) Property {
	return &objectProperty{
		objectField: obj,
		fieldBase:   base,
	}
}

func (obj *objectField) Type() FieldType {
	return FieldTypeObject
}

func (obj *objectField) Object() (*Object, error) {
	if obj._object == nil {
		msg := obj.protoValueContext.getOrCreateMutable().Message()
		msgChild, err := newChildMessageValue(obj.protoValueContext, msg)
		if err != nil {
			return nil, err
		}
		built, err := newObject(obj.schema.Schema(), msgChild)
		if err != nil {
			return nil, err
		}
		obj._object = built
	}
	return obj._object, nil
}

type OneofField interface {
	Field
	Oneof() (*Oneof, error)
}

type oneofField struct {
	protoValueContext
	schema *j5schema.OneofField
	_oneof *Oneof
}

var _ OneofField = (*oneofField)(nil)

func newOneofField(schema *j5schema.OneofField, value protoValueContext) *oneofField {
	return &oneofField{
		protoValueContext: value,
		schema:            schema,
	}
}

func (field *oneofField) asProperty(base fieldBase) Property {
	return &oneofProperty{
		oneofField: field,
		fieldBase:  base,
	}
}

func (field *oneofField) Type() FieldType {
	return FieldTypeOneof
}

func (field *oneofField) Oneof() (*Oneof, error) {
	if field._oneof == nil {
		if !field.IsSet() {
			return nil, fmt.Errorf("object field is not set")
		}
		msg := field.protoValueContext.getOrCreateMutable().Message()
		msgChild, err := newChildMessageValue(field.protoValueContext, msg)
		if err != nil {
			return nil, err
		}
		obj, err := newOneof(field.schema.Schema(), msgChild)
		if err != nil {
			return nil, err
		}
		field._oneof = obj
	}

	return field._oneof, nil
}

type EnumField interface {
	Field
	GetValue() (*j5schema.EnumOption, error)
}

type enumField struct {
	protoValueContext
	schema *j5schema.EnumField
}

var _ EnumField = (*enumField)(nil)

func newEnumField(schema *j5schema.EnumField, value protoValueContext) *enumField {
	return &enumField{
		protoValueContext: value,
		schema:            schema,
	}
}

func (ef *enumField) asProperty(base fieldBase) Property {
	return &enumProperty{
		enumField: ef,
		fieldBase: base,
	}
}

func (ef *enumField) Type() FieldType {
	return FieldTypeEnum
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

/*
	type ArrayItem interface {
		SetGoValue(interface{}) error
	}

	type arrayItem struct {
		field  *arrayField
		schema j5schema.FieldSchema
		idx    int
	}
*/
type ArrayField interface {
	Field
	Range(func(Field) error) error
}

type baseArrayField struct {
	*realProtoMessageField
	schema  *j5schema.ArrayField
	factory fieldFactory
}

func (field *baseArrayField) Type() FieldType {
	return FieldTypeArray
}

func (field *baseArrayField) Range(cb func(Field) error) error {
	if !field.IsSet() {
		return nil
	}
	list := field.getValue().List()

	for i := 0; i < list.Len(); i++ {
		val := list.Get(i)
		wrapped := &protoListItem{
			protoValueWrapper: protoValueWrapper{
				value:   val,
				prField: field.realProtoMessageField.field,
			},
			prList: list,
			idx:    i,
		}
		property := field.factory.buildField(wrapped)

		err := cb(property)
		if err != nil {
			return err
		}
	}
	return nil
}

func newArrayField(schema *j5schema.ArrayField, value *realProtoMessageField) (ArrayField, error) {
	if !value.field.IsList() {
		return nil, fmt.Errorf("ArrayField is not a list")
	}

	factory, err := newFieldFactory(schema.Schema, value.field)
	if err != nil {
		return nil, err
	}

	base := baseArrayField{
		realProtoMessageField: value,
		schema:                schema,
		factory:               factory,
	}

	if schema.Schema.Mutable() {
		return &mutableArrayField{
			baseArrayField: base,
		}, nil
	} else {
		return &leafArrayField{
			baseArrayField: base,
		}, nil
	}
}

type MutableArrayField interface {
	ArrayField
	NewElement() Field
}

type mutableArrayField struct {
	baseArrayField
}

var _ MutableArrayField = (*mutableArrayField)(nil)

func (field *mutableArrayField) asProperty(base fieldBase) Property {
	return &mutableArrayProperty{
		mutableArrayField: field,
		fieldBase:         base,
	}
}

func (field *mutableArrayField) NewElement() Field {
	list := field.realProtoMessageField.getOrCreateMutable().List()
	idx := list.Len()
	elem := list.AppendMutable()
	element := &protoListItem{
		protoValueWrapper: protoValueWrapper{
			prField: field.realProtoMessageField.field,
			value:   elem,
		},
		prList: list,
		idx:    idx,
	}
	property := field.factory.buildField(element)
	return property
}

type LeafArrayField interface {
	ArrayField
	AppendGoValue(value interface{}) error
}

type leafArrayField struct {
	baseArrayField
}

var _ LeafArrayField = (*leafArrayField)(nil)

func (field *leafArrayField) asProperty(base fieldBase) Property {
	return &leafArrayProperty{
		leafArrayField: field,
		fieldBase:      base,
	}
}

func (field *leafArrayField) AppendGoValue(value interface{}) error {
	list := field.realProtoMessageField.getOrCreateMutable().List()
	switch field.schema.Schema.(type) {
	case *j5schema.ScalarSchema:
		stringVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		list.Append(protoreflect.ValueOfString(stringVal))

	default:
		return fmt.Errorf("unsupported scalar type %T", field.schema.Schema)
	}

	return nil
}

type MapField interface {
	Field
	Range(func(string, Field) error) error
}

type baseMapField struct {
	*realProtoMessageField
	schema  *j5schema.MapField
	factory fieldFactory
}

func (field *baseMapField) Type() FieldType {
	return FieldTypeMap
}

func (field *baseMapField) Range(cb func(string, Field) error) error {
	if !field.IsSet() {
		return nil
	}
	mapVal := field.getValue().Map()
	var outerErr error

	fieldDef := field.realProtoMessageField.field.MapValue()

	mapVal.Range(func(key protoreflect.MapKey, val protoreflect.Value) bool {
		keyStr := key.Value().String()
		wrapped := &protoMapItem{
			protoValueWrapper: protoValueWrapper{
				value:   val,
				prField: fieldDef,
			},
			key:   key,
			prMap: mapVal,
		}
		itemField := field.factory.buildField(wrapped)
		outerErr = cb(keyStr, itemField)
		return outerErr == nil
	})
	return outerErr
}

func newMapField(schema *j5schema.MapField, value *realProtoMessageField) (MapField, error) {
	if !value.field.IsMap() {
		return nil, fmt.Errorf("MapField is not a map")
	}

	factory, err := newFieldFactory(schema.Schema, value.field.MapValue())
	if err != nil {
		return nil, fmt.Errorf("factory for map value: %w", err)
	}

	base := baseMapField{
		realProtoMessageField: value,
		schema:                schema,
		factory:               factory,
	}

	if schema.Schema.Mutable() {
		return &mutableMapField{
			baseMapField: base,
		}, nil
	} else {
		return &leafMapField{
			baseMapField: base,
		}, nil
	}
}

type MutableMapField interface {
	MapField
	NewValue(key string) Field
}

type mutableMapField struct {
	baseMapField
}

func (field *mutableMapField) asProperty(base fieldBase) Property {
	return &mutableMapProperty{
		mutableMapField: field,
		fieldBase:       base,
	}
}

var _ MutableMapField = (*mutableMapField)(nil)

func (field *mutableMapField) NewValue(key string) Field {
	mapVal := field.realProtoMessageField.getOrCreateMutable().Map()
	keyVal := protoreflect.ValueOfString(key).MapKey()
	val := mapVal.Get(keyVal)
	wrapped := &protoMapItem{
		protoValueWrapper: protoValueWrapper{
			value:   val,
			prField: field.realProtoMessageField.field.MapValue(),
		},
		key:   keyVal,
		prMap: mapVal,
	}
	property := field.factory.buildField(wrapped)
	return property
}

type LeafMapField interface {
	MapField
	SetGoValue(key string, value interface{}) error
}

type leafMapField struct {
	baseMapField
}

var _ LeafMapField = (*leafMapField)(nil)

func (field *leafMapField) asProperty(base fieldBase) Property {
	return &leafMapProperty{
		leafMapField: field,
		fieldBase:    base,
	}
}

func (field *leafMapField) SetGoValue(key string, value interface{}) error {
	field.setKey(protoreflect.ValueOfString(key).MapKey(), protoreflect.ValueOf(value))
	return nil
}

func (field *leafMapField) setKey(key protoreflect.MapKey, val protoreflect.Value) {
	mapVal := field.realProtoMessageField.getOrCreateMutable().Map()
	mapVal.Set(key, val)
}

type ScalarField interface {
	Field
	Schema() *j5schema.ScalarSchema
	ToGoValue() (interface{}, error)
	SetGoValue(value interface{}) error
}

type scalarField struct {
	protoValueContext
	schema *j5schema.ScalarSchema
}

func newScalarField(schema *j5schema.ScalarSchema, value protoValueContext) *scalarField {

	return &scalarField{
		protoValueContext: value,
		schema:            schema,
	}
}

func (sf *scalarField) asProperty(base fieldBase) Property {
	return &scalarProperty{
		scalarField: sf,
		fieldBase:   base,
	}
}

func (sf *scalarField) Type() FieldType {
	return FieldTypeScalar
}

func (sf *scalarField) Schema() *j5schema.ScalarSchema {
	return sf.schema
}

func (sf *scalarField) SetGoValue(value interface{}) error {
	var pv protoreflect.Value
	switch sf.schema.Proto.Type.(type) {
	case *schema_j5pb.Field_String_:
		stringVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		pv = protoreflect.ValueOfString(stringVal)
	default:
		return fmt.Errorf("unsupported scalar type %T", sf.schema.Proto.Type)
	}

	sf.setValue(pv)
	return nil
}

type AnyValue interface{}

type KeyValue string

func (sf *scalarField) ToGoValue() (interface{}, error) {
	val := sf.getValue()
	switch sf.schema.Proto.Type.(type) {
	case *schema_j5pb.Field_Any:
		return AnyValue(val.Interface()), nil

	case *schema_j5pb.Field_Boolean:
		return val.Bool(), nil

	case *schema_j5pb.Field_String_:
		return val.String(), nil

	case *schema_j5pb.Field_Key:
		return KeyValue(val.String()), nil

	case *schema_j5pb.Field_Integer:
		switch sf.schema.Proto.GetInteger().Format {
		case schema_j5pb.IntegerField_FORMAT_INT32:
			return int32(val.Int()), nil

		case schema_j5pb.IntegerField_FORMAT_INT64:
			return val.Int(), nil

		case schema_j5pb.IntegerField_FORMAT_UINT32:
			return uint32(val.Uint()), nil

		case schema_j5pb.IntegerField_FORMAT_UINT64:
			return val.Uint(), nil

		default:
			return nil, fmt.Errorf("unsupported integer format %v", sf.schema.Proto.GetInteger().Format)
		}

	case *schema_j5pb.Field_Float:
		switch sf.schema.Proto.GetFloat().Format {
		case schema_j5pb.FloatField_FORMAT_FLOAT32:
			return float32(val.Float()), nil

		case schema_j5pb.FloatField_FORMAT_FLOAT64:
			return val.Float(), nil

		default:
			return nil, fmt.Errorf("unsupported float format %v", sf.schema.Proto.GetFloat().Format)
		}

	case *schema_j5pb.Field_Bytes:
		return val.Bytes(), nil

	case *schema_j5pb.Field_Date:
		msg := val.Message()
		val := &date_j5t.Date{}
		pv := val.ProtoReflect()
		copyReflect(msg, pv)
		return val, nil

	case *schema_j5pb.Field_Decimal:
		msg := val.Message()
		val := &decimal_j5t.Decimal{}
		pv := val.ProtoReflect()
		copyReflect(msg, pv)
		return val, nil

	case *schema_j5pb.Field_Timestamp:
		msg := val.Message()
		seconds := msg.Get(msg.Descriptor().Fields().ByName("seconds")).Int()
		nanos := msg.Get(msg.Descriptor().Fields().ByName("nanos")).Int()
		t := time.Unix(seconds, nanos).In(time.UTC)
		return t, nil

	default:
		return nil, fmt.Errorf("unsupported scalar type %T", sf.schema.Proto.Type)
	}

}
