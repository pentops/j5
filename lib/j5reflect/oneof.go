package j5reflect

import (
	"fmt"

	"github.com/pentops/j5/internal/j5schema"
)

type Oneof interface {
	PropertySet
	GetOne() (Property, error)
}

type OneofImpl struct {
	schema *j5schema.OneofSchema
	value  *protoMessageWrapper
	*propSet
}

func newOneof(schema *j5schema.OneofSchema, value *protoMessageWrapper) (*OneofImpl, error) {

	props, err := collectProperties(schema.Properties, value)
	if err != nil {
		return nil, err
	}

	fieldset, err := newPropSet(schema.FullName(), props)
	if err != nil {
		return nil, err
	}
	return &OneofImpl{
		schema:  schema,
		value:   value,
		propSet: fieldset,
	}, nil
}

func (fs *OneofImpl) GetOne() (Property, error) {
	var property Property
	for _, prop := range fs.asSlice {
		if prop.IsSet() {
			if property != nil {
				return nil, fmt.Errorf("multiple values set for oneof")
			}
			property = prop
		}
	}
	return property, nil
}
