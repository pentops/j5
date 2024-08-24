package j5reflect

import (
	"github.com/pentops/j5/gen/j5/schema/v1/schema_j5pb"
	"github.com/pentops/j5/internal/j5schema"
)

type fieldBase struct {
	schema *j5schema.ObjectProperty
}

func (f fieldBase) AsScalarField() ScalarField {
	return nil
}

func (f fieldBase) Schema() *schema_j5pb.ObjectProperty {
	return f.schema.ToJ5Proto()
}

func (f fieldBase) JSONName() string {
	return f.schema.JSONName
}

func (f fieldBase) FullName() string {
	if f.schema.Parent == nil {
		return "?" + f.schema.JSONName
	}
	return f.schema.Parent.FullName() + "." + f.schema.JSONName
}

type objectProperty struct {
	field ObjectField
	fieldBase
}

func (op objectProperty) Field() Field {
	return op.field
}

func (op objectProperty) IsSet() bool {
	return op.field.IsSet()
}

var _ Property = objectProperty{}

type oneofProperty struct {
	field *oneofField
	fieldBase
}

func (op oneofProperty) Field() Field {
	return op.field
}

func (op oneofProperty) IsSet() bool {
	return op.field.IsSet()
}

var _ Property = oneofProperty{}

type enumProperty struct {
	field *enumField
	fieldBase
}

func (ep enumProperty) Field() Field {
	return ep.field
}

func (ep enumProperty) IsSet() bool {
	return ep.field.IsSet()
}

var _ Property = enumProperty{}

type arrayProperty struct {
	field ArrayField
	fieldBase
}

func (ap arrayProperty) IsSet() bool {
	return ap.field.IsSet()
}

func (ap arrayProperty) Field() Field {
	return ap.field
}

var _ Property = arrayProperty{}

type mapProperty struct {
	field MapField
	fieldBase
}

func (mp mapProperty) IsSet() bool {
	return mp.field.IsSet()
}

func (mp mapProperty) Field() Field {
	return mp.field
}

var _ Property = mapProperty{}

type scalarProperty struct {
	field *scalarField
	fieldBase
}

func (sp scalarProperty) AsScalarField() ScalarField {
	return sp.field
}

func (sp scalarProperty) Field() Field {
	return sp.field
}

func (sp scalarProperty) IsSet() bool {
	return sp.field.IsSet()
}

var _ Property = scalarProperty{}
