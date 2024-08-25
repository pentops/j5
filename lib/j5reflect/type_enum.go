package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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

type enumFieldFactory struct {
	schema *j5schema.EnumField
}

func (factory *enumFieldFactory) buildField(context fieldContext, value protoValueContext) Field {
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
	value  protoValueContext
	schema *j5schema.EnumField
}

var _ EnumField = (*enumField)(nil)

func (ef *enumField) IsSet() bool {
	return ef.value.isSet()
}

func (ef *enumField) SetDefault() error {
	ef.value.getOrCreateMutable()
	return nil
}

func (ef *enumField) AsScalar() (ScalarField, bool) {
	return ef, true
}

func (ef *enumField) Type() FieldType {
	return FieldTypeEnum
}

func (ef *enumField) GetValue() (EnumOption, error) {
	value := int32(ef.value.getValue().Enum())
	opt := ef.schema.Schema().OptionByNumber(value)
	if opt != nil {
		return opt, nil
	}
	return nil, fmt.Errorf("enum value %d not found", value)
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
	val := ef.value.getValue().Enum()
	opt := ef.schema.Schema().OptionByNumber(int32(val))
	if opt == nil {
		return nil, fmt.Errorf("enum value %d not found", val)
	}
	return opt.Name(), nil
}

type arrayOfEnumField struct {
	leafArrayField
	itemSchema *j5schema.EnumSchema
}

var _ ArrayOfEnumField = (*arrayOfEnumField)(nil)

func (field *arrayOfEnumField) AppendEnumFromString(name string) (int, error) {
	option := field.itemSchema.OptionByName(name)
	if option != nil {
		list := field.fieldInParent.getOrCreateMutable().List()
		list.Append(protoreflect.ValueOfEnum(protoreflect.EnumNumber(option.Number())))
		idx := list.Len() - 1
		return idx, nil
	}
	return -1, fmt.Errorf("enum value %s not found", name)
}

func (ef *arrayOfEnumField) AppendASTValue(value ASTValue) (int, error) {
	str, err := value.AsString()
	if err != nil {
		return -1, err
	}
	return ef.AppendEnumFromString(str)
}

func (ef *arrayOfEnumField) AppendGoScalar(value interface{}) (int, error) {
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
