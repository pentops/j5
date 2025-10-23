package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/lib/j5reflect/protoval"
	"github.com/pentops/j5/lib/j5schema"
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

	valueField protoval.Value
	value      protoval.MessageValue

	wrapped *anyField
}

var _ PolymorphField = (*polymorphField)(nil)

func (field *polymorphField) IsSet() bool {
	return field.wrapped.IsSet()
}

func (field *polymorphField) Unwrap() (AnyField, error) {
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

func (factory *polymorphFieldFactory) buildField(context fieldContext, valueGen protoval.Value) Field {
	msgValue, ok := valueGen.(protoval.MessageValue)
	if !ok {
		panic("polymorph field factory expected a MessageValue")
	}

	desc := msgValue.MessageDescriptor()
	valueField := desc.Fields().ByName("value")
	valueFieldWrapped, err := msgValue.ChildField(valueField)
	if err != nil {
		panic(fmt.Sprintf("polymorph field factory expected a value field in the message: %s", err))
	}

	emptyFieldFactory := &anyFieldFactory{
		schema: &j5schema.AnyField{},
	}

	impl := emptyFieldFactory.buildField(context, valueFieldWrapped)
	wrapped := impl.(*anyField)

	return &polymorphField{
		schema:       factory.schema,
		valueField:   valueFieldWrapped,
		fieldContext: context,
		value:        msgValue,
		wrapped:      wrapped,
	}
}
