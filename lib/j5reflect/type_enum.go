package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/lib/j5reflect/protoval"
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type EnumField interface {
	Field
	GetValue() (EnumOption, error)
	SetFromString(string) error
	SetDefaultValue() error
}

type ArrayOfEnumField interface {
	ArrayOfScalarField
	AppendEnumFromString(string) (int, error)
}

type MapOfEnumField interface {
	SetEnum(key string, value string) error
}

type EnumOption interface {
	Name() string
	Number() int32
	Description() string
}

/*** Implementation ***/

type enumFieldFactory struct {
	schema *j5schema.EnumField
}

func (factory *enumFieldFactory) buildField(context fieldContext, value protoval.Value) Field {
	return &enumField{
		value:        value,
		fieldContext: context,
		schema:       factory.schema,
	}
}

type enumField struct {
	fieldDefaults
	fieldContext

	value  protoval.Value
	schema *j5schema.EnumField
}

var _ EnumField = (*enumField)(nil)

func (ef *enumField) IsSet() bool {
	return ef.value.IsSet()
}

func (ef *enumField) AsScalar() (ScalarField, bool) {
	return ef, true
}

func (ef *enumField) SetDefaultValue() error {
	return ef.value.SetValue(protoreflect.ValueOfEnum(protoreflect.EnumNumber(0)))
}

func (ef *enumField) AsEnum() (EnumField, bool) {
	return ef, true
}

func (ef *enumField) GetValue() (EnumOption, error) {
	val, ok := ef.value.GetValue()
	if !ok {
		return nil, fmt.Errorf("enum value not set")
	}
	numVal := int32(val.Enum())
	opt := ef.schema.Schema().OptionByNumber(numVal)
	if opt != nil {
		return opt, nil
	}
	return nil, fmt.Errorf("enum value %d not found", numVal)
}

func (ef *enumField) SetFromString(val string) error {
	option := ef.schema.Schema().OptionByName(val)
	if option != nil {
		return ef.value.SetValue(protoreflect.ValueOfEnum(protoreflect.EnumNumber(option.Number())))
	}
	return fmt.Errorf("enum value %s not found", val)
}

func (ef *enumField) SetASTValue(value ASTValue) error {
	str, err := value.AsString()
	if err != nil {
		return err
	}
	return ef.SetFromString(str)
}

func (ef *enumField) SetGoValue(value any) error {
	switch v := value.(type) {
	case string:
		return ef.SetFromString(v)
	case *string:
		if v == nil {
			return ef.value.SetValue(protoreflect.ValueOfEnum(0))
		}
		return ef.SetFromString(*v)
	default:
		return fmt.Errorf("cannot set enum value from %T", value)
	}
}

func (ef *enumField) ToGoValue() (any, error) {
	val, err := ef.GetValue()
	if err != nil {
		return nil, err
	}
	return val.Name(), nil
}

type arrayOfEnumField struct {
	leafArrayField
	itemSchema *j5schema.EnumSchema
}

var _ ArrayOfEnumField = (*arrayOfEnumField)(nil)
var _ ArrayOfScalarField = (*arrayOfEnumField)(nil)

func (field *arrayOfEnumField) AsArray() (ArrayField, bool) {
	return field, true
}

func (field *arrayOfEnumField) AsArrayOfScalar() (ArrayOfScalarField, bool) {
	return field, true
}

func (field *arrayOfEnumField) AppendEnumFromString(name string) (int, error) {
	option := field.itemSchema.OptionByName(name)
	if option == nil {
		return -1, fmt.Errorf("enum value %s not found", name)
	}

	val := protoreflect.ValueOfEnum(protoreflect.EnumNumber(option.Number()))
	return field.appendProtoValue(val), nil
}

func (ef *arrayOfEnumField) AppendASTValue(value ASTValue) (int, error) {
	str, err := value.AsString()
	if err != nil {
		return -1, err
	}
	return ef.AppendEnumFromString(str)
}

func (ef *arrayOfEnumField) AppendGoValue(value any) (int, error) {
	switch v := value.(type) {
	case string:
		return ef.AppendEnumFromString(v)
	case *string:
		if v == nil {
			return -1, fmt.Errorf("cannot append nil value")
		}

		return ef.AppendEnumFromString(*v)
	default:
		return -1, fmt.Errorf("cannot set enum value from %T", value)
	}
}

type mapOfEnumField struct {
	leafMapField
	itemSchema *j5schema.EnumSchema
}

func (field *mapOfEnumField) AsMap() (MapField, bool) {
	return field, true
}

func (field *mapOfEnumField) SetEnum(key string, value string) error {
	option := field.itemSchema.OptionByName(value)
	if option == nil {
		return fmt.Errorf("enum value %s not found", value)
	}

	return field.value.ValueAt(key).SetValue(protoreflect.ValueOfEnum(protoreflect.EnumNumber(option.Number())))
}
