package j5reflect

import (
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

	valuePair *protoPair

	wrapped *anyField
}

var _ PolymorphField = (*polymorphField)(nil)

func (field *polymorphField) IsSet() bool {
	return field.wrapped.IsSet()
}

func (field *polymorphField) Unwrap() (AnyField, error) {
	if field.wrapped == nil {
		mv, err := field.valuePair.getMutableValue(true)
		if err != nil {
			return nil, err
		}

		emptyFieldFactory := &anyFieldFactory{
			schema: &j5schema.AnyField{},
		}

		impl := emptyFieldFactory.buildField(field.fieldContext, mv.Message())
		field.wrapped = impl.(*anyField)
	}
	return field.wrapped, nil
}

func (field *polymorphField) AsPolymorph() (PolymorphField, bool) {
	return field, true
}

type polymorphFieldFactory struct {
	schema *j5schema.PolymorphField
}

func (factory *polymorphFieldFactory) buildField(context fieldContext, value protoreflect.Message) Field {

	desc := value.Descriptor()
	valueField := desc.Fields().ByName("value")
	pair := newProtoPair(value, valueField)

	return &polymorphField{
		schema:       factory.schema,
		valuePair:    pair,
		fieldContext: context,
	}
}
