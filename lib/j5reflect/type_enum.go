package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type EnumField interface {
	Field
	GetValue() (EnumOption, error)
	SetFromString(string) error
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

func (factory *enumFieldFactory) buildField(context fieldContext, value protoContext) Field {
	return &enumField{
		value: value,
		fieldDefaults: fieldDefaults{
			fieldType: FieldTypeEnum,
			context:   context,
		},
		schema: factory.schema,
	}
}

type enumField struct {
	fieldDefaults
	value  protoContext
	schema *j5schema.EnumField
}

var _ EnumField = (*enumField)(nil)

func (ef *enumField) IsSet() bool {
	return ef.value.isSet()
}

func (ef *enumField) AsScalar() (ScalarField, bool) {
	return ef, true
}

func (ef *enumField) Type() FieldType {
	return FieldTypeEnum
}

func (ef *enumField) GetValue() (EnumOption, error) {
	val, ok := ef.value.getValue()
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
		ef.value.setValue(protoreflect.ValueOfEnum(protoreflect.EnumNumber(option.Number())))
		return nil
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

func (ef *enumField) SetGoValue(value interface{}) error {
	switch v := value.(type) {
	case string:
		return ef.SetFromString(v)
	case *string:
		if v == nil {
			ef.value.setValue(protoreflect.ValueOfEnum(0))
			return nil
		}
		return ef.SetFromString(*v)
	default:
		return fmt.Errorf("cannot set enum value from %T", value)
	}
}

func (ef *enumField) ToGoValue() (interface{}, error) {
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

func (ef *arrayOfEnumField) AppendGoValue(value interface{}) (int, error) {
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

func (field *mapOfEnumField) SetEnum(key string, value string) error {
	option := field.itemSchema.OptionByName(value)
	if option == nil {
		return fmt.Errorf("enum value %s not found", value)
	}

	field.setKey(key, protoreflect.ValueOfEnum(protoreflect.EnumNumber(option.Number())))
	return nil
}
