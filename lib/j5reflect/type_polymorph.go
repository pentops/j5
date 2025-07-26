package j5reflect

import (
	"github.com/pentops/j5/lib/j5reflect/protoval"
	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type PolymorphField interface {
	Field
	Unwrap() (AnyField, error)
}

/*** Implementation ***/

type polymorphField struct {
	fieldDefaults
	fieldContext
	schema *j5schema.PolymorphField

	valueField protoreflect.FieldDescriptor
	value      protoval.MessageValue

	wrapped *anyField
}

var _ PolymorphField = (*polymorphField)(nil)

func (field *polymorphField) IsSet() bool {
	return field.wrapped.IsSet()
}

func (field *polymorphField) Unwrap() (AnyField, error) {
	if field.wrapped == nil {
		emptyFieldFactory := &anyFieldFactory{
			schema: &j5schema.AnyField{},
		}

		impl := emptyFieldFactory.buildField(field.fieldContext, field.value)
		field.wrapped = impl.(*anyField)
	}
	return field.wrapped, nil
}

func (field *polymorphField) SetDefaultValue() error {
	// Default value for polymorph is not defined, so we do nothing here.
	return nil
}

func (field *polymorphField) AsPolymorph() (PolymorphField, bool) {
	return field, true
}

type polymorphFieldFactory struct {
	schema *j5schema.PolymorphField
}

func (factory *polymorphFieldFactory) buildField(context fieldContext, value protoval.Value) Field {

	msgValue, ok := value.(protoval.MessageValue)
	if !ok {
		panic("polymorph field factory expected a MessageValue")
	}

	desc := msgValue.MessageDescriptor()
	valueField := desc.Fields().ByName("value")

	return &polymorphField{
		schema:       factory.schema,
		valueField:   valueField,
		fieldContext: context,
		value:        msgValue,
	}
}
