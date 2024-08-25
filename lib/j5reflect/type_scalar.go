package j5reflect

import "github.com/pentops/j5/internal/j5schema"

type ScalarField interface {
	Field
	//Schema() *j5schema.ScalarSchema
	ToGoValue() (interface{}, error)
	SetGoValue(value interface{}) error
	SetASTValue(ASTValue) error
}

type scalarField struct {
	fieldDefaults
	field  protoValueContext
	schema *j5schema.ScalarSchema
}

type scalarFieldFactory struct {
	schema *j5schema.ScalarSchema
}

func (f *scalarFieldFactory) buildField(field fieldContext, value protoValueContext) Field {
	return &scalarField{
		fieldDefaults: fieldDefaults{
			fieldType: FieldTypeScalar,
			context:   field,
		},
		field:  value,
		schema: f.schema,
	}
}

func (sf *scalarField) IsSet() bool {
	return sf.field.isSet()
}

func (sf *scalarField) SetDefault() error {
	sf.field.getOrCreateMutable()
	return nil
}

func (sf *scalarField) AsScalar() (ScalarField, bool) {
	return sf, true
}

func (sf *scalarField) Type() FieldType {
	return FieldTypeScalar
}

func (sf *scalarField) SetASTValue(value ASTValue) error {
	reflectValue, err := scalarReflectFromAST(sf.schema.Proto, value)
	if err != nil {
		return err
	}

	sf.field.setValue(reflectValue)
	return nil
}

func (sf *scalarField) SetGoValue(value interface{}) error {
	reflectValue, err := scalarReflectFromGo(sf.schema.Proto, value)
	if err != nil {
		return err
	}

	sf.field.setValue(reflectValue)
	return nil
}

func (sf *scalarField) ToGoValue() (interface{}, error) {
	return scalarGoFromReflect(sf.schema.Proto, sf.field.getValue())
}

type arrayOfScalarField struct {
	leafArrayField
	itemSchema *j5schema.ScalarSchema
}

var _ ArrayOfScalarField = (*arrayOfScalarField)(nil)

func (field *arrayOfScalarField) AppendGoScalar(val interface{}) (int, error) {
	list := field.fieldInParent.getOrCreateMutable().List()
	value, err := scalarReflectFromGo(field.itemSchema.Proto, val)
	if err != nil {
		return -1, err
	}
	list.Append(value)
	idx := list.Len() - 1
	return idx, nil
}

func (field *arrayOfScalarField) AppendASTValue(value ASTValue) (int, error) {
	reflectValue, err := scalarReflectFromAST(field.itemSchema.Proto, value)
	if err != nil {
		return -1, err
	}
	list := field.fieldInParent.getOrCreateMutable().List()
	list.Append(reflectValue)
	idx := list.Len() - 1
	return idx, nil
}
