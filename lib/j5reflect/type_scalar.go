package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/lib/j5schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*** Interface ***/

type ScalarField interface {
	Field
	ToGoValue() (any, error)
	SetGoValue(value any) error
	SetASTValue(ASTValue) error
}

type ArrayOfScalarField interface {
	ArrayField
	AppendGoValue(value any) (int, error)
	AppendASTValue(ASTValue) (int, error)
}

type MapOfScalarField interface {
	SetGoValue(key string, value any) error
	SetASTValue(key string, value ASTValue) error
}

/*** Implementation ***/

type scalarField struct {
	fieldDefaults
	fieldContext
	value  protoContext
	schema *j5schema.ScalarSchema
}

type scalarFieldFactory struct {
	schema *j5schema.ScalarSchema
}

func (f *scalarFieldFactory) buildField(field fieldContext, value protoContext) Field {
	return &scalarField{
		fieldContext: field,
		value:        value,
		schema:       f.schema,
	}
}

func (sf *scalarField) IsSet() bool {
	return sf.value.isSet()
}

func (sf *scalarField) AsScalar() (ScalarField, bool) {
	return sf, true
}

func (sf *scalarField) SetASTValue(value ASTValue) error {
	reflectValue, err := scalarReflectFromAST(sf.schema.Proto, value)
	if err != nil {
		return err
	}

	return sf.setValue(reflectValue)
}

func (sf *scalarField) setValue(reflectValue protoreflect.Value) error {
	return sf.value.setValue(reflectValue)
}

func (sf *scalarField) SetGoValue(value any) error {
	reflectValue, err := scalarReflectFromGo(sf.schema.Proto, value)
	if err != nil {
		return fmt.Errorf("setting field %s: %w", sf.FullTypeName(), err)
	}
	return sf.setValue(reflectValue)
}

func (sf *scalarField) ToGoValue() (any, error) {
	val, ok := sf.value.getValue()
	if !ok {
		return nil, nil
	}
	return scalarGoFromReflect(sf.schema.Proto, val)
}

type arrayOfScalarField struct {
	leafArrayField
	itemSchema *j5schema.ScalarSchema
}

var _ ArrayOfScalarField = (*arrayOfScalarField)(nil)

func (array *arrayOfScalarField) AsArray() (ArrayField, bool) {
	return array, true
}

func (array *arrayOfScalarField) AsArrayOfScalar() (ArrayOfScalarField, bool) {
	return array, true
}

func (array *arrayOfScalarField) AppendGoValue(value any) (int, error) {
	reflectValue, err := scalarReflectFromGo(array.itemSchema.Proto, value)
	if err != nil {
		return -1, err
	}
	return array.appendProtoValue(reflectValue), nil
}

func (array *arrayOfScalarField) AppendASTValue(value ASTValue) (int, error) {
	reflectValue, err := scalarReflectFromAST(array.itemSchema.Proto, value)
	if err != nil {
		return -1, err
	}
	return array.appendProtoValue(reflectValue), nil
}

type mapOfScalarField struct {
	leafMapField
	itemSchema *j5schema.ScalarSchema
}

var _ MapOfScalarField = (*mapOfScalarField)(nil)

func (mapField *mapOfScalarField) AsMap() (MapField, bool) {
	return mapField, true
}

func (mapField *mapOfScalarField) SetGoValue(key string, value any) error {
	reflVal, err := scalarReflectFromGo(mapField.itemSchema.Proto, value)
	if err != nil {
		return fmt.Errorf("converting value to proto: %w", err)
	}
	mapField.setKey(key, reflVal)
	return nil
}

func (mapField *mapOfScalarField) SetASTValue(key string, value ASTValue) error {
	reflVal, err := scalarReflectFromAST(mapField.itemSchema.Proto, value)

	if err != nil {
		return fmt.Errorf("converting value to proto: %w", err)
	}
	mapField.setKey(key, reflVal)
	return nil
}
